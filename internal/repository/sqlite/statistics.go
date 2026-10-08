package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

func (r *Store) LearningData(ctx context.Context) ([]domain.Question, []domain.Attempt, map[string]bool, error) {
	questions, err := r.allQuestions(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, err := r.DB.QueryContext(ctx, `SELECT a.question_id,json_extract(q.body,'$.domain'),json_extract(q.body,'$.type'),a.selected,a.correct,a.scored_at FROM quiz_answers a JOIN questions q ON q.id=a.question_id WHERE a.scored=1 ORDER BY a.scored_at,a.rowid`)
	if err != nil {
		return nil, nil, nil, err
	}
	attempts := []domain.Attempt{}
	for rows.Next() {
		var a domain.Attempt
		if err := rows.Scan(&a.QuestionID, &a.Domain, &a.Type, &a.Selected, &a.Correct, &a.ScoredAt); err != nil {
			rows.Close()
			return nil, nil, nil, err
		}
		attempts = append(attempts, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, nil, err
	}
	rows, err = r.DB.QueryContext(ctx, "SELECT question_id FROM question_flags WHERE favorite=1")
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	flags := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, nil, nil, err
		}
		flags[id] = true
	}
	return questions, attempts, flags, rows.Err()
}
func (r *Store) allQuestions(ctx context.Context) ([]domain.Question, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT body FROM questions ORDER BY set_id,position")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Question{}
	for rows.Next() {
		var b string
		var q domain.Question
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(b), &q); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}
func (r *Store) SetFavorite(ctx context.Context, id string, value bool) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO question_flags(question_id,favorite) VALUES(?,?) ON CONFLICT(question_id) DO UPDATE SET favorite=excluded.favorite`, id, value)
	if err != nil {
		return fmt.Errorf("保存收藏失败：%w", err)
	}
	return nil
}
