package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anuarkuanysh/dental_project/infra/config"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/appointment"
	authhandler "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/auth"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	"github.com/anuarkuanysh/dental_project/internal/gemini"
	"github.com/anuarkuanysh/dental_project/internal/handler"
	"github.com/anuarkuanysh/dental_project/internal/imageproc"
	"github.com/anuarkuanysh/dental_project/internal/middleware"
	"github.com/anuarkuanysh/dental_project/internal/port"
	"github.com/anuarkuanysh/dental_project/internal/telegrambot"
)

// RouterParams wires HTTP routes.
type RouterParams struct {
	Config         config.Config
	Log            *slog.Logger
	AuthHandler    *authhandler.Handler
	AppointmentH   *appointment.Handler
	Tokens         port.TokenIssuer
	Bot          *tgbotapi.BotAPI
	TelegramHTTP TelegramHTTPClient
	GeminiHTTP   GeminiHTTPClient
}

// NewRouter builds the Gin engine with webhook and REST API routes.
func NewRouter(p RouterParams) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		jwtmiddleware.CORS(p.Config.CORSAllowOrigins),
		middleware.Recovery(p.Log),
		middleware.RequestID(),
		middleware.SlogLogger(p.Log),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/auth/telegram", p.AuthHandler.Telegram)

	protected := r.Group("/")
	protected.Use(jwtmiddleware.JWTAuth(p.Tokens))
	protected.POST("/appointments", p.AppointmentH.Create)
	protected.GET("/appointments/me", p.AppointmentH.ListMine)
	protected.GET("/appointments",
		jwtmiddleware.RequireDoctor(),
		p.AppointmentH.ListForDoctor,
	)

	tgClient := telegrambot.New(p.Bot, p.TelegramHTTP.Client)
	imgProc := imageproc.New(p.Config.MaxImageDimension)
	geminiClient := gemini.New(p.GeminiHTTP.Client, p.Config.GeminiAPIKey, p.Config.GeminiModel, p.Log)

	r.POST("/webhook",
		middleware.WebhookSecret(p.Config.TelegramWebhookSecret),
		handler.Webhook(handler.WebhookDeps{
			Downloader: tgClient,
			Sender:     tgClient,
			Analyzer:   geminiClient,
			Images:     imgProc,
			Log:        p.Log,
			ReqTimeout: p.Config.RequestTimeout,
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
