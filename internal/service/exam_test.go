package service_test

import (
	"context"
	"sync"
	"testing"

	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
)

func TestExamRestartAndIdempotency(t *testing.T) {
	ctx := context.Background()
	s, _, db, path := setup(t)
	load(t, s)
	exam, err := s.Start(ctx, "demo-v1", "exam")
	if err != nil {
		t.Fatal(err)
	}
	hidden(t, exam)
	exam, err = s.SaveChoice(ctx, exam.ID, "DEMO-001", "B", true)
	if err != nil {
		t.Fatal(err)
	}
	hidden(t, exam)
	if _, err := s.SubmitStudy(ctx, exam.ID, "DEMO-001", "B"); err == nil {
		t.Fatal("exam graded via study endpoint")
	}
	h, err := s.ListSessions(ctx)
	if err != nil || h[0].Scored != 0 || h[0].Correct != 0 {
		t.Fatal("draft counted", h, err)
	}
	db.Close()
	db, err = database.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s = service.New(sqlite.New(db))
	exam, err = s.Session(ctx, exam.ID)
	if err != nil {
		t.Fatal(err)
	}
	hidden(t, exam)
	if exam.Items[0].Selected != "B" || !exam.Items[0].Flagged {
		t.Fatal("draft lost on restart")
	}
	exam, err = s.SaveChoice(ctx, exam.ID, "DEMO-002", "A", false)
	if err != nil {
		t.Fatal(err)
	}
	// Concurrent retries must count exactly one completion.
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for n := 0; n < 8; n++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.Complete(ctx, exam.ID); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	result, err := s.Session(ctx, exam.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || !*result.Items[0].Correct || *result.Items[1].Correct || *result.Items[2].Correct || result.Items[3].Question.Answer != "D" {
		t.Fatal("exam result incorrect")
	}
	h, err = s.ListSessions(ctx)
	if err != nil || h[0].Scored != 4 || h[0].Correct != 1 {
		t.Fatal("duplicate grading", h, err)
	}
	if _, err := s.SaveChoice(ctx, exam.ID, "DEMO-001", "A", false); err == nil {
		t.Fatal("changed completed exam")
	}
}
func TestStudyDraftIsNotSubmission(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	load(t, s)
	q, err := s.Start(ctx, "demo-v1", "study")
	if err != nil {
		t.Fatal(err)
	}
	q, err = s.SaveChoice(ctx, q.ID, "DEMO-001", "B", false)
	if err != nil {
		t.Fatal(err)
	}
	hidden(t, q)
	q, err = s.Complete(ctx, q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if q.Items[0].Selected != "" || *q.Items[0].Correct {
		t.Fatal("unsubmitted study draft counted as submitted")
	}
}
