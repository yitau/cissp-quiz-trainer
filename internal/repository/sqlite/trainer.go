package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

type Store struct{ DB *sql.DB }

func New(db *sql.DB) *Store  { return &Store{DB: db} }
func canonical(v any) string { b, _ := json.Marshal(v); return string(b) }
func fingerprint(q domain.Question) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(canonical(struct {
		Text    string
		Options map[string]string
	}{q.Question, q.Options}))))
}
func (r *Store) ListSets(ctx context.Context) ([]domain.SetSummary, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT s.id,s.title,s.description,count(q.id) FROM question_sets s JOIN questions q ON q.set_id=s.id GROUP BY s.id ORDER BY s.imported_at DESC,s.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SetSummary{}
	for rows.Next() {
		var s domain.SetSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
func (r *Store) ImportCheck(ctx context.Context, f domain.QuestionFile) (string, []string, error) {
	messages := []string{}
	var doc string
	err := r.DB.QueryRowContext(ctx, "SELECT document FROM question_sets WHERE id=?", f.Set.ID).Scan(&doc)
	if err == nil {
		if doc == canonical(f) {
			return "duplicate", []string{"题集 ID 和内容完全相同，无需重复导入"}, nil
		}
		return "conflict", []string{"题集 ID 已存在且内容不同；v0.1 不支持覆盖"}, nil
	}
	if err != sql.ErrNoRows {
		return "", nil, err
	}
	status := "ready"
	seen := map[string]bool{}
	for _, q := range f.Questions {
		var exists int
		if err := r.DB.QueryRowContext(ctx, "SELECT count(*) FROM questions WHERE id=?", q.ID).Scan(&exists); err != nil {
			return "", nil, err
		}
		if exists > 0 {
			status = "conflict"
			messages = append(messages, "题目 ID 已存在："+q.ID)
		}
		fp := fingerprint(q)
		var same int
		if err := r.DB.QueryRowContext(ctx, "SELECT count(*) FROM questions WHERE fingerprint=?", fp).Scan(&same); err != nil {
			return "", nil, err
		}
		if same > 0 || seen[fp] {
			messages = append(messages, "可能重复的题干和选项："+q.ID+"（不同 ID，可确认导入）")
		}
		seen[fp] = true
	}
	return status, messages, nil
}
func (r *Store) Import(ctx context.Context, f domain.QuestionFile) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO question_sets VALUES(?,?,?,?,?)", f.Set.ID, f.Set.Title, f.Set.Description, canonical(f), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("保存题集：%w", err)
	}
	for i, q := range f.Questions {
		if _, err = tx.ExecContext(ctx, "INSERT INTO questions VALUES(?,?,?,?,?)", q.ID, f.Set.ID, i, canonical(q), fingerprint(q)); err != nil {
			return fmt.Errorf("保存题目 %s：%w", q.ID, err)
		}
	}
	return tx.Commit()
}
func (r *Store) SetQuestions(ctx context.Context, id string) ([]domain.Question, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT body FROM questions WHERE set_id=? ORDER BY position", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Question{}
	for rows.Next() {
		var body string
		var q domain.Question
		if err := rows.Scan(&body); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(body), &q); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}
func (r *Store) CreateSession(ctx context.Context, s domain.Session) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO quiz_sessions(id,title,mode,status,started_at) VALUES(?,?,?,?,?)", s.ID, s.Title, s.Mode, s.Status, s.StartedAt); err != nil {
		return err
	}
	for i, item := range s.Items {
		if _, err = tx.ExecContext(ctx, "INSERT INTO quiz_answers(session_id,question_id,position) VALUES(?,?,?)", s.ID, item.Question.ID, i); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (r *Store) Session(ctx context.Context, id string) (domain.Session, error) {
	var s domain.Session
	err := r.DB.QueryRowContext(ctx, "SELECT id,title,mode,status,started_at,completed_at FROM quiz_sessions WHERE id=?", id).Scan(&s.ID, &s.Title, &s.Mode, &s.Status, &s.StartedAt, &s.CompletedAt)
	if err != nil {
		return s, err
	}
	rows, err := r.DB.QueryContext(ctx, `SELECT q.body,a.selected,a.flagged,a.scored,a.correct,a.scored_at FROM quiz_answers a JOIN questions q ON q.id=a.question_id WHERE a.session_id=? ORDER BY a.position`, id)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	s.Items = []domain.SessionItem{}
	for rows.Next() {
		var i domain.SessionItem
		var body string
		var c sql.NullBool
		if err := rows.Scan(&body, &i.Selected, &i.Flagged, &i.Scored, &c, &i.ScoredAt); err != nil {
			return s, err
		}
		if c.Valid {
			i.Correct = &c.Bool
		}
		if err := json.Unmarshal([]byte(body), &i.Question); err != nil {
			return s, err
		}
		s.Items = append(s.Items, i)
	}
	return s, rows.Err()
}
func (r *Store) SaveItem(ctx context.Context, id string, i domain.SessionItem) error {
	result, err := r.DB.ExecContext(ctx, `UPDATE quiz_answers SET selected=?,flagged=?,scored=?,correct=?,scored_at=? WHERE session_id=? AND question_id=? AND scored=0 AND EXISTS(SELECT 1 FROM quiz_sessions WHERE id=? AND status='active')`, i.Selected, i.Flagged, i.Scored, i.Correct, i.ScoredAt, id, i.Question.ID, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("会话或题目状态已变更，请刷新")
	}
	return nil
}
func (r *Store) Complete(ctx context.Context, s domain.Session) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "UPDATE quiz_sessions SET status='completed',completed_at=? WHERE id=? AND status='active'", s.CompletedAt, s.ID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("会话状态已变更，请刷新")
	}
	for _, i := range s.Items {
		if _, err = tx.ExecContext(ctx, "UPDATE quiz_answers SET scored=1,correct=?,scored_at=? WHERE session_id=? AND question_id=? AND scored=0", i.Correct, i.ScoredAt, s.ID, i.Question.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (r *Store) ListSessions(ctx context.Context) ([]domain.SessionSummary, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT s.id,s.title,s.mode,s.status,s.started_at,s.completed_at,count(*),sum(a.scored),coalesce(sum(a.correct),0) FROM quiz_sessions s JOIN quiz_answers a ON a.session_id=s.id GROUP BY s.id ORDER BY s.started_at DESC,s.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SessionSummary{}
	for rows.Next() {
		var s domain.SessionSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.Mode, &s.Status, &s.StartedAt, &s.CompletedAt, &s.Total, &s.Scored, &s.Correct); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
