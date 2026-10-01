package service_test

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFullWorkflowBackupRestore(t *testing.T) {
	ctx := context.Background()
	s, _, db, path := setup(t)
	load(t, s)
	study, err := s.Start(ctx, "demo-v1", "study")
	if err != nil {
		t.Fatal(err)
	}
	study, err = s.SubmitStudy(ctx, study.ID, "DEMO-001", "A")
	if err != nil {
		t.Fatal(err)
	}
	if study.Items[0].Question.Explanation == "" {
		t.Fatal("missing explanation")
	}
	if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	exam, err := s.Start(ctx, "demo-v1", "exam")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveChoice(ctx, exam.ID, "DEMO-002", "C", true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Complete(ctx, exam.ID); err != nil {
		t.Fatal(err)
	}
	before, err := s.Statistics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if before.Total.Count != 5 || before.Total.Correct != 1 || before.Favorites != 1 || before.WrongQuestions != 3 {
		t.Fatal(before)
	}
	db.Close()
	db, err = database.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s = service.New(sqlite.New(db))
	after, err := s.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("restart mismatch", after, err)
	}
	backup := filepath.Join(t.TempDir(), "full.zip")
	if err := s.Backup(ctx, backup); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Backup(ctx, backup); err == nil {
		t.Fatal("overwrote existing backup")
	}
	again, _ := os.ReadFile(backup)
	if !reflect.DeepEqual(original, again) {
		t.Fatal("changed existing backup")
	}
	target, _, targetDB, targetPath := setup(t)
	safeDir := t.TempDir()
	safety, err := target.Restore(ctx, backup, safeDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(safety); err != nil {
		t.Fatal("missing safety backup", err)
	}
	after, err = target.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("restored statistics mismatch", after, err)
	}
	restoredStudy, err := target.Session(ctx, study.ID)
	if err != nil || restoredStudy.Status != "active" || !restoredStudy.Items[0].Scored || restoredStudy.Items[1].Question.Answer != "" {
		t.Fatal(restoredStudy, err)
	}
	targetDB.Close()
	targetDB, err = database.Open(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	defer targetDB.Close()
	target = service.New(sqlite.New(targetDB))
	after, err = target.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("restore restart mismatch", err)
	}
	// Restore the automatically created empty safety backup to prove it is usable.
	if _, err := target.Restore(ctx, safety, safeDir); err != nil {
		t.Fatal(err)
	}
	sets, err := target.ListSets(ctx)
	if err != nil || len(sets) != 0 {
		t.Fatal("safety backup mismatch", sets, err)
	}
}
func zipEntries(t *testing.T, path string) map[string][]byte {
	t.Helper()
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	entries := map[string][]byte{}
	for _, f := range r.File {
		reader, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		entries[f.Name] = b
	}
	return entries
}
func writeZip(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for k, b := range entries {
		o, err := w.Create(k)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := o.Write(b); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestInvalidBackupPreservesCurrentData(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	load(t, s)
	q, err := s.Start(ctx, "demo-v1", "study")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SubmitStudy(ctx, q.ID, "DEMO-001", "B"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	good := filepath.Join(dir, "good.zip")
	if err := s.Backup(ctx, good); err != nil {
		t.Fatal(err)
	}
	before, err := s.Statistics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"version", "checksum", "path", "database-version", "score", "config", "broken-sqlite"} {
		t.Run(name, func(t *testing.T) {
			entries := zipEntries(t, good)
			var meta domain.BackupMetadata
			if err := json.Unmarshal(entries["metadata.json"], &meta); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "version":
				meta.FormatVersion = 99
			case "checksum":
				entries["cissp.db"] = []byte("corrupt")
			case "path":
				entries["../unexpected"] = entries["config.json"]
				delete(entries, "config.json")
			case "config":
				entries["config.json"] = []byte(`{"language":"other"}`)
				meta.ConfigSHA256 = fmt.Sprintf("%x", sha256.Sum256(entries["config.json"]))
			case "broken-sqlite":
				entries["cissp.db"] = []byte("invalid SQLite")
				meta.DatabaseSHA256 = fmt.Sprintf("%x", sha256.Sum256(entries["cissp.db"]))
			default:
				file := filepath.Join(t.TempDir(), "tampered.db")
				if err := os.WriteFile(file, entries["cissp.db"], 0600); err != nil {
					t.Fatal(err)
				}
				db, err := sql.Open("sqlite", file)
				if err != nil {
					t.Fatal(err)
				}
				command := "PRAGMA user_version=99"
				if name == "score" {
					command = "UPDATE quiz_answers SET correct=0 WHERE scored=1"
				}
				if _, err := db.Exec(command); err != nil {
					t.Fatal(err)
				}
				db.Close()
				entries["cissp.db"], err = os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				meta.DatabaseSHA256 = fmt.Sprintf("%x", sha256.Sum256(entries["cissp.db"]))
			}
			entries["metadata.json"], _ = json.Marshal(meta)
			bad := filepath.Join(t.TempDir(), "bad.zip")
			writeZip(t, bad, entries)
			if _, err := s.Restore(ctx, bad, t.TempDir()); err == nil {
				t.Fatal("accepted invalid backup")
			}
			after, err := s.Statistics(ctx)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("original records modified", err)
			}
		})
	}
	broken := filepath.Join(dir, "broken.zip")
	if err := os.WriteFile(broken, []byte("not zip"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Restore(ctx, broken, t.TempDir()); err == nil {
		t.Fatal("accepted broken zip")
	}
	// Failure to create the safety backup must stop before modifying the active DB.
	if _, err := s.Restore(ctx, good, broken); err == nil {
		t.Fatal("restored without safety backup")
	}
	after, err := s.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("data changed after failed safety backup", err)
	}
}
func TestRestoreTransactionRollback(t *testing.T) {
	ctx := context.Background()
	s, r, db, _ := setup(t)
	load(t, s)
	snapshot := filepath.Join(t.TempDir(), "snapshot.db")
	if err := r.Snapshot(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TRIGGER fail_restore BEFORE INSERT ON questions BEGIN SELECT RAISE(ABORT,'simulated write failure'); END;"); err != nil {
		t.Fatal(err)
	}
	if err := r.RestoreSnapshot(ctx, snapshot); err == nil {
		t.Fatal("expected transactional write failure")
	}
	stats, err := s.Statistics(ctx)
	if err != nil || stats.Favorites != 1 {
		t.Fatal("rollback lost favorites", stats, err)
	}
	sets, err := s.ListSets(ctx)
	if err != nil || len(sets) != 1 {
		t.Fatal("rollback lost question set", sets, err)
	}
}
