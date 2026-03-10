// Package router configures all HTTP route definitions.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "phone-call-receptionist/backend/docs"
	"phone-call-receptionist/backend/internal/domain/port"
	"phone-call-receptionist/backend/internal/ui/rest/gin/handlers"
	"phone-call-receptionist/backend/internal/ui/rest/gin/middleware"
)

// Handlers holds all HTTP handler instances.
type Handlers struct {
	// Auth handles authentication endpoints.
	Auth *handlers.AuthHandler
	// Call handles call endpoints.
	Call *handlers.CallHandler
	// Appointment handles appointment endpoints.
	Appointment *handlers.AppointmentHandler
	// Lead handles lead endpoints.
	Lead *handlers.LeadHandler
	// Knowledge handles knowledge base endpoints.
	Knowledge *handlers.KnowledgeHandler
	// Dashboard handles dashboard endpoints.
	Dashboard *handlers.DashboardHandler
	// Settings handles settings endpoints.
	Settings *handlers.SettingsHandler
	// Webhook handles Twilio webhook endpoints.
	Webhook *handlers.WebhookHandler
	// Health handles health check endpoints.
	Health *handlers.HealthHandler
	// WS handles WebSocket connections.
	WS *handlers.WSHandler
}

// Setup creates and configures the Gin router with all routes and middleware.
func Setup(h Handlers, voiceCaller port.VoiceCaller, jwtSecret string, frontendURL string, logger *zerolog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(frontendURL))
	r.Use(middleware.Logger(logger))

	api := r.Group("/api")
	{
		api.GET("/health", h.Health.Check)
		api.POST("/auth/signup", h.Auth.SignUp)
		api.POST("/auth/signin", h.Auth.SignIn)

		webhooks := api.Group("/webhooks/twilio")
		webhooks.Use(middleware.TwilioSignature(voiceCaller, logger))
		{
			webhooks.POST("/voice", h.Webhook.HandleVoice)
			webhooks.POST("/gather", h.Webhook.HandleGather)
			webhooks.POST("/status", h.Webhook.HandleStatus)
			webhooks.POST("/recording", h.Webhook.HandleRecording)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(jwtSecret, logger))
		{
			protected.GET("/users/me", h.Auth.Me)
			protected.PUT("/users/me", h.Auth.UpdateProfile)
			protected.POST("/users/me/password", h.Auth.ChangePassword)

			protected.GET("/calls", h.Call.List)
			protected.GET("/calls/stats", h.Call.Stats)
			protected.GET("/calls/:id", h.Call.Detail)
			protected.GET("/calls/:id/rag-queries", h.Call.RAGQueries)

			protected.POST("/appointments", h.Appointment.Create)
			protected.GET("/appointments", h.Appointment.List)
			protected.GET("/appointments/availability", h.Appointment.Availability)
			protected.GET("/appointments/:id", h.Appointment.FindByID)
			protected.PUT("/appointments/:id", h.Appointment.Reschedule)
			protected.DELETE("/appointments/:id", h.Appointment.Cancel)

			protected.GET("/leads", h.Lead.List)
			protected.GET("/leads/:id", h.Lead.FindByID)
			protected.PUT("/leads/:id", h.Lead.Update)

			protected.POST("/knowledge/documents", h.Knowledge.Upload)
			protected.GET("/knowledge/documents", h.Knowledge.List)
			protected.GET("/knowledge/documents/:id", h.Knowledge.FindByID)
			protected.DELETE("/knowledge/documents/:id", h.Knowledge.Delete)
			protected.POST("/knowledge/documents/:id/reindex", h.Knowledge.Reindex)
			protected.POST("/knowledge/search", h.Knowledge.Search)

			protected.GET("/dashboard/stats", h.Dashboard.Stats)
			protected.GET("/dashboard/costs", h.Dashboard.Costs)
			protected.GET("/dashboard/volume", h.Dashboard.Volume)

			protected.GET("/settings", h.Settings.Find)
			protected.PUT("/settings", h.Settings.Update)
		}
	}

	r.GET("/ws", h.WS.HandleWebSocket)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
