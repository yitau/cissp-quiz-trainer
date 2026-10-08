package database

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigration003FailureRollsBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v2.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(initial + setArchives + `CREATE TABLE learning_progress(collision TEXT);`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if upgraded, err := Open(path); err == nil {
		upgraded.Close()
		t.Fatal("accepted migration collision")
	}
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var version, units int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 2 {
		t.Fatal(version, err)
	}
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='learning_units'").Scan(&units); err != nil || units != 0 {
		t.Fatal("partial migration", units, err)
	}
	var language string
	if err := db.QueryRow("SELECT value FROM app_settings WHERE key='language'").Scan(&language); err != nil || language != "zh-CN" {
		t.Fatal(language, err)
	}
}

func TestLearningForeignKeyAndStateConstraints(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "new.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("INSERT INTO learning_progress VALUES('missing','concept','in_progress',NULL,'stamp')"); err == nil {
		t.Fatal("foreign key disabled")
	}
	if _, err := db.Exec("INSERT INTO learning_units VALUES('lesson','{}','hash','stamp')"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO learning_progress VALUES('lesson','concept','not_started',NULL,'stamp')"); err == nil {
		t.Fatal("derived state persisted")
	}
}
