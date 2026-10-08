package repository

import (
	"context"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

type Lessons interface {
	ListLessons(context.Context) ([]domain.StoredLesson, error)
	Lesson(context.Context, string) (domain.StoredLesson, error)
	ImportLesson(context.Context, domain.StoredLesson) error
	LessonProgress(context.Context, string) ([]domain.LessonProgress, error)
	SaveLessonProgress(context.Context, domain.LessonProgress, bool) error
}
