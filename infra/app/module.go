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
	"github.com/anuarkuanysh/dental_project/infra/cron"
	"github.com/anuarkuanysh/dental_project/infra/doctor"
	infrahttp "github.com/anuarkuanysh/dental_project/infra/http"
	"github.com/anuarkuanysh/dental_project/infra/postgresql"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driven/jwt"
	postgresadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/postgres"
	subscriptionadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/subscription"
	telegramadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/telegram"
	adminhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/admin"
	appointmenthandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/appointment"
	authhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/auth"
	photoreviewhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/photo_review"
	predictionhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/prediction"
	subscriptionhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/subscription"
	"github.com/anuarkuanysh/dental_project/internal/gemini"
	"github.com/anuarkuanysh/dental_project/internal/imageproc"
	"github.com/anuarkuanysh/dental_project/internal/port"
	"github.com/anuarkuanysh/dental_project/internal/telegrambot"
	authuc "github.com/anuarkuanysh/dental_project/internal/usecase/auth"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
	appointmentuc "github.com/anuarkuanysh/dental_project/internal/usecase/appointment"
	photoreviewuc "github.com/anuarkuanysh/dental_project/internal/usecase/photo_review"
	predictionuc "github.com/anuarkuanysh/dental_project/internal/usecase/prediction"
	subscriptionuc "github.com/anuarkuanysh/dental_project/internal/usecase/subscription"
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
			provideSubscriptionRepo,
			provideInvoicePayloadSigner,
			provideTelegramPayments,
			provideSubscriptionPlan,
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
			provideRespondAppointment,
			provideSetZoomLink,
			provideSuggestSlots,
			provideSendAppointmentReminders,
			providePurgeOutdatedAppointments,
			provideAppointmentLocation,
			provideSubmitPhoto,
			provideListPendingSubmissions,
			provideListAnsweredSubmissions,
			provideGetSubmission,
			provideGetSubmissionImage,
			provideGenerateDraft,
			provideRespondSubmission,
			provideAdminStatistics,
			provideAdminListUsers,
			provideAdminGetUser,
			provideAdminUpdateUser,
			provideAdminSetBlocked,
			providePredictionExamples,
			provideTextGenerator,
			providePredict,
			provideGetSubscriptionStatus,
			provideCreateSubscriptionInvoice,
			provideAnswerPreCheckout,
			provideConfirmPayment,
			provideSubscriptionChecker,
			newAuthHandler,
			newAppointmentHandler,
			newPhotoReviewHandler,
			newAdminHandler,
			newPredictionHandler,
			newSubscriptionHandler,
			newGinEngine,
			newHTTPServer,
		),
		fx.Invoke(runMigrations, startHTTPServer, cron.RegisterAppointmentReminders, cron.RegisterAppointmentCleanup),
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

func provideSubscriptionRepo(pool *pgxpool.Pool) port.SubscriptionRepository {
	return postgresadapter.NewSubscriptionRepository(pool)
}

func provideInvoicePayloadSigner(cfg config.Config) port.InvoicePayloadSigner {
	return subscriptionadapter.NewInvoicePayloadSigner(cfg.SubscriptionInvoiceSecret)
}

func provideTelegramPayments(cfg config.Config, tgHTTP infrahttp.TelegramHTTPClient) port.TelegramPayments {
	return telegramadapter.NewPaymentsClient(cfg.TelegramBotToken, tgHTTP.Client)
}

