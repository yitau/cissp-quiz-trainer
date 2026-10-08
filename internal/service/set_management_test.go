package service_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/importer"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
)

func TestDeleteUnusedSetAndReimport(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	load(t, s)
	if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	p, err := s.Preview(ctx, sample(t))
	if err != nil {
		t.Fatal(err)
	}
	sets, err := s.ListSets(ctx)
	if err != nil || len(sets) != 1 || sets[0].HasHistory {
		t.Fatal(sets, err)
	}
	if err := s.DeleteUnusedSet(ctx, "demo-v1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Import(ctx, p.Token); err == nil {
		t.Fatal("stale preview recreated deleted set")
	}
	sets, err = s.ListSets(ctx)
	if err != nil || len(sets) != 0 {
		t.Fatal(sets, err)
	}
	stats, err := s.Statistics(ctx)
	if err != nil || stats.Favorites != 0 || stats.Total.Count != 0 {
		t.Fatal(stats, err)
	}
	load(t, s) // The same stable IDs can be imported again after a genuine unused deletion.
	sets, err = s.ListSets(ctx)
	if err != nil || len(sets) != 1 || sets[0].Count != 4 {
		t.Fatal(sets, err)
	}
	if err := s.DeleteUnusedSet(ctx, "missing"); err == nil {
		t.Fatal("missing set silently deleted")
	}
}

func TestDeleteProtectsEverySessionReference(t *testing.T) {
	for _, kind := range []string{"empty-study", "exam-draft", "scored-study", "completed-exam", "favorite-review"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s, r, _, _ := setup(t)
			load(t, s)
			mode := "study"
			if kind == "exam-draft" || kind == "completed-exam" {
				mode = "exam"
			}
			var session domain.Session
			var err error
			if kind == "favorite-review" {
				if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
					t.Fatal(err)
				}
				session, err = s.StartReview(ctx, "favorites")
			} else {
				session, err = s.Start(ctx, "demo-v1", mode)
			}
			if err != nil {
				t.Fatal(err)
			}
			if kind == "exam-draft" {
				_, err = s.SaveChoice(ctx, session.ID, "DEMO-001", "B", true)
			}
			if kind == "scored-study" {
				_, err = s.SubmitStudy(ctx, session.ID, "DEMO-001", "A")
			}
			if kind == "completed-exam" {
				_, err = s.Complete(ctx, session.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			before, err := s.Session(ctx, session.ID)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.DeleteUnusedSet(ctx, "demo-v1"); err == nil {
				t.Fatal("deleted used set")
			}
			if err := r.DeleteUnusedSet(ctx, "demo-v1"); err == nil {
				t.Fatal("repository did not guard deletion")
			}
			after, err := s.Session(ctx, session.ID)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("history changed", err)
			}
			sets, err := s.ListSets(ctx)
			if err != nil || !sets[0].HasHistory {
				t.Fatal(sets, err)
			}
		})
	}
}

func TestDeleteSetTransactionRollback(t *testing.T) {
	ctx := context.Background()
	s, _, db, _ := setup(t)
	load(t, s)
	if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TRIGGER prevent_set_delete BEFORE DELETE ON question_sets BEGIN SELECT RAISE(ABORT,'simulated delete failure'); END;"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteUnusedSet(ctx, "demo-v1"); err == nil {
		t.Fatal("expected delete failure")
	}
	sets, err := s.ListSets(ctx)
	if err != nil || len(sets) != 1 || sets[0].Count != 4 {
		t.Fatal("partial deletion", sets, err)
	}
	stats, err := s.Statistics(ctx)
	if err != nil || stats.Favorites != 1 {
		t.Fatal("favorite lost on rollback", stats, err)
	}
}

