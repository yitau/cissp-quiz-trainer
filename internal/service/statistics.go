package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

func statistics(attempts []domain.Attempt) domain.Statistics {
	out := domain.Statistics{Total: domain.Metric{Label: "全部"}, Domains: []domain.Metric{}, Types: []domain.Metric{}}
	domains := map[string]*domain.Metric{}
	types := map[string]*domain.Metric{}
	for d := 1; d <= 8; d++ {
		key := strconv.Itoa(d)
		domains[key] = &domain.Metric{Label: key}
	}
	for _, key := range []string{"FIRST", "BEST", "MOST", "NEXT", "PRIMARY", "LEAST", "MOST_LIKELY", "SCENARIO", "KNOWLEDGE"} {
		types[key] = &domain.Metric{Label: key}
	}
	for _, a := range attempts {
		addMetric(&out.Total, a)
		if m := domains[strconv.Itoa(a.Domain)]; m != nil {
			addMetric(m, a)
		}
		if m := types[a.Type]; m != nil {
			addMetric(m, a)
		}
	}
	for d := 1; d <= 8; d++ {
		out.Domains = append(out.Domains, *domains[strconv.Itoa(d)])
	}
	keys := []string{}
	for k := range types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out.Types = append(out.Types, *types[k])
	}
	return out
}
func addMetric(m *domain.Metric, a domain.Attempt) {
	m.Count++
	if a.Correct {
		m.Correct++
	} else {
		m.Wrong++
	}
	if a.Selected == "" {
		m.Unanswered++
	}
	rate := float64(m.Correct) * 100 / float64(m.Count)
	m.Rate = &rate
}
func sessionStatistics(s domain.Session) domain.Statistics {
	attempts := []domain.Attempt{}
	for _, i := range s.Items {
		if i.Scored && i.Correct != nil {
			attempts = append(attempts, domain.Attempt{Domain: i.Question.Domain, Type: i.Question.Type, Selected: i.Selected, Correct: *i.Correct})
		}
	}
	return statistics(attempts)
}
func reviewData(qs []domain.Question, attempts []domain.Attempt, flags map[string]bool) []domain.ReviewQuestion {
	// Sort parsed timestamps so old RFC3339 values with different precision remain ordered.
	sort.SliceStable(attempts, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339Nano, attempts[i].ScoredAt)
		b, _ := time.Parse(time.RFC3339Nano, attempts[j].ScoredAt)
		return a.Before(b)
	})
	out := make([]domain.ReviewQuestion, len(qs))
	positions := map[string]int{}
	for i, q := range qs {
		positions[q.ID] = i
		out[i] = domain.ReviewQuestion{Question: domain.HideSolution(q), Favorite: flags[q.ID]}
	}
	for _, a := range attempts {
		n, ok := positions[a.QuestionID]
		if !ok {
			continue
		}
		p := &out[n]
		p.Attempts++
		p.LastAnsweredAt = a.ScoredAt
		if a.Correct {
			p.Correct++
			p.Streak++
		} else {
			p.Wrong++
			p.Streak = 0
		}
		p.Mastered = p.Streak >= 3
	}
	return out
}
func (t *Trainer) Statistics(ctx context.Context) (domain.Statistics, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	qs, attempts, flags, err := t.repo.LearningData(ctx)
	if err != nil {
		return domain.Statistics{}, err
	}
	out := statistics(attempts)
	out.Favorites = len(flags)
	for _, q := range reviewData(qs, attempts, flags) {
		if q.Wrong > 0 {
			out.WrongQuestions++
		}
	}
	return out, nil
}
func (t *Trainer) Review(ctx context.Context, kind string) ([]domain.ReviewQuestion, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if kind != "wrong" && kind != "favorites" && kind != "all" {
		return nil, fmt.Errorf("复习范围无效")
	}
	qs, attempts, flags, err := t.repo.LearningData(ctx)
	if err != nil {
		return nil, err
	}
	out := []domain.ReviewQuestion{}
	for _, q := range reviewData(qs, attempts, flags) {
		if kind == "all" || (kind == "wrong" && q.Wrong > 0) || (kind == "favorites" && q.Favorite) {
			out = append(out, q)
		}
	}
	return out, nil
}
func (t *Trainer) SetFavorite(ctx context.Context, id string, value bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.SetFavorite(ctx, id, value)
}
func (t *Trainer) StartReview(ctx context.Context, kind string) (domain.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if kind != "wrong" && kind != "favorites" {
		return domain.Session{}, fmt.Errorf("复习范围无效")
	}
	qs, attempts, flags, err := t.repo.LearningData(ctx)
	if err != nil {
		return domain.Session{}, err
	}
	progress := reviewData(qs, attempts, flags)
	selected := []domain.Question{}
	for i, p := range progress {
		if (kind == "wrong" && p.Wrong > 0) || (kind == "favorites" && p.Favorite) {
			selected = append(selected, qs[i])
		}
	}
	if len(selected) == 0 {
		return domain.Session{}, fmt.Errorf("此范围暂无可复习题目")
	}
	title := "错题复习"
	if kind == "favorites" {
		title = "收藏复习"
	}
	return t.create(ctx, title, "study", selected)
}
