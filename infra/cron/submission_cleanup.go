package cron

import (
	"context"
	"log/slog"
	"time"

	"go.uber.org/fx"

	"github.com/anuarkuanysh/dental_project/infra/config"
	photoreviewuc "github.com/anuarkuanysh/dental_project/internal/usecase/photo_review"
)

// RegisterSubmissionCleanup starts a background ticker that deletes stale pending submissions.
func RegisterSubmissionCleanup(
	lc fx.Lifecycle,
	purge *photoreviewuc.PurgeStale,
	cfg config.Config,
	log *slog.Logger,
) {
	if purge == nil {
		return
	}

	interval := cfg.SubmissionCleanupInterval
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	var stopCh chan struct{}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			stopCh = make(chan struct{})
			go func() {
				ticker := time.NewTicker(interval)
				defer ticker.Stop()
				runSubmissionCleanup(purge, log)
				for {
					select {
					case <-stopCh:
						return
					case <-ticker.C:
						runSubmissionCleanup(purge, log)
					}
				}
			}()
			log.Info("submission cleanup worker started", "interval", interval.String())
			return nil
		},
		OnStop: func(context.Context) error {
			if stopCh != nil {
				close(stopCh)
			}
			return nil
		},
	})
}

func runSubmissionCleanup(purge *photoreviewuc.PurgeStale, log *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	deleted, err := purge.Execute(ctx)
	if err != nil {
		log.Error("submission cleanup failed", "err", err)
		return
	}
	if deleted > 0 {
		log.Info("submission cleanup completed", "deleted", deleted)
	}
}
