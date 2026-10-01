package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/importer"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
)

func setup(t *testing.T) (*service.Trainer, *sqlite.Store, *sql.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := sqlite.New(db)
	return service.New(r), r, db, path
}
func sample(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("../../samples/demo-questions.json")
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func load(t *testing.T, s *service.Trainer) {
	t.Helper()
	p, err := s.Preview(context.Background(), sample(t))
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "ready" {
		t.Fatal(p)
	}
	if err := s.Import(context.Background(), p.Token); err != nil {
		t.Fatal(err)
	}
}
func hidden(t *testing.T, s domain.Session) {
	t.Helper()
	for _, i := range s.Items {
		if i.Question.Answer != "" || i.Question.Explanation != "" || i.Question.WhyCorrect != "" || i.Question.WhyOthersWrong != nil || i.Correct != nil {
			t.Fatal("solution disclosed", i.Question.ID)
		}
	}
}
func TestImportAndStudy(t *testing.T) {
	ctx := context.Background()
	s, r, db, _ := setup(t)
	load(t, s)
	p, err := s.Preview(ctx, sample(t))
	if err != nil || p.Status != "duplicate" {
		t.Fatal(p, err)
	}
	if err := s.Import(ctx, p.Token); err != nil {
		t.Fatal(err)
	}
	f, err := importer.Parse(sample(t))
	if err != nil {
		t.Fatal(err)
	}
	f.Set.Title = "changed"
	b, _ := json.Marshal(f)
	p, err = s.Preview(ctx, b)
	if err != nil || p.Status != "conflict" {
		t.Fatal(p, err)
	}
	if err := s.Import(ctx, p.Token); err == nil {
		t.Fatal("conflict accepted")
	}
	f.Set.ID = "other"
	b, _ = json.Marshal(f)
	p, err = s.Preview(ctx, b)
	if err != nil || p.Status != "conflict" {
		t.Fatal(p, err)
	}
	// Bypass preview to force a repository failure after the set row was written.
	if err := r.Import(ctx, f); err == nil {
		t.Fatal("expected duplicate question constraint")
	}
	sets, err := s.ListSets(ctx)
	if err != nil || len(sets) != 1 {
		t.Fatal("partial import", sets, err)
	}
	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil || fk != 1 {
		t.Fatal("foreign keys off", err)
	}
	session, err := s.Start(ctx, "demo-v1", "study")
	if err != nil {
		t.Fatal(err)
	}
	hidden(t, session)
	session, err = s.SubmitStudy(ctx, session.ID, "DEMO-001", "B")
	if err != nil {
		t.Fatal(err)
	}
	if session.Items[0].Correct == nil || !*session.Items[0].Correct || session.Items[0].Question.Answer != "B" {
		t.Fatal("wrong scoring")
	}
	if session.Items[1].Question.Answer != "" {
		t.Fatal("unsubmitted answer leaked")
	}
	if _, err = s.SubmitStudy(ctx, session.ID, "DEMO-001", "B"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SubmitStudy(ctx, session.ID, "DEMO-001", "A"); err == nil {
		t.Fatal("changed submitted answer")
	}
	finished, err := s.Complete(ctx, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if finished.Status != "completed" || *finished.Items[1].Correct {
		t.Fatal("unanswered scoring")
	}
	if _, err = s.Complete(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	history, err := s.ListSessions(ctx)
	if err != nil || history[0].Scored != 4 || history[0].Correct != 1 {
		t.Fatal(history, err)
	}
}
func TestMigrationPreservesUnknownDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unknown.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TABLE private_data(value TEXT); INSERT INTO private_data VALUES('keep')"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if _, err := database.Open(path); err == nil {
		t.Fatal("rewrote unversioned database")
	}
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var value string
	if err := db.QueryRow("SELECT value FROM private_data").Scan(&value); err != nil || value != "keep" {
		t.Fatal(value, err)
	}
}
