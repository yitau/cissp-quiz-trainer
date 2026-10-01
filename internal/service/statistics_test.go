package service_test

import (
	"context"
	"testing"
)

func TestReviewAndStatistics(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setup(t)
	load(t, s)
	empty, err := s.Statistics(ctx)
	if err != nil || empty.Total.Rate != nil || empty.Total.Count != 0 {
		t.Fatal(empty, err)
	}
	first, err := s.Start(ctx, "demo-v1", "study")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SubmitStudy(ctx, first.ID, "DEMO-001", "A"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	if err := s.SetFavorite(ctx, "DEMO-001", true); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 3; n++ {
		q, err := s.StartReview(ctx, "wrong")
		if err != nil {
			t.Fatal(err)
		}
		if len(q.Items) != 1 {
			t.Fatal("wrong review selection")
		}
		if _, err := s.SubmitStudy(ctx, q.ID, "DEMO-001", "B"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Complete(ctx, q.ID); err != nil {
			t.Fatal(err)
		}
	}
	wrong, err := s.Review(ctx, "wrong")
	if err != nil || len(wrong) != 1 || !wrong[0].Mastered || wrong[0].Wrong != 1 || wrong[0].Streak != 3 {
		t.Fatal("wrong history lost", wrong, err)
	}
	stats, err := s.Statistics(ctx)
	if err != nil || stats.Total.Count != 4 || stats.Total.Correct != 3 || *stats.Total.Rate != 75 || stats.WrongQuestions != 1 || stats.Favorites != 1 || stats.Domains[0].Count != 4 {
		t.Fatal(stats, err)
	}
	favorite, err := s.StartReview(ctx, "favorites")
	if err != nil || len(favorite.Items) != 1 {
		t.Fatal(favorite, err)
	}
	hidden(t, favorite)
	exam, err := s.Start(ctx, "demo-v1", "exam")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveChoice(ctx, exam.ID, "DEMO-001", "B", false); err != nil {
		t.Fatal(err)
	}
	stats, err = s.Statistics(ctx)
	if err != nil || stats.Total.Count != 4 {
		t.Fatal("draft included", stats, err)
	}
	result, err := s.Complete(ctx, exam.ID)
	if err != nil || result.Statistics.Total.Count != 4 || result.Statistics.Total.Unanswered != 3 {
		t.Fatal(result, err)
	}
	stats, err = s.Statistics(ctx)
	if err != nil || stats.Total.Count != 8 || stats.Total.Correct != 4 || stats.Total.Unanswered != 3 {
		t.Fatal(stats, err)
	}
	sum := 0
	for _, d := range stats.Domains {
		sum += d.Count
	}
	if sum != stats.Total.Count {
		t.Fatal("domain count mismatch")
	}
	sum = 0
	for _, d := range stats.Types {
		sum += d.Count
	}
	if sum != stats.Total.Count {
		t.Fatal("type count mismatch")
	}
	if err := s.SetFavorite(ctx, "DEMO-001", false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartReview(ctx, "favorites"); err == nil {
		t.Fatal("empty review accepted")
	}
}
