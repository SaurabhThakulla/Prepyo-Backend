package fullmock

import (
	"context"

	"github.com/prepyo/backend/internal/listeningmock"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reading"
	"github.com/prepyo/backend/internal/speakingmock"
	"github.com/prepyo/backend/internal/writingmock"
)

// Starters deals and reopens each section's papers through its section mock.
// A full mock's papers are dealt fresh, not charged, and kept apart from any
// section mock the learner has open on its own.
func Starters(listening *listeningmock.Service, readingService *reading.Service,
	writing *writingmock.Service, speaking *speakingmock.Service) map[models.SkillType]Starter {
	return map[models.SkillType]Starter{
		models.SkillListening: section{
			deal: func(ctx context.Context, user models.User) (string, any, error) {
				paper, err := listening.Start(ctx, user, listeningmock.StartOptions{FullMock: true})
				return paper.ID, paper, err
			},
			resume: func(ctx context.Context, user models.User, id string) (any, error) {
				return listening.Resume(ctx, user, id)
			},
		},
		models.SkillReading: section{
			deal: func(ctx context.Context, user models.User) (string, any, error) {
				paper, err := readingService.StartMockForFullMock(ctx, user)
				return paper.ID, paper, err
			},
			resume: func(ctx context.Context, user models.User, id string) (any, error) {
				return readingService.ResumeMock(ctx, user, id)
			},
		},
		models.SkillWriting: section{
			deal: func(ctx context.Context, user models.User) (string, any, error) {
				paper, err := writing.StartForFullMock(ctx, user)
				return paper.ID, paper, err
			},
			resume: func(ctx context.Context, user models.User, id string) (any, error) {
				return writing.Resume(ctx, user, id)
			},
		},
		models.SkillSpeaking: section{
			deal: func(ctx context.Context, user models.User) (string, any, error) {
				paper, err := speaking.StartForFullMock(ctx, user)
				return paper.ID, paper, err
			},
			resume: func(ctx context.Context, user models.User, id string) (any, error) {
				return speaking.Resume(ctx, user, id)
			},
		},
	}
}

type section struct {
	deal   func(ctx context.Context, user models.User) (string, any, error)
	resume func(ctx context.Context, user models.User, id string) (any, error)
}

func (s section) Deal(ctx context.Context, user models.User) (string, any, error) {
	return s.deal(ctx, user)
}

func (s section) Resume(ctx context.Context, user models.User, id string) (any, error) {
	return s.resume(ctx, user, id)
}
