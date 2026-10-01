package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/importer"
	"github.com/yitau/cissp-quiz-trainer/internal/repository"
	"io"
	"os"
	"sync"
	"time"
)

type Trainer struct {
	mu      sync.Mutex
	repo    repository.Trainer
	preview *domain.QuestionFile
	token   string
}

func New(repo repository.Trainer) *Trainer { return &Trainer{repo: repo} }

// Busy prevents closing the desktop window while a database/file operation is active.
func (t *Trainer) Busy() bool {
	if !t.mu.TryLock() {
		return true
	}
	t.mu.Unlock()
	return false
}
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
func (t *Trainer) PreviewFile(ctx context.Context, path string) (domain.Preview, error) {
	f, err := os.Open(path)
	if err != nil {
		return domain.Preview{}, fmt.Errorf("读取题集：%w", err)
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, importer.MaxFileBytes+1))
	if err != nil {
		return domain.Preview{}, fmt.Errorf("读取题集：%w", err)
	}
	return t.Preview(ctx, b)
}
func (t *Trainer) Preview(ctx context.Context, data []byte) (domain.Preview, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.preview = nil
	t.token = ""
	f, err := importer.Parse(data)
	if err != nil {
		return domain.Preview{}, err
	}
	status, messages, err := t.repo.ImportCheck(ctx, f)
	if err != nil {
		return domain.Preview{}, fmt.Errorf("检查导入：%w", err)
	}
	t.preview = &f
	t.token = fmt.Sprintf("%x", sha256.Sum256(data))
	questions := make([]domain.Question, len(f.Questions))
	for i, q := range f.Questions {
		questions[i] = domain.HideSolution(q)
	}
	return domain.Preview{Set: f.Set, Questions: questions, Status: status, Messages: messages, Token: t.token}, nil
}
func (t *Trainer) Import(ctx context.Context, token string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.preview == nil || token != t.token {
		return fmt.Errorf("预览已失效，请重新选择题集")
	}
	status, _, err := t.repo.ImportCheck(ctx, *t.preview)
	if err != nil {
		return fmt.Errorf("检查导入：%w", err)
	}
	if status == "duplicate" {
		return nil
	}
	if status != "ready" {
		return fmt.Errorf("存在 ID 冲突，未导入任何题目")
	}
	if err := t.repo.Import(ctx, *t.preview); err != nil {
		return fmt.Errorf("导入失败，未保存任何题目：%w", err)
	}
	t.preview = nil
	t.token = ""
	return nil
}
func (t *Trainer) ListSets(ctx context.Context) ([]domain.SetSummary, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.ListSets(ctx)
}
func (t *Trainer) SetQuestions(ctx context.Context, id string) ([]domain.Question, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	qs, err := t.repo.SetQuestions(ctx, id)
	for i := range qs {
		qs[i] = domain.HideSolution(qs[i])
	}
	return qs, err
}
func (t *Trainer) Start(ctx context.Context, setID, mode string) (domain.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if mode != "study" && mode != "exam" {
		return domain.Session{}, fmt.Errorf("不支持的练习模式")
	}
	sets, err := t.repo.ListSets(ctx)
	if err != nil {
		return domain.Session{}, err
	}
	title := ""
	for _, s := range sets {
		if s.ID == setID {
			title = s.Title
		}
	}
	qs, err := t.repo.SetQuestions(ctx, setID)
	if err != nil {
		return domain.Session{}, err
	}
	if len(qs) == 0 {
		return domain.Session{}, fmt.Errorf("题集不存在或没有题目")
	}
	return t.create(ctx, title, mode, qs)
}
func (t *Trainer) create(ctx context.Context, title, mode string, qs []domain.Question) (domain.Session, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return domain.Session{}, fmt.Errorf("创建会话标识：%w", err)
	}
	s := domain.Session{ID: fmt.Sprintf("%x", id), Title: title, Mode: mode, Status: "active", StartedAt: now(), Items: []domain.SessionItem{}}
	for _, q := range qs {
		s.Items = append(s.Items, domain.SessionItem{Question: q})
	}
	if err := t.repo.CreateSession(ctx, s); err != nil {
		return domain.Session{}, fmt.Errorf("创建会话：%w", err)
	}
	return visible(s), nil
}
func visible(s domain.Session) domain.Session {
	s.Statistics = sessionStatistics(s)
	if s.CompletedAt != "" {
		start, _ := time.Parse(time.RFC3339Nano, s.StartedAt)
		end, _ := time.Parse(time.RFC3339Nano, s.CompletedAt)
		s.DurationSeconds = int64(end.Sub(start).Seconds())
	}
	for i := range s.Items {
		if s.Status != "completed" && (s.Mode == "exam" || !s.Items[i].Scored) {
			s.Items[i].Question = domain.HideSolution(s.Items[i].Question)
			s.Items[i].Correct = nil
		}
	}
	return s
}
func (t *Trainer) Session(ctx context.Context, id string) (domain.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s, err := t.repo.Session(ctx, id)
	return visible(s), err
}
func (t *Trainer) SubmitStudy(ctx context.Context, id, questionID, selected string) (domain.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s, err := t.repo.Session(ctx, id)
	if err != nil {
		return s, err
	}
	if s.Mode != "study" {
		return domain.Session{}, fmt.Errorf("考试必须交卷后统一判分")
	}
	if selected != "A" && selected != "B" && selected != "C" && selected != "D" {
		return domain.Session{}, fmt.Errorf("请先选择 A/B/C/D 中的一个答案")
	}
	for i := range s.Items {
		item := &s.Items[i]
		if item.Question.ID != questionID {
			continue
		}
		if item.Scored {
			if item.Selected == selected {
				return visible(s), nil
			}
			return domain.Session{}, fmt.Errorf("本题已提交，不能更改答案")
		}
		if s.Status != "active" {
			return domain.Session{}, fmt.Errorf("会话已结束")
		}
		item.Selected = selected
		score(item)
		if err := t.repo.SaveItem(ctx, id, *item); err != nil {
			return domain.Session{}, err
		}
		return visible(s), nil
	}
	return domain.Session{}, fmt.Errorf("题目不属于当前会话")
}
func score(i *domain.SessionItem) {
	correct := i.Selected == i.Question.Answer
	i.Correct = &correct
	i.Scored = true
	i.ScoredAt = now()
}
func (t *Trainer) Complete(ctx context.Context, id string) (domain.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s, err := t.repo.Session(ctx, id)
	if err != nil {
		return s, err
	}
	if s.Status == "completed" {
		return visible(s), nil
	}
	for i := range s.Items {
		if !s.Items[i].Scored {
			if s.Mode == "study" {
				s.Items[i].Selected = ""
			}
			score(&s.Items[i])
		}
	}
	s.Status = "completed"
	s.CompletedAt = now()
	if err := t.repo.Complete(ctx, s); err != nil {
		return domain.Session{}, fmt.Errorf("交卷失败，请重试：%w", err)
	}
	return visible(s), nil
}

// SaveChoice persists drafts without grading, including study choices not yet submitted.
func (t *Trainer) SaveChoice(ctx context.Context, id, questionID, selected string, flagged bool) (domain.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if selected != "" && selected != "A" && selected != "B" && selected != "C" && selected != "D" {
		return domain.Session{}, fmt.Errorf("答案必须为 A/B/C/D 或留空")
	}
	s, err := t.repo.Session(ctx, id)
	if err != nil {
		return domain.Session{}, err
	}
	if s.Status != "active" {
		return domain.Session{}, fmt.Errorf("会话已完成，不能修改")
	}
	for n := range s.Items {
		i := &s.Items[n]
		if i.Question.ID != questionID {
			continue
		}
		if i.Scored {
			return domain.Session{}, fmt.Errorf("本题已提交，不能修改")
		}
		i.Selected = selected
		i.Flagged = flagged
		if err := t.repo.SaveItem(ctx, id, *i); err != nil {
			return domain.Session{}, err
		}
		return visible(s), nil
	}
	return domain.Session{}, fmt.Errorf("题目不属于当前会话")
}
func (t *Trainer) ListSessions(ctx context.Context) ([]domain.SessionSummary, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.ListSessions(ctx)
}
