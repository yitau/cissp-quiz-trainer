package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// DeleteUnusedSet rechecks all session references inside the deletion transaction.
// Even unselected draft rows protect an in-progress session from losing its questions.
func (r *Store) DeleteUnusedSet(ctx context.Context, id string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var history bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM quiz_answers a JOIN questions q ON q.id=a.question_id WHERE q.set_id=s.id) FROM question_sets s WHERE s.id=?`, id).Scan(&history)
	if err == sql.ErrNoRows {
		return fmt.Errorf("题集不存在，请刷新后重试")
	}
	if err != nil {
		return err
	}
	if history {
		return fmt.Errorf("题集已有练习记录（包括未完成会话），只能归档，不能删除")
	}
	for _, query := range []string{
		"DELETE FROM question_flags WHERE question_id IN (SELECT id FROM questions WHERE set_id=?)",
		"DELETE FROM question_set_archives WHERE set_id=?",
		"DELETE FROM questions WHERE set_id=?",
		"DELETE FROM question_sets WHERE id=?",
	} {
		if _, err := tx.ExecContext(ctx, query, id); err != nil {
			return fmt.Errorf("删除题集失败，原数据已保留：%w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交删除题集事务：%w", err)
	}
	return nil
}
func (r *Store) SetArchived(ctx context.Context, id string, archived bool) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var found string
	if err := tx.QueryRowContext(ctx, "SELECT id FROM question_sets WHERE id=?", id).Scan(&found); err != nil {
		return err
	}
	if archived {
		_, err = tx.ExecContext(ctx, "INSERT INTO question_set_archives(set_id) VALUES(?) ON CONFLICT(set_id) DO NOTHING", id)
	} else {
		_, err = tx.ExecContext(ctx, "DELETE FROM question_set_archives WHERE set_id=?", id)
	}
	if err != nil {
		return fmt.Errorf("更新归档状态：%w", err)
	}
	return tx.Commit()
}
