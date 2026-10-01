package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
	"os"
	"path/filepath"
	"time"
)

// App is the thin Wails-facing application adapter.
type App struct {
	ctx     context.Context
	db      *sql.DB
	trainer *service.Trainer
	initErr error
	dataDir string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.dataDir = os.Getenv("CISSP_QUIZ_DATA_DIR")
	if a.dataDir == "" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			a.initErr = fmt.Errorf("无法找到 LOCALAPPDATA，请设置 CISSP_QUIZ_DATA_DIR")
			return
		}
		a.dataDir = filepath.Join(base, "CISSPQuizTrainer", "data")
	}
	a.db, a.initErr = database.Open(filepath.Join(a.dataDir, "cissp.db"))
	if a.initErr == nil {
		a.trainer = service.New(sqlite.New(a.db))
	}
}

// HealthCheck proves the initial Wails binding surface is available.
func (a *App) HealthCheck() string {
	if a.initErr != nil {
		return a.initErr.Error()
	}
	return "ok"
}
func (a *App) shutdown(context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}
func (a *App) beforeClose(context.Context) bool { return a.trainer != nil && a.trainer.Busy() }
func (a *App) ready() error {
	if a.initErr != nil {
		return fmt.Errorf("数据库未就绪：%w", a.initErr)
	}
	if a.trainer == nil {
		return fmt.Errorf("应用尚未启动")
	}
	return nil
}
func (a *App) DataDirectory() string { return a.dataDir }
func (a *App) ChooseImport() (domain.Preview, error) {
	if err := a.ready(); err != nil {
		return domain.Preview{}, err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "选择 JSON 题集", Filters: []runtime.FileFilter{{DisplayName: "JSON 题集", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return domain.Preview{}, err
	}
	return a.trainer.PreviewFile(a.ctx, path)
}
func (a *App) ConfirmImport(token string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.trainer.Import(a.ctx, token)
}
func (a *App) ListSets() ([]domain.SetSummary, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.trainer.ListSets(a.ctx)
}
func (a *App) SetQuestions(id string) ([]domain.Question, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.trainer.SetQuestions(a.ctx, id)
}
func (a *App) StartQuiz(id, mode string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	return a.trainer.Start(a.ctx, id, mode)
}
func (a *App) GetSession(id string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	return a.trainer.Session(a.ctx, id)
}
func (a *App) SubmitStudy(id, qid, selected string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	return a.trainer.SubmitStudy(a.ctx, id, qid, selected)
}
func (a *App) CompleteQuiz(id string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	return a.trainer.Complete(a.ctx, id)
}
func (a *App) ListSessions() ([]domain.SessionSummary, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.trainer.ListSessions(a.ctx)
}
func (a *App) SaveChoice(id, qid, selected string, flagged bool) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	return a.trainer.SaveChoice(a.ctx, id, qid, selected, flagged)
}

func (a *App) Statistics() (domain.Statistics, error) {
	if err := a.ready(); err != nil {
		return domain.Statistics{}, err
	}
	return a.trainer.Statistics(a.ctx)
}
func (a *App) Review(kind string) ([]domain.ReviewQuestion, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.trainer.Review(a.ctx, kind)
}
func (a *App) SetFavorite(id string, value bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.trainer.SetFavorite(a.ctx, id, value)
}
func (a *App) StartReview(kind string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	return a.trainer.StartReview(a.ctx, kind)
}
func (a *App) CreateBackup() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: "保存完整备份（请选择新文件名）", DefaultFilename: "cissp-backup-" + time.Now().Format("20060102-150405") + ".zip", Filters: []runtime.FileFilter{{DisplayName: "完整备份 ZIP", Pattern: "*.zip"}}})
	if err != nil || path == "" {
		return "", err
	}
	if err := a.trainer.Backup(a.ctx, path); err != nil {
		return "", err
	}
	return path, nil
}
func (a *App) RestoreBackup() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "选择需要恢复的完整备份", Filters: []runtime.FileFilter{{DisplayName: "完整备份 ZIP", Pattern: "*.zip"}}})
	if err != nil || path == "" {
		return "", err
	}
	choice, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{Type: runtime.QuestionDialog, Title: "确认恢复备份", Message: "恢复将替换当前全部题库、学习记录和收藏。校验通过后会先保留当前数据的安全备份。\n\n文件：" + path, Buttons: []string{"恢复", "取消"}, DefaultButton: "取消", CancelButton: "取消"})
	if err != nil || choice != "恢复" {
		return "", err
	}
	return a.trainer.Restore(a.ctx, path, filepath.Join(a.dataDir, "backups"))
}
