package cron

import (
	"context"
	"log/slog"
	"time"

	"go.uber.org/fx"

	"github.com/anuarkuanysh/dental_project/infra/config"
	appointmentuc "github.com/anuarkuanysh/dental_project/internal/usecase/appointment"
)

// RegisterAppointmentCleanup starts a background ticker that deletes past appointments.
func RegisterAppointmentCleanup(
	lc fx.Lifecycle,
	purge *appointmentuc.PurgeOutdated,
	cfg config.Config,
	log *slog.Logger,
) {
	if purge == nil {
		return
	}

	interval := cfg.AppointmentCleanupInterval
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
				runAppointmentCleanup(purge, log)
				for {
					select {
					case <-stopCh:
						return
					case <-ticker.C:
						runAppointmentCleanup(purge, log)
					}
				}
			}()
			log.Info("appointment cleanup worker started", "interval", interval.String())
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

func runAppointmentCleanup(purge *appointmentuc.PurgeOutdated, log *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	deleted, err := purge.Execute(ctx)
	if err != nil {
		log.Error("appointment cleanup failed", "err", err)
		return
	}
	if deleted > 0 {
		log.Info("appointment cleanup completed", "deleted", deleted)
	}
}
