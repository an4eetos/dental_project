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

	"github.com/anuarkuanysh/dental_project/infra/admin"
	"github.com/anuarkuanysh/dental_project/infra/config"
	"github.com/anuarkuanysh/dental_project/infra/doctor"
	infrahttp "github.com/anuarkuanysh/dental_project/infra/http"
	"github.com/anuarkuanysh/dental_project/infra/postgresql"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driven/jwt"
	postgresadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/postgres"
	telegramadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/telegram"
	adminhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/admin"
	appointmenthandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/appointment"
	authhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/auth"
	photoreviewhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/photo_review"
	predictionhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/prediction"
	exceladapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/excel"
	"github.com/anuarkuanysh/dental_project/internal/gemini"
	"github.com/anuarkuanysh/dental_project/internal/imageproc"
	"github.com/anuarkuanysh/dental_project/internal/port"
	"github.com/anuarkuanysh/dental_project/internal/telegrambot"
	authuc "github.com/anuarkuanysh/dental_project/internal/usecase/auth"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
	appointmentuc "github.com/anuarkuanysh/dental_project/internal/usecase/appointment"
	photoreviewuc "github.com/anuarkuanysh/dental_project/internal/usecase/photo_review"
	predictionuc "github.com/anuarkuanysh/dental_project/internal/usecase/prediction"
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
			providePhotoSubmissionRepo,
			provideStatsRepo,
			provideInitDataValidator,
			provideDoctorRegistry,
			provideAdminRegistry,
			provideTokenIssuer,
			provideClock,
			provideImageProcessor,
			provideTelegramBotClient,
			provideVisionAnalyzer,
			provideTelegramLogin,
			provideCreateAppointment,
			provideListAppointments,
			provideListForDoctor,
			provideSubmitPhoto,
			provideListPendingSubmissions,
			provideListAnsweredSubmissions,
			provideGetSubmission,
			provideGetSubmissionImage,
			provideGenerateDraft,
			provideRespondSubmission,
			provideAdminStatistics,
			providePredictionExamples,
			provideTextGenerator,
			providePredict,
			newAuthHandler,
			newAppointmentHandler,
			newPhotoReviewHandler,
			newAdminHandler,
			newPredictionHandler,
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

func providePhotoSubmissionRepo(pool *pgxpool.Pool) port.PhotoSubmissionRepository {
	return postgresadapter.NewPhotoSubmissionRepository(pool)
}

func provideStatsRepo(pool *pgxpool.Pool) port.StatsRepository {
	return postgresadapter.NewStatsRepository(pool)
}

func provideInitDataValidator(cfg config.Config) port.TelegramInitDataValidator {
	return telegramadapter.NewInitDataValidator(cfg.TelegramBotToken, cfg.TelegramAuthMaxAge)
}

func provideDoctorRegistry(cfg config.Config) port.DoctorRegistry {
	return doctor.NewTelegramIDRegistry(cfg.DoctorTelegramIDs)
}

func provideAdminRegistry(cfg config.Config) port.AdminRegistry {
	return admin.NewTelegramIDRegistry(cfg.AdminTelegramIDs)
}

func provideImageProcessor(cfg config.Config) port.ImageProcessor {
	return imageproc.New(cfg.MaxImageDimension)
}

func provideTelegramBotClient(bot *tgbotapi.BotAPI, tgHTTP infrahttp.TelegramHTTPClient) *telegrambot.Client {
	return telegrambot.New(bot, tgHTTP.Client)
}

