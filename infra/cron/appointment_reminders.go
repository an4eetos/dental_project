package cron

import (
	"context"
	"log/slog"
	"time"

	"go.uber.org/fx"

	"github.com/anuarkuanysh/dental_project/infra/config"
	appointmentuc "github.com/anuarkuanysh/dental_project/internal/usecase/appointment"
)

// RegisterAppointmentReminders starts a background ticker that sends appointment reminders.
func RegisterAppointmentReminders(
	lc fx.Lifecycle,
	reminders *appointmentuc.SendReminders,
	cfg config.Config,
	log *slog.Logger,
) {
	if reminders == nil {
		return
	}

	interval := cfg.ReminderPollInterval
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	var stopCh chan struct{}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			stopCh = make(chan struct{})
			go func() {
				ticker := time.NewTicker(interval)
				defer ticker.Stop()
				runAppointmentReminders(reminders, log)
				for {
					select {
					case <-stopCh:
						return
					case <-ticker.C:
						runAppointmentReminders(reminders, log)
					}
				}
			}()
			log.Info("appointment reminder worker started", "interval", interval.String())
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

func runAppointmentReminders(reminders *appointmentuc.SendReminders, log *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := reminders.Execute(ctx); err != nil {
		log.Error("appointment reminders failed", "err", err)
	}
}
