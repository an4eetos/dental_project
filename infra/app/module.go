package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/anuarkuanysh/dental_project/infra/config"
	"github.com/anuarkuanysh/dental_project/infra/doctor"
	infrahttp "github.com/anuarkuanysh/dental_project/infra/http"
	"github.com/anuarkuanysh/dental_project/infra/postgresql"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driven/jwt"
	postgresadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/postgres"
	telegramadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/telegram"
	appointmenthandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/appointment"
	authhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/auth"
	"github.com/anuarkuanysh/dental_project/internal/port"
	authuc "github.com/anuarkuanysh/dental_project/internal/usecase/auth"
	appointmentuc "github.com/anuarkuanysh/dental_project/internal/usecase/appointment"
	pkgclock "github.com/anuarkuanysh/dental_project/pkg/clock"
)

// Module is the root Fx module for the teeth-bot service.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			loadConfig,
			newLogger,
			newBot,
			newTelegramHTTPClient,
			newGeminiHTTPClient,
			newPool,
			provideUserRepo,
			provideAppointmentRepo,
			provideInitDataValidator,
			provideDoctorRegistry,
			provideTokenIssuer,
			provideClock,
			provideTelegramLogin,
			provideCreateAppointment,
			provideListAppointments,
			provideListForDoctor,
			newAuthHandler,
			newAppointmentHandler,
			newGinEngine,
			newHTTPServer,
		),
		fx.Invoke(runMigrations, startHTTPServer),
	)
}

func loadConfig() (config.Config, error) {
	return config.Load()
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(h)
}

func newBot(cfg config.Config, log *slog.Logger) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}
	bot.Debug = false
	log.Info("authorized bot", "username", bot.Self.UserName)
	return bot, nil
}

func newTelegramHTTPClient(cfg config.Config) infrahttp.TelegramHTTPClient {
	return infrahttp.NewTelegramHTTPClient(cfg.TelegramDownloadTimeout)
}

func newGeminiHTTPClient(cfg config.Config) infrahttp.GeminiHTTPClient {
	return infrahttp.NewGeminiHTTPClient(cfg.GeminiHTTPTimeout)
}

func newPool(lc fx.Lifecycle, cfg config.Config) (*pgxpool.Pool, error) {
	ctx := context.Background()
	pool, err := postgresql.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool, nil
}

func runMigrations(pool *pgxpool.Pool) error {
	return postgresql.RunMigrations(context.Background(), pool)
}

func provideUserRepo(pool *pgxpool.Pool) port.UserRepository {
	return postgresadapter.NewUserRepository(pool)
}

func provideAppointmentRepo(pool *pgxpool.Pool) port.AppointmentRepository {
	return postgresadapter.NewAppointmentRepository(pool)
}

func provideInitDataValidator(cfg config.Config) port.TelegramInitDataValidator {
	return telegramadapter.NewInitDataValidator(cfg.TelegramBotToken, cfg.TelegramAuthMaxAge)
}

func provideDoctorRegistry(cfg config.Config) port.DoctorRegistry {
	return doctor.NewTelegramIDRegistry(cfg.DoctorTelegramIDs)
}

func provideTokenIssuer(cfg config.Config) port.TokenIssuer {
	return jwt.NewTokenIssuer(cfg.JWTSecret, cfg.JWTTTL)
}

func provideClock() port.Clock {
	return pkgclock.System{}
}

func provideTelegramLogin(
	v port.TelegramInitDataValidator,
	users port.UserRepository,
	tokens port.TokenIssuer,
	doctors port.DoctorRegistry,
) *authuc.TelegramLogin {
	return &authuc.TelegramLogin{
		Validator: v,
		Users:     users,
		Tokens:    tokens,
		Doctors:   doctors,
	}
}

func provideCreateAppointment(
	repo port.AppointmentRepository,
	users port.UserRepository,
	clock port.Clock,
) *appointmentuc.Create {
	return &appointmentuc.Create{Appointments: repo, Users: users, Clock: clock}
}

func provideListAppointments(repo port.AppointmentRepository) *appointmentuc.ListMine {
	return &appointmentuc.ListMine{Appointments: repo}
}

func provideListForDoctor(
	repo port.AppointmentRepository,
	users port.UserRepository,
) *appointmentuc.ListForDoctor {
	return &appointmentuc.ListForDoctor{Appointments: repo, Users: users}
}

func newAuthHandler(login *authuc.TelegramLogin, log *slog.Logger) *authhandler.Handler {
	return &authhandler.Handler{Login: login, Log: log}
}

func newAppointmentHandler(
	create *appointmentuc.Create,
	list *appointmentuc.ListMine,
	listDoctor *appointmentuc.ListForDoctor,
) *appointmenthandler.Handler {
	return &appointmenthandler.Handler{
		CreateUC:        create,
		ListMineUC:      list,
		ListForDoctorUC: listDoctor,
	}
}

func newGinEngine(
	cfg config.Config,
	log *slog.Logger,
	auth *authhandler.Handler,
	appt *appointmenthandler.Handler,
	tokens port.TokenIssuer,
	bot *tgbotapi.BotAPI,
	tgHTTP infrahttp.TelegramHTTPClient,
	geminiHTTP infrahttp.GeminiHTTPClient,
) *gin.Engine {
	if len(cfg.CORSAllowOrigins) == 0 {
		log.Warn("cors: CORS_ALLOW_ORIGINS is empty — all origins allowed (ok for dev)")
	} else {
		log.Info("cors: allowed origins", "origins", cfg.CORSAllowOrigins)
	}
	return infrahttp.NewRouter(infrahttp.RouterParams{
		Config:       cfg,
		Log:          log,
		AuthHandler:  auth,
		AppointmentH: appt,
		Tokens:       tokens,
		Bot:          bot,
		TelegramHTTP: tgHTTP,
		GeminiHTTP:   geminiHTTP,
	})
}

func newHTTPServer(cfg config.Config, engine *gin.Engine) *http.Server {
	return infrahttp.NewHTTPServer(cfg.HTTPAddr, engine, cfg.RequestTimeout)
}

func startHTTPServer(lc fx.Lifecycle, srv *http.Server, log *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("listening", "addr", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("server", "err", err)
					os.Exit(1)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			return srv.Shutdown(stopCtx)
		},
	})
}
