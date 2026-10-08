package service

import (
	"context"
	"fmt"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

// Derive navigation from persisted answers. Exam drafts never expose correctness.
func itemProgress(s domain.Session, i domain.SessionItem) string {
	if s.Status == "completed" {
		if i.Selected == "" {
			return "omitted"
		}
		if i.Correct != nil && *i.Correct {
			return "correct"
		}
		return "wrong"
	}
	if s.Mode == "study" && i.Scored {
		return "submitted"
	}
	if i.Selected == "" {
		return "unanswered"
	}
	if s.Mode == "study" {
		return "draft"
	}
	return "selected"
}

func resumeIndex(s domain.Session) int {
	if s.Status == "completed" {
		return 0
	}
	for n, i := range s.Items {
		if (s.Mode == "study" && !i.Scored) || (s.Mode == "exam" && i.Selected == "") {
			return n
		}
	}
	for n, i := range s.Items {
		if i.Flagged {
			return n
		}
	}
	return 0
}

// StartSessionReview uses the immutable result, never the learner's current mastery.
func (t *Trainer) StartSessionReview(ctx context.Context, id string) (domain.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s, err := t.repo.Session(ctx, id)
	if err != nil {
		return domain.Session{}, fmt.Errorf("读取原练习：%w", err)
	}
	if s.Status != "completed" {
		return domain.Session{}, fmt.Errorf("请先交卷，再重练本次错题")
	}
	qs := []domain.Question{}
	for _, i := range s.Items {
		if i.Scored && i.Correct != nil && !*i.Correct {
			qs = append(qs, i.Question)
		}
	}
	if len(qs) == 0 {
		return domain.Session{}, fmt.Errorf("本次没有错题或漏答")
	}
	return t.create(ctx, "错题复习 · "+s.Title, "study", qs)
}
