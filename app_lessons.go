package main

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

func (a *App) ChooseLessonImport() (domain.LessonPreview, error) {
	if err := a.ready(); err != nil {
		return domain.LessonPreview{}, err
	}
	a.trainer.CancelLessonImport()
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "选择 Lesson JSON 课程", Filters: []runtime.FileFilter{{DisplayName: "JSON 课程", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return domain.LessonPreview{}, err
	}
	value, err := a.trainer.PreviewLessonFile(a.ctx, path)
	return value, a.report(err)
}
func (a *App) ConfirmLessonImport(token string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.report(a.trainer.ImportLesson(a.ctx, token))
}
func (a *App) CancelLessonImport() error {
	if err := a.ready(); err != nil {
		return err
	}
	a.trainer.CancelLessonImport()
	return nil
}
func (a *App) ListLessons() ([]domain.LessonSummary, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	value, err := a.trainer.ListLessons(a.ctx)
	return value, a.report(err)
}
func (a *App) GetLesson(id string) (domain.LessonView, error) {
	if err := a.ready(); err != nil {
		return domain.LessonView{}, err
	}
	value, err := a.trainer.GetLesson(a.ctx, id)
	return value, a.report(err)
}
func (a *App) OpenConcept(id, conceptID string) (domain.LessonView, error) {
	if err := a.ready(); err != nil {
		return domain.LessonView{}, err
	}
	value, err := a.trainer.OpenConcept(a.ctx, id, conceptID)
	return value, a.report(err)
}
func (a *App) SetConceptStatus(id, conceptID, status string) (domain.LessonView, error) {
	if err := a.ready(); err != nil {
		return domain.LessonView{}, err
	}
	value, err := a.trainer.SetConceptStatus(a.ctx, id, conceptID, status)
	return value, a.report(err)
}
