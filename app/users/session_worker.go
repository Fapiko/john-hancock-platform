package users

import (
	"context"
	"time"

	"github.com/fapiko/john-hancock-platform/app/context/logger"
)

type SessionWorker struct {
	running        bool
	userRepository Repository
}

func NewSessionWorker(userRepository Repository) *SessionWorker {
	return &SessionWorker{
		running:        false,
		userRepository: userRepository,
	}
}

func (w *SessionWorker) Start(ctx context.Context) {
	log := logger.Get(ctx)
	w.running = true
	firstRun := true

	for w.running {
		if !firstRun {
			time.Sleep(time.Minute * 10)
		}
		firstRun = false

		numDeleted, err := w.userRepository.CleanupSessions(ctx)
		if err != nil {
			log.WithError(err).Error("Error cleaning up sessions")
			continue
		}

		if numDeleted > 0 {
			log.Infof("Cleaned up %d sessions", numDeleted)
		}
	}
}

func (w *SessionWorker) Stop(ctx context.Context) {
	w.running = false
}
