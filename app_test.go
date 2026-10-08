package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestApplicationAdapterRestart(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	t.Setenv("CISSP_QUIZ_DATA_DIR", dir)
	a := NewApp()
	a.startup(ctx)
	if a.HealthCheck() != "ok" {
		t.Fatal(a.HealthCheck())
	}
	defer a.shutdown(ctx)
	p, err := a.trainer.PreviewFile(ctx, filepath.Join("samples", "demo-questions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := a.ConfirmImport(p.Token); err != nil {
		t.Fatal(err)
	}
	q, err := a.StartQuiz("demo-v1", "study")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.SubmitStudy(q.ID, "DEMO-001", "B"); err != nil {
		t.Fatal(err)
	}
	a.shutdown(ctx)
	a = NewApp()
	a.startup(ctx)
	defer a.shutdown(ctx)
	q, err = a.GetSession(q.ID)
	if err != nil || q.Items[0].Question.Answer != "B" || q.Items[1].Question.Answer != "" {
		t.Fatal(q, err)
	}
	if a.DataDirectory() != dir {
		t.Fatal("wrong data directory")
	}
	if _, err := os.Stat(filepath.Join(dir, "cissp.db")); err != nil {
		t.Fatal(err)
	}
	if _, err := a.GetSession("missing"); err == nil || err.Error() != "记录不存在，请刷新后重试" {
		t.Fatal("unhelpful missing record error", err)
	}
}
