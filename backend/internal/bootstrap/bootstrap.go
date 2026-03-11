// Package bootstrap wires all application dependencies together.
package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"
	goRedis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"

	"phone-call-receptionist/backend/internal/configuration"
	"phone-call-receptionist/backend/internal/domain/api"
	"phone-call-receptionist/backend/internal/domain/port"
	"phone-call-receptionist/backend/internal/infrastructure/claude"
	"phone-call-receptionist/backend/internal/infrastructure/deepseek"
	"phone-call-receptionist/backend/internal/infrastructure/elevenlabs"
	"phone-call-receptionist/backend/internal/infrastructure/filesystem"
	"phone-call-receptionist/backend/internal/infrastructure/gemini"
	"phone-call-receptionist/backend/internal/infrastructure/glm"
	"phone-call-receptionist/backend/internal/infrastructure/googlecalendar"
	"phone-call-receptionist/backend/internal/infrastructure/llm"
	"phone-call-receptionist/backend/internal/infrastructure/mistral"
	"phone-call-receptionist/backend/internal/infrastructure/openai"
	"phone-call-receptionist/backend/internal/infrastructure/postgres"
	infraRedis "phone-call-receptionist/backend/internal/infrastructure/redis"
	"phone-call-receptionist/backend/internal/infrastructure/twilio"
	infraWeaviate "phone-call-receptionist/backend/internal/infrastructure/weaviate"
	"phone-call-receptionist/backend/internal/infrastructure/ws"
	"phone-call-receptionist/backend/internal/ui/rest/gin/handlers"
	"phone-call-receptionist/backend/internal/ui/rest/gin/router"
)

// App holds all initialized application components.
type App struct {
	// Router is the configured Gin engine.
	Router *gin.Engine
	// PostgresClient is the database client for cleanup.
	PostgresClient *postgres.Client
	// Logger is the application logger.
	Logger *zerolog.Logger
}

