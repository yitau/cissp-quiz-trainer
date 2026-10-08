package service_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
)

func TestProgressAndResumeAfterRestart(t *testing.T) {
	for _, mode := range []string{"study", "exam"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			s, _, db, path := setup(t)
			load(t, s)
			q, err := s.Start(ctx, "demo-v1", mode)
			if err != nil {
				t.Fatal(err)
			}
			if q.ResumeIndex != 0 || q.Items[0].Progress != "unanswered" {
				t.Fatal(q)
			}
			q, err = s.SaveChoice(ctx, q.ID, "DEMO-001", "B", false)
			if err != nil {
				t.Fatal(err)
			}
			hidden(t, q)
			wantIndex, wantProgress := 1, "selected"
			if mode == "study" {
				wantIndex, wantProgress = 0, "draft"
			}
			if q.ResumeIndex != wantIndex || q.Items[0].Progress != wantProgress {
				t.Fatal(q)
			}
			h, err := s.ListSessions(ctx)
			if err != nil || h[0].Selected != 1 || h[0].Scored != 0 || h[0].Correct != 0 {
				t.Fatal(h, err)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			db, err = database.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			s = service.New(sqlite.New(db))
			q, err = s.Session(ctx, q.ID)
			if err != nil || q.ResumeIndex != wantIndex || q.Items[0].Progress != wantProgress {
				t.Fatal(q, err)
			}
			hidden(t, q)
			for n, item := range q.Items {
				q, err = s.SaveChoice(ctx, q.ID, item.Question.ID, "B", n == 2)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "study" {
					q, err = s.SubmitStudy(ctx, q.ID, item.Question.ID, "B")
					if err != nil || q.Items[n].Progress != "submitted" {
						t.Fatal(q, err)
					}
				}
			}
			if q.ResumeIndex != 2 {
				t.Fatal("all done should resume flagged item", q.ResumeIndex)
			}
			if mode == "exam" {
				hidden(t, q)
				h, err = s.ListSessions(ctx)
				if err != nil || h[0].Selected != 4 || h[0].Scored != 0 {
					t.Fatal(h, err)
				}
				q, err = s.SaveChoice(ctx, q.ID, "DEMO-003", "B", false)
				if err != nil || q.ResumeIndex != 0 {
					t.Fatal(q, err)
				}
				q, err = s.SaveChoice(ctx, q.ID, "DEMO-002", "", false)
				if err != nil || q.ResumeIndex != 1 || q.Items[1].Progress != "unanswered" {
					t.Fatal(q, err)
				}
			}
			q, err = s.Complete(ctx, q.ID)
			if err != nil || q.ResumeIndex != 0 {
				t.Fatal(q, err)
			}
		})
	}
}

func TestSessionMistakesReviewPreservesResults(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	load(t, s)
	for _, mode := range []string{"study", "exam"} {
		t.Run(mode, func(t *testing.T) {
			q, err := s.Start(ctx, "demo-v1", mode)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.StartSessionReview(ctx, q.ID); err == nil {
				t.Fatal("active session review accepted")
			}
			for n, answer := range []string{"B", "A"} {
				if mode == "study" {
					q, err = s.SubmitStudy(ctx, q.ID, q.Items[n].Question.ID, answer)
				} else {
					q, err = s.SaveChoice(ctx, q.ID, q.Items[n].Question.ID, answer, true)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			q, err = s.Complete(ctx, q.ID)
			if err != nil {
				t.Fatal(err)
			}
			if q.Items[0].Progress != "correct" || q.Items[1].Progress != "wrong" || q.Items[2].Progress != "omitted" {
				t.Fatal(q)
			}
			before, err := s.Statistics(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.SetArchived(ctx, "demo-v1", true); err != nil {
				t.Fatal(err)
			}
			review, err := s.StartSessionReview(ctx, q.ID)
			if err != nil || review.ID == q.ID || review.Mode != "study" || len(review.Items) != 3 {
				t.Fatal(review, err)
			}
			hidden(t, review)
			for n, item := range review.Items {
				if item.Question.ID != q.Items[n+1].Question.ID || item.Selected != "" || item.Flagged || item.Scored || item.Progress != "unanswered" {
					t.Fatal(item)
				}
			}
			after, err := s.Statistics(ctx)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("creating review changed statistics", err)
			}
			if _, err := s.SubmitStudy(ctx, review.ID, review.Items[0].Question.ID, "C"); err != nil {
				t.Fatal(err)
			}
			original, err := s.Session(ctx, q.ID)
			if err != nil || !reflect.DeepEqual(q, original) {
				t.Fatal("review changed original result", err)
			}
			if err := s.SetArchived(ctx, "demo-v1", false); err != nil {
				t.Fatal(err)
			}
		})
	}
	if _, err := s.StartSessionReview(ctx, "missing"); err == nil {
		t.Fatal("missing session accepted")
	}
	q, err := s.Start(ctx, "demo-v1", "exam")
	if err != nil {
		t.Fatal(err)
	}
	for n, answer := range []string{"B", "C", "A", "D"} {
		q, err = s.SaveChoice(ctx, q.ID, q.Items[n].Question.ID, answer, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Complete(ctx, q.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartSessionReview(ctx, q.ID); err == nil {
		t.Fatal("empty review accepted")
	}
}

func TestVersion011BackupPreservesResumeAndReview(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	load(t, s)
	q, err := s.Start(ctx, "demo-v1", "exam")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveChoice(ctx, q.ID, "DEMO-001", "B", false); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "current.zip")
	if err := s.Backup(ctx, path); err != nil {
		t.Fatal(err)
	}
	entries := zipEntries(t, path)
	var meta domain.BackupMetadata
	if err := json.Unmarshal(entries["metadata.json"], &meta); err != nil {
		t.Fatal(err)
	}
	meta.AppVersion = "0.1.1"
	entries["metadata.json"], err = json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	legacyV2Entries(t, entries, "0.1.1")
	legacy := filepath.Join(t.TempDir(), "v011.zip")
	writeZip(t, legacy, entries)
	target, _, _, _ := setup(t)
	if _, err := target.Restore(ctx, legacy, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	q, err = target.Session(ctx, q.ID)
	if err != nil || q.ResumeIndex != 1 || q.Items[0].Progress != "selected" {
		t.Fatal(q, err)
	}
	hidden(t, q)
	if _, err := target.Complete(ctx, q.ID); err != nil {
		t.Fatal(err)
	}
	r, err := target.StartSessionReview(ctx, q.ID)
	if err != nil || len(r.Items) != 3 {
		t.Fatal(r, err)
	}
}
