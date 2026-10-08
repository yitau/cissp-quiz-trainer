package sqlite

import (
	"context"
	"fmt"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

func (r *Store) ListLessons(ctx context.Context) ([]domain.StoredLesson, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id,document,fingerprint,imported_at FROM learning_units ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StoredLesson{}
	for rows.Next() {
		var s domain.StoredLesson
		if err := rows.Scan(&s.ID, &s.Document, &s.Fingerprint, &s.ImportedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
func (r *Store) Lesson(ctx context.Context, id string) (domain.StoredLesson, error) {
	var s domain.StoredLesson
	err := r.DB.QueryRowContext(ctx, "SELECT id,document,fingerprint,imported_at FROM learning_units WHERE id=?", id).Scan(&s.ID, &s.Document, &s.Fingerprint, &s.ImportedAt)
	return s, err
}
func (r *Store) ImportLesson(ctx context.Context, s domain.StoredLesson) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "INSERT INTO learning_units(id,document,fingerprint,imported_at) VALUES(?,?,?,?)", s.ID, s.Document, s.Fingerprint, s.ImportedAt); err != nil {
		return fmt.Errorf("保存课程 %s：%w", s.ID, err)
	}
	return tx.Commit()
}
func (r *Store) LessonProgress(ctx context.Context, id string) ([]domain.LessonProgress, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT unit_id,concept_id,status,coalesce(last_opened_at,''),updated_at FROM learning_progress WHERE unit_id=? ORDER BY concept_id", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LessonProgress{}
	for rows.Next() {
		var p domain.LessonProgress
		if err := rows.Scan(&p.UnitID, &p.ConceptID, &p.Status, &p.LastOpenedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *Store) SaveLessonProgress(ctx context.Context, p domain.LessonProgress, opened bool) error {
	// Opening updates the location only; explicit self-assessment updates the status only.
	_, err := r.DB.ExecContext(ctx, `INSERT INTO learning_progress(unit_id,concept_id,status,last_opened_at,updated_at) VALUES(?,?,?,nullif(?,''),?)
 ON CONFLICT(unit_id,concept_id) DO UPDATE SET status=CASE WHEN ? THEN learning_progress.status ELSE excluded.status END,
 last_opened_at=CASE WHEN ? THEN excluded.last_opened_at ELSE learning_progress.last_opened_at END,updated_at=excluded.updated_at`, p.UnitID, p.ConceptID, p.Status, p.LastOpenedAt, p.UpdatedAt, opened, opened)
	return err
}
