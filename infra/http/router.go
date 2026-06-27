package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anuarkuanysh/dental_project/infra/config"
	adminhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/admin"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/appointment"
	authhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/auth"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	photoreviewhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/photo_review"
	predictionhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/prediction"
	subscriptionhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/subscription"
	"github.com/anuarkuanysh/dental_project/internal/handler"
	"github.com/anuarkuanysh/dental_project/internal/middleware"
	"github.com/anuarkuanysh/dental_project/internal/port"
	"github.com/anuarkuanysh/dental_project/internal/telegrambot"
	photoreviewuc "github.com/anuarkuanysh/dental_project/internal/usecase/photo_review"
	subscriptionuc "github.com/anuarkuanysh/dental_project/internal/usecase/subscription"
)

// RouterParams wires HTTP routes.
type RouterParams struct {
	Config         config.Config
	Log            *slog.Logger
	AuthHandler    *authhandler.Handler
	AppointmentH   *appointment.Handler
	PhotoReviewH   *photoreviewhandler.Handler
	AdminH         *adminhandler.Handler
	PredictionH    *predictionhandler.Handler
	SubscriptionH  *subscriptionhandler.Handler
	SubmitPhoto    *photoreviewuc.SubmitFromTelegram
	AnswerPreCheckout *subscriptionuc.AnswerPreCheckout
	ConfirmPayment *subscriptionuc.ConfirmPayment
	Tokens         port.TokenIssuer
	Users          port.UserRepository
	Bot            *tgbotapi.BotAPI
	TelegramHTTP   TelegramHTTPClient
}

// NewRouter builds the Gin engine with webhook and REST API routes.
func NewRouter(p RouterParams) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		middleware.Recovery(p.Log),
		middleware.RequestID(),
		middleware.SlogLogger(p.Log),
		jwtmiddleware.CORS(p.Log, p.Config.CORSAllowOrigins),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/auth/telegram", p.AuthHandler.Telegram)

	protected := r.Group("/")
	protected.Use(jwtmiddleware.JWTAuth(p.Tokens))
	protected.Use(jwtmiddleware.RequireNotBlocked(p.Users))
	protected.POST("/appointments", p.AppointmentH.Create)
	protected.GET("/appointments/me", p.AppointmentH.ListMine)
	protected.GET("/appointments",
		jwtmiddleware.RequireDoctor(),
		p.AppointmentH.ListForDoctor,
	)
	protected.PATCH("/appointments/:id/respond",
		jwtmiddleware.RequireDoctor(),
		p.AppointmentH.Respond,
	)
	protected.PATCH("/appointments/:id/zoom-link",
		jwtmiddleware.RequireDoctor(),
		p.AppointmentH.SetZoomLink,
	)
	protected.POST("/appointments/:id/suggest-slots",
		jwtmiddleware.RequireDoctor(),
		p.AppointmentH.SuggestSlots,
	)

	if p.PhotoReviewH != nil {
		doctor := protected.Group("/submissions")
		doctor.Use(jwtmiddleware.RequireDoctor())
		doctor.GET("/pending", p.PhotoReviewH.ListPending)
		doctor.GET("/answered", p.PhotoReviewH.ListAnswered)
		doctor.GET("/:id", p.PhotoReviewH.Get)
		doctor.GET("/:id/photo", p.PhotoReviewH.GetImage)
		doctor.POST("/:id/draft", p.PhotoReviewH.GenerateDraft)
		doctor.POST("/:id/respond", p.PhotoReviewH.Respond)
	}

	if p.AdminH != nil {
		adminGroup := protected.Group("/admin")
		adminGroup.Use(jwtmiddleware.RequireAdmin())
		adminGroup.GET("/statistics", p.AdminH.GetStatistics)
		adminGroup.GET("/users", p.AdminH.ListUsers)
		adminGroup.GET("/users/:id", p.AdminH.GetUser)
		adminGroup.PATCH("/users/:id", p.AdminH.UpdateUser)
		adminGroup.PATCH("/users/:id/block", p.AdminH.SetBlocked)
	}

	if p.PredictionH != nil {
		protected.POST("/predict", p.PredictionH.Predict)
	}

	if p.SubscriptionH != nil {
		protected.GET("/subscription/me", p.SubscriptionH.GetMe)
		protected.POST("/subscription/invoice", p.SubscriptionH.CreateInvoice)
	}

	tgClient := telegrambot.New(p.Bot, p.TelegramHTTP.Client)

	r.POST("/webhook",
		middleware.WebhookSecret(p.Config.TelegramWebhookSecret),
		handler.Webhook(handler.WebhookDeps{
			SubmitPhoto:       p.SubmitPhoto,
			AnswerPreCheckout: p.AnswerPreCheckout,
			ConfirmPayment:    p.ConfirmPayment,
			Sender:            tgClient,
			Log:               p.Log,
			ReqTimeout:        p.Config.RequestTimeout,
		}),
	)

	return r
}

// NewHTTPServer returns a configured http.Server for the Gin engine.
func NewHTTPServer(addr string, handler http.Handler, reqTimeout time.Duration) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       reqTimeout + 5*time.Second,
		WriteTimeout:      reqTimeout + 5*time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
