package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/importer"
)

const maxBackupBytes = 256 << 20

func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func (t *Trainer) Backup(ctx context.Context, path string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.backup(ctx, path)
}
func (t *Trainer) backup(ctx context.Context, path string) error {
	dir, err := os.MkdirTemp("", "cissp-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	snapshot := filepath.Join(dir, "cissp.db")
	if err := t.repo.Snapshot(ctx, snapshot); err != nil {
		return err
	}
	contents, err := t.repo.InspectSnapshot(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("检查备份快照：%w", err)
	}
	if err := validateContents(contents); err != nil {
		return fmt.Errorf("当前数据库未通过备份校验：%w", err)
	}
	data, err := readLimited(snapshot, maxBackupBytes)
	if err != nil {
		return err
	}
	config, err := json.Marshal(contents.Settings)
	if err != nil {
		return err
	}
	meta, err := json.Marshal(domain.BackupMetadata{FormatVersion: 1, AppVersion: domain.AppVersion, DatabaseVersion: domain.DatabaseVersion, CreatedAt: now(), DatabaseSHA256: digest(data), ConfigSHA256: digest(config)})
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("创建备份文件失败（不覆盖已有文件）：%w", err)
	}
	success := false
	defer func() {
		f.Close()
		if !success {
			os.Remove(path)
		}
	}()
	w := zip.NewWriter(f)
	for _, entry := range []struct {
		name string
		data []byte
	}{{"metadata.json", meta}, {"config.json", config}, {"cissp.db", data}} {
		out, err := w.Create(entry.name)
		if err != nil {
			return err
		}
		if _, err := out.Write(entry.data); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	success = true
	return nil
}
func readLimited(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("文件超过 %d MiB 限制", limit>>20)
	}
	return b, nil
}
func (t *Trainer) Restore(ctx context.Context, path, safetyDir string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	dir, err := os.MkdirTemp("", "cissp-restore-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	snapshot, config, version, err := unpack(path, dir)
	if err != nil {
		return "", fmt.Errorf("备份无效，原数据未修改：%w", err)
	}
	if err := t.repo.PrepareSnapshot(ctx, snapshot, version); err != nil {
		return "", fmt.Errorf("备份版本检查或升级失败，原数据未修改：%w", err)
	}
	contents, err := t.repo.InspectSnapshot(ctx, snapshot)
	if err != nil {
		return "", fmt.Errorf("备份数据库校验失败，原数据未修改：%w", err)
	}
	if err := validateContents(contents); err != nil {
		return "", fmt.Errorf("备份业务记录无效，原数据未修改：%w", err)
	}
	if !reflect.DeepEqual(config, contents.Settings) {
		return "", fmt.Errorf("备份配置与数据库不一致，原数据未修改")
	}
	if err := os.MkdirAll(safetyDir, 0700); err != nil {
		return "", fmt.Errorf("创建恢复前安全备份目录：%w", err)
	}
	safety := filepath.Join(safetyDir, "before-restore-"+time.Now().UTC().Format("20060102T150405.000000000Z")+".zip")
	if err := t.backup(ctx, safety); err != nil {
		return "", fmt.Errorf("无法保留恢复前安全备份，已取消恢复：%w", err)
	}
	if err := t.repo.RestoreSnapshot(ctx, snapshot); err != nil {
		return "", fmt.Errorf("恢复事务失败，原数据已保留，安全备份 %s：%w", safety, err)
	}
	t.preview = nil
	t.token = ""
	return safety, nil
}
func unpack(path, dir string) (string, map[string]string, int, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", nil, 0, err
	}
	defer r.Close()
	entries := map[string][]byte{}
	if len(r.File) != 3 {
		return "", nil, 0, fmt.Errorf("完整备份必须仅包含 metadata.json、config.json、cissp.db")
	}
	for _, f := range r.File {
		limit := int64(1 << 20)
		if f.Name == "cissp.db" {
			limit = maxBackupBytes
		} else if f.Name != "metadata.json" && f.Name != "config.json" {
			return "", nil, 0, fmt.Errorf("备份包含未知路径：%s", f.Name)
		}
		if _, exists := entries[f.Name]; exists {
			return "", nil, 0, fmt.Errorf("重复备份条目")
		}
		if f.UncompressedSize64 > uint64(limit) {
			return "", nil, 0, fmt.Errorf("备份条目过大")
		}
		input, err := f.Open()
		if err != nil {
			return "", nil, 0, err
		}
		data, err := io.ReadAll(io.LimitReader(input, limit+1))
		input.Close()
		if err != nil {
			return "", nil, 0, err
		}
		if int64(len(data)) > limit {
			return "", nil, 0, fmt.Errorf("备份条目超过限制")
		}
		entries[f.Name] = data
	}
	var meta domain.BackupMetadata
	if err := json.Unmarshal(entries["metadata.json"], &meta); err != nil {
		return "", nil, 0, err
	}
	compatible := (meta.DatabaseVersion == domain.DatabaseVersion && meta.AppVersion == domain.AppVersion) || (meta.DatabaseVersion == 1 && meta.AppVersion == "0.1.0")
	if meta.FormatVersion != 1 || !compatible {
		return "", nil, 0, fmt.Errorf("备份/App/数据库版本不兼容，要求 1 / %s / %d", domain.AppVersion, domain.DatabaseVersion)
	}
	if _, err := time.Parse(time.RFC3339Nano, meta.CreatedAt); err != nil {
		return "", nil, 0, fmt.Errorf("备份时间无效")
	}
	if digest(entries["cissp.db"]) != meta.DatabaseSHA256 || digest(entries["config.json"]) != meta.ConfigSHA256 {
		return "", nil, 0, fmt.Errorf("备份校验和不匹配")
	}
	var config map[string]string
	if err := json.Unmarshal(entries["config.json"], &config); err != nil {
		return "", nil, 0, err
	}
	snapshot := filepath.Join(dir, "cissp.db")
	if err := os.WriteFile(snapshot, entries["cissp.db"], 0600); err != nil {
		return "", nil, 0, err
	}
	return snapshot, config, meta.DatabaseVersion, nil
}
func validateContents(c domain.BackupContents) error {
	expected := map[string]domain.StoredQuestion{}
	sets := map[string]bool{}
	for _, doc := range c.Documents {
		f, err := importer.Parse([]byte(doc))
		if err != nil {
			return err
		}
		if sets[f.Set.ID] {
			return fmt.Errorf("重复题集")
		}
		sets[f.Set.ID] = true
		for n, q := range f.Questions {
			if _, exists := expected[q.ID]; exists {
				return fmt.Errorf("重复题目 ID")
			}
			expected[q.ID] = domain.StoredQuestion{SetID: f.Set.ID, Position: n, Question: q}
		}
	}
	if len(expected) != len(c.Questions) {
		return fmt.Errorf("题集与题目数量不匹配")
	}
	for _, q := range c.Questions {
		want, ok := expected[q.Question.ID]
		a, _ := json.Marshal(want)
		b, _ := json.Marshal(q)
		if !ok || !bytes.Equal(a, b) {
			return fmt.Errorf("题集与题目内容不一致：%s", q.Question.ID)
		}
	}
	if len(c.Settings) != 1 || c.Settings["language"] != "zh-CN" {
		return fmt.Errorf("不支持的应用配置")
	}
	for _, s := range c.Sessions {
		start, err := time.Parse(time.RFC3339Nano, s.StartedAt)
		if err != nil {
			return fmt.Errorf("会话开始时间无效")
		}
		if s.Status != "active" && s.Status != "completed" {
			return fmt.Errorf("会话状态无效")
		}
		if s.Mode != "study" && s.Mode != "exam" {
			return fmt.Errorf("会话模式无效")
		}
		var end time.Time
		if s.Status == "completed" {
			end, err = time.Parse(time.RFC3339Nano, s.CompletedAt)
			if err != nil || end.Before(start) {
				return fmt.Errorf("会话结束时间无效")
			}
		} else if s.CompletedAt != "" {
			return fmt.Errorf("未完成会话存在结束时间")
		}
		if len(s.Items) == 0 {
			return fmt.Errorf("空会话")
		}
		seen := map[string]bool{}
		for _, i := range s.Items {
			if seen[i.Question.ID] {
				return fmt.Errorf("重复会话题目")
			}
			seen[i.Question.ID] = true
			if _, ok := expected[i.Question.ID]; !ok {
				return fmt.Errorf("会话题目缺失")
			}
			if i.Selected != "" && i.Selected != "A" && i.Selected != "B" && i.Selected != "C" && i.Selected != "D" {
				return fmt.Errorf("无效选项")
			}
			if s.Status == "completed" && !i.Scored {
				return fmt.Errorf("已完成会话存在未计分记录")
			}
			if s.Status == "active" && s.Mode == "exam" && i.Scored {
				return fmt.Errorf("未交卷考试存在评分")
			}
			if i.Scored {
				stamp, err := time.Parse(time.RFC3339Nano, i.ScoredAt)
				if err != nil || stamp.Before(start) || (!end.IsZero() && stamp.After(end)) {
					return fmt.Errorf("计分时间无效")
				}
				if i.Correct == nil || *i.Correct != (i.Selected == i.Question.Answer) {
					return fmt.Errorf("评分与答案不一致")
				}
			} else if i.Correct != nil || i.ScoredAt != "" {
				return fmt.Errorf("草稿包含评分")
			}
		}
	}
	return nil
}
