package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/importer"
	"sort"
	"time"
)

func (t *Trainer) PreviewLessonFile(ctx context.Context, path string) (domain.LessonPreview, error) {
	b, err := readLimited(path, importer.MaxFileBytes)
	if err != nil {
		return domain.LessonPreview{}, fmt.Errorf("读取课程：%w", err)
	}
	return t.PreviewLesson(ctx, b)
}
func (t *Trainer) PreviewLesson(ctx context.Context, data []byte) (domain.LessonPreview, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lessonPreview = nil
	t.lessonToken = ""
	f, err := importer.ParseLesson(data)
	if err != nil {
		return domain.LessonPreview{}, err
	}
	status, err := t.lessonImportStatus(ctx, f)
	if err != nil {
		return domain.LessonPreview{}, err
	}
	t.lessonPreview = &f
	t.lessonToken = digest(data)
	return domain.LessonPreview{Lesson: f.Lesson, Sections: f.Sections, ConceptCount: len(f.Concepts), Status: status, Token: t.lessonToken}, nil
}
func (t *Trainer) lessonImportStatus(ctx context.Context, f domain.LessonFile) (string, error) {
	s, err := t.repo.Lesson(ctx, f.Lesson.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return "ready", nil
	}
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(f)
	if s.Document == string(b) {
		return "duplicate", nil
	}
	return "conflict", nil
}
func (t *Trainer) CancelLessonImport() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lessonPreview = nil
	t.lessonToken = ""
}
func (t *Trainer) ImportLesson(ctx context.Context, token string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.lessonPreview == nil || token != t.lessonToken {
		return fmt.Errorf("课程预览已失效，请重新选择课程")
	}
	f := *t.lessonPreview
	status, err := t.lessonImportStatus(ctx, f)
	if err != nil {
		return err
	}
	if status == "conflict" {
		return fmt.Errorf("课程 lesson.id %s 内容冲突，拒绝覆盖；原有进度保持不变", f.Lesson.ID)
	}
	if status == "ready" {
		b, _ := json.Marshal(f)
		if err := t.repo.ImportLesson(ctx, domain.StoredLesson{ID: f.Lesson.ID, Document: string(b), Fingerprint: digest(b), ImportedAt: now()}); err != nil {
			return fmt.Errorf("导入课程失败，未保存任何课程：%w", err)
		}
	}
	t.lessonPreview = nil
	t.lessonToken = ""
	return nil
}
func lessonSummary(f domain.LessonFile, progress []domain.LessonProgress) domain.LessonSummary {
	s := domain.LessonSummary{Lesson: f.Lesson, SectionCount: len(f.Sections), ConceptCount: len(f.Concepts), ResumeConceptID: f.Sections[0].ConceptIDs[0]}
	var latest time.Time
	for _, p := range progress {
		if p.Status == "understood" {
			s.Understood++
		}
		stamp, err := time.Parse(time.RFC3339Nano, p.LastOpenedAt)
		if err == nil && stamp.After(latest) {
			latest = stamp
			s.ResumeConceptID = p.ConceptID
		}
	}
	s.Completed = s.Understood == s.ConceptCount
	return s
}
func (t *Trainer) ListLessons(ctx context.Context) ([]domain.LessonSummary, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	lessons, err := t.repo.ListLessons(ctx)
	if err != nil {
		return nil, err
	}
	out := []domain.LessonSummary{}
	for _, stored := range lessons {
		f, err := importer.ParseLesson([]byte(stored.Document))
		if err != nil {
			return nil, err
		}
		p, err := t.repo.LessonProgress(ctx, stored.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, lessonSummary(f, p))
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Lesson, out[j].Lesson
		if a.Week != b.Week {
			return a.Week < b.Week
		}
		if a.Day != b.Day {
			return a.Day < b.Day
		}
		return a.ID < b.ID
	})
	return out, nil
}
func (t *Trainer) GetLesson(ctx context.Context, id string) (domain.LessonView, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lessonView(ctx, id)
}
func (t *Trainer) lessonView(ctx context.Context, id string) (domain.LessonView, error) {
	var v domain.LessonView
	stored, err := t.repo.Lesson(ctx, id)
	if err != nil {
		return v, err
	}
	v.Document, err = importer.ParseLesson([]byte(stored.Document))
	if err != nil {
		return v, err
	}
	v.Progress, err = t.repo.LessonProgress(ctx, id)
	if err != nil {
		return v, err
	}
	v.Summary = lessonSummary(v.Document, v.Progress)
	v.LinkedSetStatus = "unlinked"
	if linked := v.Document.Lesson.LinkedQuestionSetID; linked != nil {
		v.LinkedSetStatus = "missing"
		sets, err := t.repo.ListSets(ctx)
		if err != nil {
			return v, err
		}
		for _, s := range sets {
			if s.ID == *linked {
				v.LinkedSetStatus = "available"
				if s.Archived {
					v.LinkedSetStatus = "archived"
				}
				break
			}
		}
	}
	return v, nil
}
func (t *Trainer) OpenConcept(ctx context.Context, id, conceptID string) (domain.LessonView, error) {
	return t.changeConcept(ctx, id, conceptID, "in_progress", true)
}
func (t *Trainer) SetConceptStatus(ctx context.Context, id, conceptID, status string) (domain.LessonView, error) {
	return t.changeConcept(ctx, id, conceptID, status, false)
}
func (t *Trainer) changeConcept(ctx context.Context, id, conceptID, status string, opened bool) (domain.LessonView, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if status != "in_progress" && status != "understood" && status != "needs_review" {
		return domain.LessonView{}, fmt.Errorf("无效知识点状态 %q", status)
	}
	v, err := t.lessonView(ctx, id)
	if err != nil {
		return v, err
	}
	found := false
	for _, c := range v.Document.Concepts {
		if c.ID == conceptID {
			found = true
			break
		}
	}
	if !found {
		return v, fmt.Errorf("conceptId %s 不属于课程 %s", conceptID, id)
	}
	stamp := time.Now().UTC()
	// Maintain an unambiguous reading order even if the local clock moves backwards.
	for _, p := range v.Progress {
		last, e := time.Parse(time.RFC3339Nano, p.LastOpenedAt)
		if e == nil && !stamp.After(last) {
			stamp = last.Add(time.Nanosecond)
		}
	}
	p := domain.LessonProgress{UnitID: id, ConceptID: conceptID, Status: status, UpdatedAt: stamp.Format(time.RFC3339Nano)}
	if opened {
		p.LastOpenedAt = p.UpdatedAt
	}
	if err := t.repo.SaveLessonProgress(ctx, p, opened); err != nil {
		return v, fmt.Errorf("保存知识点进度：%w", err)
	}
	return t.lessonView(ctx, id)
}
