package repository

import (
	"context"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

type Trainer interface {
	LearningData(context.Context) ([]domain.Question, []domain.Attempt, map[string]bool, error)
	SetFavorite(context.Context, string, bool) error
	ListSets(context.Context) ([]domain.SetSummary, error)
	ImportCheck(context.Context, domain.QuestionFile) (string, []string, error)
	Import(context.Context, domain.QuestionFile) error
	SetQuestions(context.Context, string) ([]domain.Question, error)
	CreateSession(context.Context, domain.Session) error
	Session(context.Context, string) (domain.Session, error)
	SaveItem(context.Context, string, domain.SessionItem) error
	Complete(context.Context, domain.Session) error
	ListSessions(context.Context) ([]domain.SessionSummary, error)
}
