package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

func (r *Store) Snapshot(ctx context.Context, path string) error {
	if _, err := r.DB.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return fmt.Errorf("创建一致数据库快照：%w", err)
	}
	return nil
}
func readOnly(path string) string {
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := url.URL{Scheme: "file", Path: path}
	return u.String() + "?mode=ro"
}
func schema(ctx context.Context, db *sql.DB) (string, error) {
	rows, err := db.QueryContext(ctx, "SELECT type,name,coalesce(sql,'') FROM sqlite_master WHERE name NOT LIKE 'sqlite_%' ORDER BY type,name")
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var b strings.Builder
	for rows.Next() {
		var kind, name, query string
		if err := rows.Scan(&kind, &name, &query); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s|%s|%s\n", kind, name, query)
	}
	return b.String(), rows.Err()
}
func (r *Store) InspectSnapshot(ctx context.Context, path string) (domain.BackupContents, error) {
	var out domain.BackupContents
	db, err := sql.Open("sqlite", readOnly(path))
	if err != nil {
		return out, err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return out, err
	}
	if version != domain.DatabaseVersion {
		return out, fmt.Errorf("备份数据库版本不兼容：%d", version)
	}
	var check string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&check); err != nil {
		return out, err
	}
	if check != "ok" {
		return out, fmt.Errorf("备份数据库完整性检查失败：%s", check)
	}
	expected, err := schema(ctx, r.DB)
	if err != nil {
		return out, err
	}
	actual, err := schema(ctx, db)
	if err != nil {
		return out, err
	}
	if actual != expected {
		return out, fmt.Errorf("备份数据库结构与版本不匹配")
	}
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return out, err
	}
	bad := rows.Next()
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	if bad {
		return out, fmt.Errorf("备份含无效外键")
	}
	rows, err = db.QueryContext(ctx, "SELECT id,title,description,document FROM question_sets ORDER BY id")
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var id, title, description, doc string
		if err := rows.Scan(&id, &title, &description, &doc); err != nil {
			rows.Close()
			return out, err
		}
		var f domain.QuestionFile
		if err := json.Unmarshal([]byte(doc), &f); err != nil {
			rows.Close()
			return out, err
		}
		if f.Set.ID != id || f.Set.Title != title || f.Set.Description != description {
			rows.Close()
			return out, fmt.Errorf("题集索引与内容不一致")
		}
		out.Documents = append(out.Documents, doc)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	rows, err = db.QueryContext(ctx, "SELECT id,set_id,position,body,fingerprint FROM questions ORDER BY set_id,position")
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var q domain.StoredQuestion
		var id, body, fp string
		if err := rows.Scan(&id, &q.SetID, &q.Position, &body, &fp); err != nil {
			rows.Close()
			return out, err
		}
		if err := json.Unmarshal([]byte(body), &q.Question); err != nil {
			rows.Close()
			return out, err
		}
		if q.Question.ID != id || fingerprint(q.Question) != fp {
			rows.Close()
			return out, fmt.Errorf("题目身份或内容指纹不一致")
		}
		out.Questions = append(out.Questions, q)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	other := New(db)
	summaries, err := other.ListSessions(ctx)
	if err != nil {
		return out, err
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM quiz_sessions").Scan(&count); err != nil {
		return out, err
	}
	if count != len(summaries) {
		return out, fmt.Errorf("备份包含空会话")
	}
	for _, summary := range summaries {
		s, err := other.Session(ctx, summary.ID)
		if err != nil {
			return out, err
		}
		out.Sessions = append(out.Sessions, s)
	}
	// A session's positions must form a contiguous sequence starting at zero.
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM (SELECT session_id FROM quiz_answers GROUP BY session_id HAVING min(position)<>0 OR max(position)<>count(*)-1)").Scan(&count); err != nil {
		return out, err
	}
	if count != 0 {
		return out, fmt.Errorf("会话题序不完整")
	}
	rows, err = db.QueryContext(ctx, "SELECT key,value FROM app_settings")
	if err != nil {
		return out, err
	}
	defer rows.Close()
	out.Settings = map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return out, err
		}
		out.Settings[k] = v
	}
	return out, rows.Err()
}
func (r *Store) RestoreSnapshot(ctx context.Context, path string) error {
	// Keep ATTACH/transaction/DETACH on one connection. Only known tables are copied.
	conn, err := r.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "ATTACH DATABASE ? AS restore_source", readOnly(path)); err != nil {
		return err
	}
	defer conn.ExecContext(context.Background(), "DETACH DATABASE restore_source")
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range []string{"quiz_answers", "question_flags", "quiz_sessions", "questions", "question_sets", "app_settings"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM main."+table); err != nil {
			return fmt.Errorf("恢复清理事务失败：%w", err)
		}
	}
	for _, table := range []string{"question_sets", "questions", "quiz_sessions", "quiz_answers", "question_flags", "app_settings"} {
		if _, err := tx.ExecContext(ctx, "INSERT INTO main."+table+" SELECT * FROM restore_source."+table); err != nil {
			return fmt.Errorf("恢复写入事务失败：%w", err)
		}
	}
	return tx.Commit()
}