func TestArchivePreservesHistoryReviewsAndBackup(t *testing.T) {
	ctx := context.Background()
	s, _, db, path := setup(t)
	load(t, s)
	study, err := s.Start(ctx, "demo-v1", "study")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SubmitStudy(ctx, study.ID, "DEMO-001", "A"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	before, err := s.Statistics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	history, err := s.ListSessions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 2; n++ {
		if err := s.SetArchived(ctx, "demo-v1", true); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Start(ctx, "demo-v1", "exam"); err == nil {
		t.Fatal("started archived set")
	}
	after, err := s.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("archive changed statistics", err)
	}
	h, err := s.ListSessions(ctx)
	if err != nil || !reflect.DeepEqual(history, h) {
		t.Fatal("archive changed sessions", err)
	}
	p, err := s.Preview(ctx, sample(t))
	if err != nil || p.Status != "duplicate" {
		t.Fatal(p, err)
	}
	if err := s.Import(ctx, p.Token); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = database.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s = service.New(sqlite.New(db))
	sets, err := s.ListSets(ctx)
	if err != nil || !sets[0].Archived || !sets[0].HasHistory {
		t.Fatal("archive lost after restart/import", sets, err)
	}
	backup := filepath.Join(t.TempDir(), "archived.zip")
	if err := s.Backup(ctx, backup); err != nil {
		t.Fatal(err)
	}
	target, _, _, _ := setup(t)
	if _, err := target.Restore(ctx, backup, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	sets, err = target.ListSets(ctx)
	if err != nil || !sets[0].Archived {
		t.Fatal("archive not restored", sets, err)
	}
	after, err = target.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("restore changed stats", err)
	}
	if _, err := target.SubmitStudy(ctx, study.ID, "DEMO-002", "C"); err != nil {
		t.Fatal("cannot resume archived session", err)
	}
	for _, kind := range []string{"wrong", "favorites"} {
		review, err := target.StartReview(ctx, kind)
		if err != nil || len(review.Items) != 1 {
			t.Fatal("archive hid review", kind, review, err)
		}
	}
	if err := target.DeleteUnusedSet(ctx, "demo-v1"); err == nil {
		t.Fatal("deleted archived history")
	}
	for n := 0; n < 2; n++ {
		if err := target.SetArchived(ctx, "demo-v1", false); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := target.Start(ctx, "demo-v1", "study"); err != nil {
		t.Fatal("unarchive did not restore start", err)
	}
}

// Builds the actual v1 schema from its unchanged migration, not a mock repository.
func legacyDatabase(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	ddl, err := os.ReadFile("../database/migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	// Older Windows builds may embed CRLF migrations; schema validation must tolerate line endings.
	legacySQL := strings.ReplaceAll(strings.ReplaceAll(string(ddl), "\r\n", "\n"), "\n", "\r\n")
	if _, err := db.Exec(legacySQL); err != nil {
		t.Fatal(err)
	}
	f, err := importer.Parse(sample(t))
	if err != nil {
		t.Fatal(err)
	}
	r := sqlite.New(db)
	if err := r.Import(ctx, f); err != nil {
		t.Fatal(err)
	}
	session := domain.Session{ID: "legacy-study", Title: f.Set.Title, Mode: "study", Status: "active", StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	for _, q := range f.Questions {
		session.Items = append(session.Items, domain.SessionItem{Question: q})
	}
	if err := r.CreateSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	svc := service.New(r)
	if _, err := svc.SubmitStudy(ctx, session.ID, "DEMO-001", "B"); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMigrationV1AndLegacyBackup(t *testing.T) {
	ctx := context.Background()
	path := legacyDatabase(t)
	legacy, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	config := []byte(`{"language":"zh-CN"}`)
	meta := domain.BackupMetadata{FormatVersion: 1, AppVersion: "0.1.0", DatabaseVersion: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), DatabaseSHA256: fmt.Sprintf("%x", sha256.Sum256(legacy)), ConfigSHA256: fmt.Sprintf("%x", sha256.Sum256(config))}
	metadata, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "legacy.zip")
	writeZip(t, archive, map[string][]byte{"metadata.json": metadata, "config.json": config, "cissp.db": legacy})
	original, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	// Normal startup upgrades old data in place without losing records.
	db, err := database.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := service.New(sqlite.New(db))
	stats, err := svc.Statistics(ctx)
	if err != nil || stats.Total.Correct != 1 || stats.Favorites != 1 {
		t.Fatal("migration lost data", stats, err)
	}
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != domain.DatabaseVersion {
		t.Fatal(version, err)
	}
	sets, err := svc.ListSets(ctx)
	if err != nil || len(sets) != 1 || sets[0].Archived || !sets[0].HasHistory {
		t.Fatal(sets, err)
	}
	target, _, _, _ := setup(t)
	load(t, target)
	if err := target.SetArchived(ctx, "demo-v1", true); err != nil {
		t.Fatal(err)
	}
	if _, err := target.Restore(ctx, archive, t.TempDir()); err != nil {
		t.Fatal("legacy restore", err)
	}
	restored, err := target.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(stats, restored) {
		t.Fatal(restored, err)
	}
	sets, err = target.ListSets(ctx)
	if err != nil || sets[0].Archived {
		t.Fatal("old backup inherited current archive state", sets, err)
	}
	unchanged, err := os.ReadFile(archive)
	if err != nil || !reflect.DeepEqual(original, unchanged) {
		t.Fatal("modified original backup", err)
	}
	// A falsely labelled archive must not be upgraded or replace active data.
	meta.DatabaseVersion = 2
	meta.AppVersion = domain.AppVersion
	metadata, _ = json.Marshal(meta)
	bad := filepath.Join(t.TempDir(), "mismatch.zip")
	writeZip(t, bad, map[string][]byte{"metadata.json": metadata, "config.json": config, "cissp.db": legacy})
	if _, err := target.Restore(ctx, bad, t.TempDir()); err == nil {
		t.Fatal("accepted metadata/DB mismatch")
	}
	restored, err = target.Statistics(ctx)
	if err != nil || !reflect.DeepEqual(stats, restored) {
		t.Fatal("failed restore changed data", err)
	}
}
