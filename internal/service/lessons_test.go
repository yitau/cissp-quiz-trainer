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
	"testing"

	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/importer"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
)

func lessonData(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../../samples/cissp-week01-day01-lesson.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func loadLesson(t *testing.T, s *service.Trainer) domain.LessonPreview {
	t.Helper()
	p, e := s.PreviewLesson(context.Background(), lessonData(t))
	if e != nil {
		t.Fatal(e)
	}
	if e := s.ImportLesson(context.Background(), p.Token); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestLessonWorkflowRestartBackup(t *testing.T) {
	ctx := context.Background()
	s, _, db, path := setup(t)
	load(t, s)
	before, _ := s.Statistics(ctx)
	p := loadLesson(t, s)
	id := p.Lesson.ID
	if p.Status != "ready" || p.Lesson.Week != 1 || p.Lesson.Day != 1 || p.Lesson.Domain != 1 || len(p.Sections) != 2 || p.ConceptCount != 8 {
		t.Fatal(p)
	}
	v, e := s.GetLesson(ctx, id)
	if e != nil {
		t.Fatal(e)
	}
	if len(v.Progress) != 0 || v.Summary.Understood != 0 || v.LinkedSetStatus != "missing" {
		t.Fatal(v)
	}
	first := v.Summary.ResumeConceptID
	second := v.Document.Sections[0].ConceptIDs[1]
	v, e = s.OpenConcept(ctx, id, first)
	if e != nil || len(v.Progress) != 1 || v.Progress[0].Status != "in_progress" {
		t.Fatal(v, e)
	}
	for _, status := range []string{"understood", "needs_review", "in_progress", "understood"} {
		v, e = s.SetConceptStatus(ctx, id, first, status)
		if e != nil || len(v.Progress) != 1 || v.Progress[0].Status != status {
			t.Fatal(v, e)
		}
	}
	v, e = s.OpenConcept(ctx, id, second)
	if e != nil {
		t.Fatal(e)
	}
	v, e = s.SetConceptStatus(ctx, id, second, "needs_review")
	if e != nil {
		t.Fatal(e)
	}
	if v.Summary.Understood != 1 || v.Summary.Completed || v.Summary.ResumeConceptID != second {
		t.Fatal(v.Summary)
	}
	db.Close()
	db, e = database.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s = service.New(sqlite.New(db))
	restarted, e := s.GetLesson(ctx, id)
	if e != nil || !reflect.DeepEqual(v, restarted) {
		t.Fatal("restart differs", e)
	}
	v, e = s.OpenConcept(ctx, id, second)
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range v.Progress {
		if p.ConceptID == second && p.Status != "needs_review" {
			t.Fatal("opening downgraded status")
		}
	}
	var reordered map[string]any
	json.Unmarshal(lessonData(t), &reordered)
	reencoded, _ := json.MarshalIndent(reordered, "", "    ")
	duplicate, e := s.PreviewLesson(ctx, reencoded)
	if e != nil || duplicate.Status != "duplicate" {
		t.Fatal(duplicate, e)
	}
	if e := s.ImportLesson(ctx, duplicate.Token); e != nil {
		t.Fatal(e)
	}
	unchanged, _ := s.GetLesson(ctx, id)
	if !reflect.DeepEqual(v, unchanged) {
		t.Fatal("duplicate changed progress")
	}
	f, _ := importer.ParseLesson(lessonData(t))
	f.Lesson.Title = "changed"
	b, _ := json.Marshal(f)
	conflict, e := s.PreviewLesson(ctx, b)
	if e != nil || conflict.Status != "conflict" {
		t.Fatal(conflict, e)
	}
	if e := s.ImportLesson(ctx, conflict.Token); e == nil {
		t.Fatal("overwrote conflicting course")
	}
	unchanged, _ = s.GetLesson(ctx, id)
	if !reflect.DeepEqual(v, unchanged) {
		t.Fatal("conflict changed data")
	}
	if _, e := s.SetConceptStatus(ctx, id, first, "not_started"); e == nil {
		t.Fatal("accepted derived state")
	}
	if _, e := s.OpenConcept(ctx, id, "unknown"); e == nil {
		t.Fatal("accepted unknown concept")
	}
	after, _ := s.Statistics(ctx)
	sessions, _ := s.ListSessions(ctx)
	if !reflect.DeepEqual(before, after) || len(sessions) != 0 {
		t.Fatal("reading polluted quiz data")
	}
	backup := filepath.Join(t.TempDir(), "lesson.zip")
	if e := s.Backup(ctx, backup); e != nil {
		t.Fatal(e)
	}
	target, _, _, _ := setup(t)
	safety, e := target.Restore(ctx, backup, t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(safety); e != nil {
		t.Fatal(e)
	}
	restored, e := target.GetLesson(ctx, id)
	if e != nil || !reflect.DeepEqual(v, restored) {
		t.Fatal("backup mismatch", e)
	}
	targetStats, _ := target.Statistics(ctx)
	if !reflect.DeepEqual(before, targetStats) {
		t.Fatal("quiz backup mismatch")
	}
	for _, c := range v.Document.Concepts {
		v, e = s.SetConceptStatus(ctx, id, c.ID, "understood")
		if e != nil {
			t.Fatal(e)
		}
	}
	if !v.Summary.Completed || v.Summary.Understood != 8 || v.Summary.ConceptCount != 8 {
		t.Fatal(v.Summary)
	}
	v, e = s.SetConceptStatus(ctx, id, first, "needs_review")
	if e != nil || v.Summary.Completed || v.Summary.Understood != 7 {
		t.Fatal(v.Summary, e)
	}
}

func TestLessonCancelInvalidAndTransaction(t *testing.T) {
	ctx := context.Background()
	s, _, db, _ := setup(t)
	p, e := s.PreviewLesson(ctx, lessonData(t))
	if e != nil {
		t.Fatal(e)
	}
	list, _ := s.ListLessons(ctx)
	if len(list) != 0 {
		t.Fatal("preview wrote course")
	}
	s.CancelLessonImport()
	if e := s.ImportLesson(ctx, p.Token); e == nil {
		t.Fatal("cancelled token remained valid")
	}
	p, _ = s.PreviewLesson(ctx, lessonData(t))
	if _, e := s.PreviewLesson(ctx, []byte(`{}`)); e == nil {
		t.Fatal("invalid preview accepted")
	}
	if e := s.ImportLesson(ctx, p.Token); e == nil {
		t.Fatal("stale preview remained valid")
	}
	if _, e := db.Exec(`CREATE TRIGGER fail_lesson AFTER INSERT ON learning_units BEGIN SELECT RAISE(ABORT,'simulated failure'); END;`); e != nil {
		t.Fatal(e)
	}
	p, _ = s.PreviewLesson(ctx, lessonData(t))
	if e := s.ImportLesson(ctx, p.Token); e == nil {
		t.Fatal("expected failure")
	}
	for _, table := range []string{"learning_units", "learning_progress"} {
		var n int
		if e := db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); e != nil || n != 0 {
			t.Fatal(table, n, e)
		}
	}
}

func TestLessonLinkedQuiz(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	p := loadLesson(t, s)
	id := *p.Lesson.LinkedQuestionSetID
	if _, e := s.Start(ctx, id, "study"); e == nil {
		t.Fatal("started missing linked set")
	}
	f, _ := importer.Parse(sample(t))
	f.Set.ID = id
	b, _ := json.Marshal(f)
	q, e := s.Preview(ctx, b)
	if e != nil {
		t.Fatal(e)
	}
	if e := s.Import(ctx, q.Token); e != nil {
		t.Fatal(e)
	}
	v, _ := s.GetLesson(ctx, p.Lesson.ID)
	if v.LinkedSetStatus != "available" {
		t.Fatal(v.LinkedSetStatus)
	}
	for _, mode := range []string{"study", "exam"} {
		quiz, e := s.Start(ctx, id, mode)
		if e != nil || len(quiz.Items) != len(f.Questions) {
			t.Fatal(quiz, e)
		}
		hidden(t, quiz)
	}
	if e := s.SetArchived(ctx, id, true); e != nil {
		t.Fatal(e)
	}
	v, _ = s.GetLesson(ctx, p.Lesson.ID)
	if v.LinkedSetStatus != "archived" {
		t.Fatal(v.LinkedSetStatus)
	}
	if _, e := s.Start(ctx, id, "exam"); e == nil {
		t.Fatal("started archived set")
	}
	if e := s.SetArchived(ctx, id, false); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Start(ctx, id, "exam"); e != nil {
		t.Fatal(e)
	}
}

// Produce a genuine DB v2 snapshot, not merely an old App label on a v3 schema.
func legacyV2Entries(t *testing.T, entries map[string][]byte, version string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "v2.db")
	if e := os.WriteFile(path, entries["cissp.db"], 0600); e != nil {
		t.Fatal(e)
	}
	db, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec("DROP TABLE learning_progress; DROP TABLE learning_units; PRAGMA user_version=2"); e != nil {
		t.Fatal(e)
	}
	db.Close()
	entries["cissp.db"], e = os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var meta domain.BackupMetadata
	if e := json.Unmarshal(entries["metadata.json"], &meta); e != nil {
		t.Fatal(e)
	}
	meta.AppVersion = version
	meta.DatabaseVersion = 2
	meta.DatabaseSHA256 = fmt.Sprintf("%x", sha256.Sum256(entries["cissp.db"]))
	entries["metadata.json"], _ = json.Marshal(meta)
}
func TestLessonV2MigrationAndRestore(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	load(t, s)
	quiz, e := s.Start(ctx, "demo-v1", "study")
	if e != nil {
		t.Fatal(e)
	}
	if _, e := s.SubmitStudy(ctx, quiz.ID, "DEMO-001", "A"); e != nil {
		t.Fatal(e)
	}
	s.SetFavorite(ctx, "DEMO-001", true)
	s.SetArchived(ctx, "demo-v1", true)
	before, _ := s.Statistics(ctx)
	sets, _ := s.ListSets(ctx)
	sessions, _ := s.ListSessions(ctx)
	backup := filepath.Join(t.TempDir(), "source.zip")
	if e := s.Backup(ctx, backup); e != nil {
		t.Fatal(e)
	}
	for _, version := range []string{"0.1.1", "0.1.2"} {
		t.Run(version, func(t *testing.T) {
			entries := zipEntries(t, backup)
			legacyV2Entries(t, entries, version)
			path := filepath.Join(t.TempDir(), "startup.db")
			os.WriteFile(path, entries["cissp.db"], 0600)
			db, e := database.Open(path)
			if e != nil {
				t.Fatal(e)
			}
			defer db.Close()
			migrated := service.New(sqlite.New(db))
			got, _ := migrated.Statistics(ctx)
			if !reflect.DeepEqual(before, got) {
				t.Fatal("startup migration lost statistics")
			}
			migratedSets, _ := migrated.ListSets(ctx)
			if !reflect.DeepEqual(sets, migratedSets) {
				t.Fatal("startup lost archive")
			}
			archive := filepath.Join(t.TempDir(), "legacy.zip")
			writeZip(t, archive, entries)
			original, _ := os.ReadFile(archive)
			target, _, _, _ := setup(t)
			loadLesson(t, target)
			safety, e := target.Restore(ctx, archive, t.TempDir())
			if e != nil {
				t.Fatal(e)
			}
			after, _ := target.Statistics(ctx)
			newSets, _ := target.ListSets(ctx)
			newSessions, _ := target.ListSessions(ctx)
			lessons, _ := target.ListLessons(ctx)
			if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(sets, newSets) || !reflect.DeepEqual(sessions, newSessions) || len(lessons) != 0 {
				t.Fatal("legacy restore mismatch")
			}
			unchanged, _ := os.ReadFile(archive)
			if !reflect.DeepEqual(original, unchanged) {
				t.Fatal("modified original backup")
			}
			if _, e := target.Restore(ctx, safety, t.TempDir()); e != nil {
				t.Fatal(e)
			}
			lessons, _ = target.ListLessons(ctx)
			if len(lessons) != 1 {
				t.Fatal("safety backup lost course")
			}
		})
	}
}

func TestLessonCorruptBackupAndRollback(t *testing.T) {
	ctx := context.Background()
	s, r, db, _ := setup(t)
	p := loadLesson(t, s)
	s.OpenConcept(ctx, p.Lesson.ID, p.Sections[0].ConceptIDs[0])
	before, _ := s.GetLesson(ctx, p.Lesson.ID)
	good := filepath.Join(t.TempDir(), "good.zip")
	if e := s.Backup(ctx, good); e != nil {
		t.Fatal(e)
	}
	for _, command := range []string{
		"UPDATE learning_units SET fingerprint='bad'",
		"UPDATE learning_units SET document='{}'",
		"UPDATE learning_progress SET concept_id='missing'",
		"UPDATE learning_progress SET last_opened_at='invalid'",
		"PRAGMA ignore_check_constraints=ON; UPDATE learning_progress SET status='not_started'",
	} {
		t.Run(command, func(t *testing.T) {
			entries := zipEntries(t, good)
			path := filepath.Join(t.TempDir(), "bad.db")
			os.WriteFile(path, entries["cissp.db"], 0600)
			badDB, e := sql.Open("sqlite", path)
			if e != nil {
				t.Fatal(e)
			}
			if _, e := badDB.Exec(command); e != nil {
				t.Fatal(e)
			}
			badDB.Close()
			entries["cissp.db"], _ = os.ReadFile(path)
			var meta domain.BackupMetadata
			json.Unmarshal(entries["metadata.json"], &meta)
			meta.DatabaseSHA256 = fmt.Sprintf("%x", sha256.Sum256(entries["cissp.db"]))
			entries["metadata.json"], _ = json.Marshal(meta)
			bad := filepath.Join(t.TempDir(), "bad.zip")
			writeZip(t, bad, entries)
			if _, e := s.Restore(ctx, bad, t.TempDir()); e == nil {
				t.Fatal("accepted bad course backup")
			}
			after, _ := s.GetLesson(ctx, p.Lesson.ID)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("invalid restore modified data")
			}
		})
	}
	snapshot := filepath.Join(t.TempDir(), "snapshot.db")
	if e := r.Snapshot(ctx, snapshot); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec("CREATE TRIGGER fail_progress BEFORE INSERT ON learning_progress BEGIN SELECT RAISE(ABORT,'restore failure'); END"); e != nil {
		t.Fatal(e)
	}
	if e := r.RestoreSnapshot(ctx, snapshot); e == nil {
		t.Fatal("expected restore rollback")
	}
	after, _ := s.GetLesson(ctx, p.Lesson.ID)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("rollback changed course data")
	}
}
