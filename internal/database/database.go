package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	_ "modernc.org/sqlite"
)

const Version = domain.DatabaseVersion

//go:embed migrations/001_initial.sql
var initial string

//go:embed migrations/002_set_archives.sql
var setArchives string

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("创建数据目录：%w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库：%w", err)
	}
	db.SetMaxOpenConns(1)
	if err = initialize(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
func initialize(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000;"); err != nil {
		return fmt.Errorf("设置数据库连接：%w", err)
	}
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version > Version || version < 0 {
		return fmt.Errorf("不支持数据库版本 %d，当前支持 %d", version, Version)
	}
	if version == 0 {
		var tables int
		if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables); err != nil {
			return err
		}
		if tables != 0 {
			return fmt.Errorf("未标记版本的已有数据库，拒绝自动改写")
		}
	}
	if version < Version {
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if version == 0 {
			if _, err := tx.Exec(initial); err != nil {
				return fmt.Errorf("数据库迁移 001：%w", err)
			}
		}
		if _, err := tx.Exec(setArchives); err != nil {
			return fmt.Errorf("数据库迁移 002：%w", err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