func provideVisionAnalyzer(
	cfg config.Config,
	geminiHTTP infrahttp.GeminiHTTPClient,
	log *slog.Logger,
) port.VisionAnalyzer {
	return gemini.New(geminiHTTP.Client, cfg.GeminiAPIKey, cfg.GeminiModel, log)
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
	admins port.AdminRegistry,
) *authuc.TelegramLogin {
	return &authuc.TelegramLogin{
		Validator: v,
		Users:     users,
		Tokens:    tokens,
		Doctors:   doctors,
		Admins:    admins,
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

func provideSubmitPhoto(
	users port.UserRepository,
	submissions port.PhotoSubmissionRepository,
	images port.ImageProcessor,
	tg *telegrambot.Client,
	doctors port.DoctorRegistry,
	admins port.AdminRegistry,
) *photoreviewuc.SubmitFromTelegram {
	return &photoreviewuc.SubmitFromTelegram{
		Users:       users,
		Submissions: submissions,
		Images:      images,
		Downloader:  tg,
		Sender:      tg,
		Doctors:     doctors,
		Admins:      admins,
	}
}

func provideListPendingSubmissions(
	submissions port.PhotoSubmissionRepository,
	users port.UserRepository,
) *photoreviewuc.ListPending {
	return &photoreviewuc.ListPending{Submissions: submissions, Users: users}
}

func provideListAnsweredSubmissions(
	submissions port.PhotoSubmissionRepository,
	users port.UserRepository,
) *photoreviewuc.ListAnswered {
	return &photoreviewuc.ListAnswered{Submissions: submissions, Users: users}
}

func provideGetSubmission(
	submissions port.PhotoSubmissionRepository,
	users port.UserRepository,
) *photoreviewuc.Get {
	return &photoreviewuc.Get{Submissions: submissions, Users: users}
}

func provideGetSubmissionImage(
	submissions port.PhotoSubmissionRepository,
	users port.UserRepository,
) *photoreviewuc.GetImage {
	return &photoreviewuc.GetImage{Submissions: submissions, Users: users}
}

func provideGenerateDraft(
	submissions port.PhotoSubmissionRepository,
	users port.UserRepository,
	analyzer port.VisionAnalyzer,
) *photoreviewuc.GenerateDraft {
	return &photoreviewuc.GenerateDraft{Submissions: submissions, Users: users, Analyzer: analyzer}
}

func provideRespondSubmission(
	submissions port.PhotoSubmissionRepository,
	users port.UserRepository,
	tg *telegrambot.Client,
	clock port.Clock,
) *photoreviewuc.Respond {
	return &photoreviewuc.Respond{Submissions: submissions, Users: users, Sender: tg, Clock: clock}
}

func provideAdminStatistics(
	stats port.StatsRepository,
	users port.UserRepository,
) *adminuc.Statistics {
	return &adminuc.Statistics{Stats: stats, Users: users}
}

func providePredictionExamples(cfg config.Config, log *slog.Logger) (port.PredictionExampleRepository, error) {
	repo, err := exceladapter.NewRepository(cfg.PredictionExamplesPath)
	if err != nil {
		return nil, err
	}
	examples, err := repo.ListExamples(context.Background())
	if err != nil {
		return nil, err
	}
	log.Info("prediction examples loaded", "path", cfg.PredictionExamplesPath, "count", len(examples))
	return repo, nil
}

func provideTextGenerator(
	cfg config.Config,
	geminiHTTP infrahttp.GeminiHTTPClient,
	log *slog.Logger,
) port.TextGenerator {
	return gemini.New(geminiHTTP.Client, cfg.GeminiAPIKey, cfg.GeminiModel, log)
}

func providePredict(
	examples port.PredictionExampleRepository,
	generator port.TextGenerator,
) *predictionuc.Predict {
	return &predictionuc.Predict{Examples: examples, Generator: generator}
}

func newPredictionHandler(predict *predictionuc.Predict) *predictionhandler.Handler {
	return &predictionhandler.Handler{PredictUC: predict}
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

func newPhotoReviewHandler(
	listPending *photoreviewuc.ListPending,
	listAnswered *photoreviewuc.ListAnswered,
	get *photoreviewuc.Get,
	getImage *photoreviewuc.GetImage,
	generateDraft *photoreviewuc.GenerateDraft,
	respond *photoreviewuc.Respond,
) *photoreviewhandler.Handler {
	return &photoreviewhandler.Handler{
		ListPendingUC:   listPending,
		ListAnsweredUC:  listAnswered,
		GetUC:           get,
		GetImageUC:      getImage,
		GenerateDraftUC: generateDraft,
		RespondUC:       respond,
	}
}

func newAdminHandler(statistics *adminuc.Statistics) *adminhandler.Handler {
	return &adminhandler.Handler{StatisticsUC: statistics}
}

func newGinEngine(
	cfg config.Config,
	log *slog.Logger,
	auth *authhandler.Handler,
	appt *appointmenthandler.Handler,
	photoReview *photoreviewhandler.Handler,
	adminH *adminhandler.Handler,
	predict *predictionhandler.Handler,
	submitPhoto *photoreviewuc.SubmitFromTelegram,
	tokens port.TokenIssuer,
	bot *tgbotapi.BotAPI,
	tgHTTP infrahttp.TelegramHTTPClient,
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
		PhotoReviewH: photoReview,
		AdminH:       adminH,
		PredictionH:  predict,
		SubmitPhoto:  submitPhoto,
		Tokens:       tokens,
		Bot:          bot,
		TelegramHTTP: tgHTTP,
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