func provideSubscriptionPlan(cfg config.Config) subscriptionuc.PlanConfig {
	return subscriptionuc.PlanConfig{
		StarsPrice:   cfg.SubscriptionStarsPrice,
		Duration:     cfg.SubscriptionDuration,
		InvoiceTitle: cfg.SubscriptionInvoiceTitle,
		InvoiceDesc:  cfg.SubscriptionInvoiceDesc,
	}
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
	doctors port.DoctorRegistry,
	tg *telegrambot.Client,
	clock port.Clock,
) *appointmentuc.Create {
	return &appointmentuc.Create{
		Appointments: repo,
		Users:        users,
		Doctors:      doctors,
		Sender:       tg,
		Clock:        clock,
	}
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

func provideRespondAppointment(
	repo port.AppointmentRepository,
	users port.UserRepository,
	tg *telegrambot.Client,
	clock port.Clock,
) *appointmentuc.Respond {
	return &appointmentuc.Respond{
		Appointments: repo,
		Users:        users,
		Sender:       tg,
		Clock:        clock,
	}
}

func provideSetZoomLink(
	repo port.AppointmentRepository,
	users port.UserRepository,
	tg *telegrambot.Client,
) *appointmentuc.SetZoomLink {
	return &appointmentuc.SetZoomLink{
		Appointments: repo,
		Users:        users,
		Sender:       tg,
	}
}

func provideSuggestSlots(
	repo port.AppointmentRepository,
	users port.UserRepository,
	clock port.Clock,
) *appointmentuc.SuggestSlots {
	return &appointmentuc.SuggestSlots{
		Appointments: repo,
		Users:        users,
		Clock:        clock,
	}
}

func provideSendAppointmentReminders(
	repo port.AppointmentRepository,
	users port.UserRepository,
	tg *telegrambot.Client,
	clock port.Clock,
	loc *time.Location,
) *appointmentuc.SendReminders {
	return &appointmentuc.SendReminders{
		Appointments: repo,
		Users:        users,
		Sender:       tg,
		Clock:        clock,
		Location:     loc,
	}
}

func providePurgeOutdatedAppointments(
	repo port.AppointmentRepository,
	clock port.Clock,
	loc *time.Location,
) *appointmentuc.PurgeOutdated {
	return &appointmentuc.PurgeOutdated{
		Appointments: repo,
		Clock:        clock,
		Location:     loc,
	}
}

func provideAppointmentLocation(cfg config.Config, log *slog.Logger) (*time.Location, error) {
	loc, err := time.LoadLocation(cfg.AppointmentTimezone)
	if err != nil {
		log.Warn("invalid appointment timezone, falling back to UTC", "timezone", cfg.AppointmentTimezone, "err", err)
		return time.UTC, nil
	}
	return loc, nil
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

func provideAdminListUsers(users port.UserRepository) *adminuc.ListUsers {
	return &adminuc.ListUsers{Users: users}
}

func provideAdminGetUser(users port.UserRepository) *adminuc.GetUser {
	return &adminuc.GetUser{Users: users}
}

func provideAdminUpdateUser(users port.UserRepository) *adminuc.UpdateUser {
	return &adminuc.UpdateUser{Users: users}
}

func provideAdminSetBlocked(users port.UserRepository) *adminuc.SetBlocked {
	return &adminuc.SetBlocked{Users: users}
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

func provideGetSubscriptionStatus(
	repo port.SubscriptionRepository,
	clock port.Clock,
	plan subscriptionuc.PlanConfig,
) *subscriptionuc.GetStatus {
	return &subscriptionuc.GetStatus{Subscriptions: repo, Clock: clock, Plan: plan}
}

func provideCreateSubscriptionInvoice(
	users port.UserRepository,
	payments port.TelegramPayments,
	signer port.InvoicePayloadSigner,
	plan subscriptionuc.PlanConfig,
) *subscriptionuc.CreateInvoice {
	return &subscriptionuc.CreateInvoice{
		Users:    users,
		Payments: payments,
		Signer:   signer,
		Plan:     plan,
	}
}

func provideAnswerPreCheckout(
	users port.UserRepository,
	payments port.TelegramPayments,
	signer port.InvoicePayloadSigner,
) *subscriptionuc.AnswerPreCheckout {
	return &subscriptionuc.AnswerPreCheckout{
		Users:    users,
		Payments: payments,
		Signer:   signer,
	}
}

func provideConfirmPayment(
	repo port.SubscriptionRepository,
	signer port.InvoicePayloadSigner,
	clock port.Clock,
	plan subscriptionuc.PlanConfig,
) *subscriptionuc.ConfirmPayment {
	return &subscriptionuc.ConfirmPayment{
		Subscriptions: repo,
		Signer:        signer,
		Clock:         clock,
		Plan:          plan,
	}
}

func provideSubscriptionChecker(
	getStatus *subscriptionuc.GetStatus,
	users port.UserRepository,
) port.SubscriptionChecker {
	return &subscriptionuc.Checker{GetStatus: getStatus, Users: users}
}

func newPredictionHandler(predict *predictionuc.Predict) *predictionhandler.Handler {
	return &predictionhandler.Handler{PredictUC: predict}
}

func newSubscriptionHandler(
	getStatus *subscriptionuc.GetStatus,
	createInvoice *subscriptionuc.CreateInvoice,
	users port.UserRepository,
	log *slog.Logger,
) *subscriptionhandler.Handler {
	return &subscriptionhandler.Handler{
		GetStatus:       getStatus,
		CreateInvoiceUC: createInvoice,
		Users:           users,
		Log:             log,
	}
}

func newAuthHandler(login *authuc.TelegramLogin, getStatus *subscriptionuc.GetStatus, log *slog.Logger) *authhandler.Handler {
	return &authhandler.Handler{Login: login, GetStatus: getStatus, Log: log}
}

func newAppointmentHandler(
	create *appointmentuc.Create,
	list *appointmentuc.ListMine,
	listDoctor *appointmentuc.ListForDoctor,
	respond *appointmentuc.Respond,
	setZoomLink *appointmentuc.SetZoomLink,
	suggestSlots *appointmentuc.SuggestSlots,
) *appointmenthandler.Handler {
	return &appointmenthandler.Handler{
		CreateUC:        create,
		ListMineUC:      list,
		ListForDoctorUC: listDoctor,
		RespondUC:       respond,
		SetZoomLinkUC:   setZoomLink,
		SuggestSlotsUC:  suggestSlots,
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

func newAdminHandler(
	statistics *adminuc.Statistics,
	listUsers *adminuc.ListUsers,
	getUser *adminuc.GetUser,
	updateUser *adminuc.UpdateUser,
	setBlocked *adminuc.SetBlocked,
) *adminhandler.Handler {
	return &adminhandler.Handler{
		StatisticsUC: statistics,
		ListUsersUC:  listUsers,
		GetUserUC:    getUser,
		UpdateUserUC: updateUser,
		SetBlockedUC: setBlocked,
	}
}

func newGinEngine(
	cfg config.Config,
	log *slog.Logger,
	auth *authhandler.Handler,
	appt *appointmenthandler.Handler,
	photoReview *photoreviewhandler.Handler,
	adminH *adminhandler.Handler,
	predict *predictionhandler.Handler,
	subscriptionH *subscriptionhandler.Handler,
	submitPhoto *photoreviewuc.SubmitFromTelegram,
	answerPreCheckout *subscriptionuc.AnswerPreCheckout,
	confirmPayment *subscriptionuc.ConfirmPayment,
	tokens port.TokenIssuer,
	users port.UserRepository,
	bot *tgbotapi.BotAPI,
	tgHTTP infrahttp.TelegramHTTPClient,
) *gin.Engine {
	if len(cfg.CORSAllowOrigins) == 0 {
		log.Warn("cors: CORS_ALLOW_ORIGINS is empty — all origins allowed (ok for dev)")
	} else {
		log.Info("cors: allowed origins", "origins", cfg.CORSAllowOrigins)
	}
	return infrahttp.NewRouter(infrahttp.RouterParams{
		Config:            cfg,
		Log:               log,
		AuthHandler:       auth,
		AppointmentH:      appt,
		PhotoReviewH:      photoReview,
		AdminH:            adminH,
		PredictionH:       predict,
		SubscriptionH:     subscriptionH,
		SubmitPhoto:       submitPhoto,
		AnswerPreCheckout: answerPreCheckout,
		ConfirmPayment:    confirmPayment,
		Tokens:            tokens,
		Users:             users,
		Bot:               bot,
		TelegramHTTP:      tgHTTP,
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