// Initialize creates and wires all application dependencies.
func Initialize(logger *zerolog.Logger) (*App, error) {
	cfg := configuration.Config

	// PostgreSQL
	pgClient, err := postgres.NewClient(
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
		cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Repository adapters
	userPort := postgres.NewUserAdapter(pgClient, logger)
	docPort := postgres.NewKnowledgeDocumentAdapter(pgClient, logger)
	callPort := postgres.NewInboundCallAdapter(pgClient, logger)
	apptPort := postgres.NewAppointmentAdapter(pgClient, logger)
	leadPort := postgres.NewLeadAdapter(pgClient, logger)
	smsLogPort := postgres.NewSMSLogAdapter(pgClient, logger)
	audioCachePort := postgres.NewAudioCacheAdapter(pgClient, logger)
	settingsPort := postgres.NewSystemSettingsAdapter(pgClient, logger)

	// Weaviate client
	weaviateConfig := weaviate.Config{
		Host:   fmt.Sprintf("%s:%s", cfg.Weaviate.Host, cfg.Weaviate.Port),
		Scheme: cfg.Weaviate.Scheme,
	}
	weaviateClient, err := weaviate.NewClient(weaviateConfig)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create Weaviate client, vector search will be unavailable")
	}
	vectorDB := infraWeaviate.NewWeaviateAdapter(weaviateClient, logger)

	// LLM providers — ordered by preference: Gemini → Mistral → DeepSeek → others
	var llmProviders []port.LLM
	geminiAdapter := gemini.NewGeminiAdapter(cfg.LLM.GeminiAPIKey, logger)
	if cfg.LLM.GeminiAPIKey != "" {
		llmProviders = append(llmProviders, geminiAdapter)
	}
	if cfg.LLM.MistralAPIKey != "" {
		llmProviders = append(llmProviders, mistral.NewMistralAdapter(cfg.LLM.MistralAPIKey, logger))
	}
	if cfg.LLM.DeepSeekAPIKey != "" {
		llmProviders = append(llmProviders, deepseek.NewDeepSeekAdapter(cfg.LLM.DeepSeekAPIKey, logger))
	}
	if cfg.LLM.ClaudeAPIKey != "" {
		llmProviders = append(llmProviders, claude.NewClaudeAdapter(cfg.LLM.ClaudeAPIKey, logger))
	}
	if cfg.LLM.OpenAIAPIKey != "" {
		llmProviders = append(llmProviders, openai.NewOpenAIAdapter(cfg.LLM.OpenAIAPIKey, logger))
	}
	if cfg.LLM.GLMAPIKey != "" {
		llmProviders = append(llmProviders, glm.NewGLMAdapter(cfg.LLM.GLMAPIKey, logger))
	}

	llmRouter := llm.NewRouter(llmProviders, logger)
	var embeddingPort port.Embedding = geminiAdapter

	// External services
	ttsPort := elevenlabs.NewElevenLabsAdapter(cfg.Voice.ElevenLabsAPIKey, logger)
	_ = geminiAdapter.AsSpeechToText()
	voiceCallerPort := twilio.NewTwilioAdapter(cfg.Voice.TwilioAccountSID, cfg.Voice.TwilioAuthToken, cfg.Voice.TwilioPhoneNumber, logger)

	calendarPort, err := googlecalendar.NewGoogleCalendarAdapter(cfg.GoogleCalendar.CredentialsJSON, cfg.GoogleCalendar.CalendarID, logger)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create Google Calendar adapter, calendar features will be unavailable")
	}

	redisClient := goRedis.NewClient(&goRedis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})
	cachePort := infraRedis.NewRedisAdapter(redisClient, logger)

	fileStoragePort, err := filesystem.NewFilesystemAdapter(cfg.Storage.UploadPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem adapter: %w", err)
	}

	// WebSocket hub
	wsHub := ws.NewHub(logger)

	// Domain services
	ragApi := api.NewRAGApi(docPort, vectorDB, embeddingPort, llmRouter, fileStoragePort, logger)
	apptApi := api.NewAppointmentApi(apptPort, calendarPort, voiceCallerPort, smsLogPort, wsHub, logger)
	voiceCallApi := api.NewVoiceCallApi(callPort, leadPort, ragApi, apptApi, ttsPort, cachePort, audioCachePort, fileStoragePort, wsHub, settingsPort, logger)
	leadApi := api.NewLeadApi(leadPort, logger)
	kbApi := api.NewKnowledgeBaseApi(docPort, ragApi, logger)
	authApi := api.NewAuthApi(userPort, cfg.JWT.Secret, cfg.JWT.ExpiryHours, logger)
	dashboardApi := api.NewDashboardApi(callPort, apptPort, leadPort, logger)

	// HTTP handlers
	h := router.Handlers{
		Auth:        handlers.NewAuthHandler(authApi, logger),
		Call:        handlers.NewCallHandler(voiceCallApi, dashboardApi, logger),
		Appointment: handlers.NewAppointmentHandler(apptApi, logger),
		Lead:        handlers.NewLeadHandler(leadApi, logger),
		Knowledge:   handlers.NewKnowledgeHandler(kbApi, ragApi, logger),
		Dashboard:   handlers.NewDashboardHandler(dashboardApi, logger),
		Settings:    handlers.NewSettingsHandler(settingsPort, logger),
		Webhook:     handlers.NewWebhookHandler(voiceCallApi, logger),
		Health:      handlers.NewHealthHandler(logger),
		WS:          handlers.NewWSHandler(wsHub, logger),
	}

	r := router.Setup(h, voiceCallerPort, cfg.JWT.Secret, cfg.FrontendURL, logger)

	return &App{
		Router:         r,
		PostgresClient: pgClient,
		Logger:         logger,
	}, nil
}

// Close gracefully shuts down all application resources.
func (a *App) Close() {
	if a.PostgresClient != nil {
		a.PostgresClient.Close()
	}
}
