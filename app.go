package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yitau/cissp-quiz-trainer/internal/database"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/repository/sqlite"
	"github.com/yitau/cissp-quiz-trainer/internal/service"
	driver "modernc.org/sqlite"
)

// App is the thin Wails-facing application adapter.
type App struct {
	ctx     context.Context
	db      *sql.DB
	trainer *service.Trainer
	initErr error
	dataDir string
	logFile *os.File
	logger  *slog.Logger
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
	a.logFile, _ = os.OpenFile(filepath.Join(a.dataDir, "app.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if a.logFile != nil {
		a.logger = slog.New(slog.NewTextHandler(a.logFile, nil))
	}
	if a.initErr != nil {
		a.initErr = a.report(a.initErr)
	}
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
	if a.logFile != nil {
		a.logFile.Close()
	}
}
func (a *App) report(err error) error {
	if err == nil {
		return nil
	}
	if a.logger != nil {
		a.logger.Error("operation failed", "error", err)
	}
	var databaseError *driver.Error
	if errors.As(err, &databaseError) {
		return fmt.Errorf("本地数据库操作失败，请重试；详细原因见数据目录 app.log")
	}
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("记录不存在，请刷新后重试")
	}
	return err
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
	value, err := a.trainer.PreviewFile(a.ctx, path)
	return value, a.report(err)
}
func (a *App) ConfirmImport(token string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.report(a.trainer.Import(a.ctx, token))
}
func (a *App) ListSets() ([]domain.SetSummary, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	value, err := a.trainer.ListSets(a.ctx)
	return value, a.report(err)
}
func (a *App) SetQuestions(id string) ([]domain.Question, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	value, err := a.trainer.SetQuestions(a.ctx, id)
	return value, a.report(err)
}
func (a *App) StartQuiz(id, mode string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	value, err := a.trainer.Start(a.ctx, id, mode)
	return value, a.report(err)
}
func (a *App) GetSession(id string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	value, err := a.trainer.Session(a.ctx, id)
	return value, a.report(err)
}
func (a *App) SubmitStudy(id, qid, selected string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	value, err := a.trainer.SubmitStudy(a.ctx, id, qid, selected)
	return value, a.report(err)
}
func (a *App) CompleteQuiz(id string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	value, err := a.trainer.Complete(a.ctx, id)
	return value, a.report(err)
}
func (a *App) ListSessions() ([]domain.SessionSummary, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	value, err := a.trainer.ListSessions(a.ctx)
	return value, a.report(err)
}
func (a *App) SaveChoice(id, qid, selected string, flagged bool) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	value, err := a.trainer.SaveChoice(a.ctx, id, qid, selected, flagged)
	return value, a.report(err)
}

func (a *App) Statistics() (domain.Statistics, error) {
	if err := a.ready(); err != nil {
		return domain.Statistics{}, err
	}
	value, err := a.trainer.Statistics(a.ctx)
	return value, a.report(err)
}
func (a *App) Review(kind string) ([]domain.ReviewQuestion, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	value, err := a.trainer.Review(a.ctx, kind)
	return value, a.report(err)
}
func (a *App) SetFavorite(id string, value bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.report(a.trainer.SetFavorite(a.ctx, id, value))
}
func (a *App) StartReview(kind string) (domain.Session, error) {
	if err := a.ready(); err != nil {
		return domain.Session{}, err
	}
	value, err := a.trainer.StartReview(a.ctx, kind)
	return value, a.report(err)
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
		return "", a.report(err)
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
	choice, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{Type: runtime.QuestionDialog, Title: "确认恢复备份", Message: "恢复将替换当前全部题库、学习记录和收藏。校验通过后会先保留当前数据的安全备份。\n\n文件：" + path, Buttons: []string{"Yes", "No"}, DefaultButton: "No", CancelButton: "No"})
	if err != nil || choice != "Yes" {
		return "", err
	}
	value, err := a.trainer.Restore(a.ctx, path, filepath.Join(a.dataDir, "backups"))
	return value, a.report(err)
}
