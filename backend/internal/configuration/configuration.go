// Package configuration provides centralized application configuration
// loaded from environment variables using Viper.
package configuration

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config is the global application configuration instance.
// It is initialized via init() and available throughout the application.
var Config *AppConfig

func init() {
	Config = loadFromEnv()
}

// AppConfig holds all application configuration values.
type AppConfig struct {
	// Server contains HTTP server configuration.
	Server ServerConfig
	// Database contains PostgreSQL connection configuration.
	Database DatabaseConfig
	// Redis contains Redis connection configuration.
	Redis RedisConfig
	// Weaviate contains Weaviate vector DB configuration.
	Weaviate WeaviateConfig
	// JWT contains JSON Web Token configuration.
	JWT JWTConfig
	// LLM contains LLM provider API keys.
	LLM LLMConfig
	// Voice contains voice service configuration.
	Voice VoiceConfig
	// GoogleCalendar contains Google Calendar API configuration.
	GoogleCalendar GoogleCalendarConfig
	// Storage contains file storage paths.
	Storage StorageConfig
	// FrontendURL is the frontend URL for CORS configuration.
	FrontendURL string
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	// Port is the HTTP server listen port.
	Port string
	// Env is the environment name (development, production).
	Env string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	// Host is the database server hostname.
	Host string
	// Port is the database server port.
	Port string
	// Name is the database name.
	Name string
	// User is the database user.
	User string
	// Password is the database password.
	Password string
	// SSLMode is the SSL connection mode.
	SSLMode string
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	// Host is the Redis server hostname.
	Host string
	// Port is the Redis server port.
	Port string
	// Password is the Redis authentication password.
	Password string
}

// WeaviateConfig holds Weaviate vector DB settings.
type WeaviateConfig struct {
	// Host is the Weaviate server hostname.
	Host string
	// Port is the Weaviate server port.
	Port string
	// Scheme is the connection scheme (http or https).
	Scheme string
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	// Secret is the signing key for JWT tokens.
	Secret string
	// ExpiryHours is the token expiration time in hours.
	ExpiryHours int
}

// LLMConfig holds API keys for all LLM providers.
type LLMConfig struct {
	// GeminiAPIKey is the Google Gemini API key.
	GeminiAPIKey string
	// ClaudeAPIKey is the Anthropic Claude API key.
	ClaudeAPIKey string
	// OpenAIAPIKey is the OpenAI API key.
	OpenAIAPIKey string
	// GLMAPIKey is the GLM API key.
	GLMAPIKey string
	// MistralAPIKey is the Mistral API key.
	MistralAPIKey string
	// DeepSeekAPIKey is the DeepSeek API key.
	DeepSeekAPIKey string
}

// VoiceConfig holds voice service API keys and settings.
type VoiceConfig struct {
	// TwilioAccountSID is the Twilio account identifier.
	TwilioAccountSID string
	// TwilioAuthToken is the Twilio authentication token.
	TwilioAuthToken string
	// TwilioPhoneNumber is the Twilio phone number for outbound calls/SMS.
	TwilioPhoneNumber string
	// ElevenLabsAPIKey is the ElevenLabs TTS API key.
	ElevenLabsAPIKey string
}

// GoogleCalendarConfig holds Google Calendar API settings.
type GoogleCalendarConfig struct {
	// CredentialsJSON is the service account credentials in JSON format.
	CredentialsJSON string
	// CalendarID is the Google Calendar ID to manage.
	CalendarID string
}

// StorageConfig holds file storage path settings.
type StorageConfig struct {
	// UploadPath is the directory path for uploaded files.
	UploadPath string
	// AudioCachePath is the directory path for cached audio files.
	AudioCachePath string
}

func loadFromEnv() *AppConfig {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("No .env file found, using environment variables")
	}

	setDefaults()

	return &AppConfig{
		Server: ServerConfig{
			Port: viper.GetString("PORT"),
			Env:  viper.GetString("ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			Name:     viper.GetString("DB_NAME"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			SSLMode:  viper.GetString("DB_SSL_MODE"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
		},
		Weaviate: WeaviateConfig{
			Host:   viper.GetString("WEAVIATE_HOST"),
			Port:   viper.GetString("WEAVIATE_PORT"),
			Scheme: viper.GetString("WEAVIATE_SCHEME"),
		},
		JWT: JWTConfig{
			Secret:      viper.GetString("JWT_SECRET"),
			ExpiryHours: viper.GetInt("JWT_EXPIRY_HOURS"),
		},
		LLM: LLMConfig{
			GeminiAPIKey:   viper.GetString("GEMINI_API_KEY"),
			ClaudeAPIKey:   viper.GetString("CLAUDE_API_KEY"),
			OpenAIAPIKey:   viper.GetString("OPENAI_API_KEY"),
			GLMAPIKey:      viper.GetString("GLM_API_KEY"),
			MistralAPIKey:  viper.GetString("MISTRAL_API_KEY"),
			DeepSeekAPIKey: viper.GetString("DEEPSEEK_API_KEY"),
		},
		Voice: VoiceConfig{
			TwilioAccountSID:  viper.GetString("TWILIO_ACCOUNT_SID"),
			TwilioAuthToken:   viper.GetString("TWILIO_AUTH_TOKEN"),
			TwilioPhoneNumber: viper.GetString("TWILIO_PHONE_NUMBER"),
			ElevenLabsAPIKey:  viper.GetString("ELEVENLABS_API_KEY"),
		},
		GoogleCalendar: GoogleCalendarConfig{
			CredentialsJSON: viper.GetString("GOOGLE_CALENDAR_CREDENTIALS_JSON"),
			CalendarID:      viper.GetString("GOOGLE_CALENDAR_ID"),
		},
		Storage: StorageConfig{
			UploadPath:     viper.GetString("UPLOAD_PATH"),
			AudioCachePath: viper.GetString("AUDIO_CACHE_PATH"),
		},
		FrontendURL: viper.GetString("FRONTEND_URL"),
	}
}

func setDefaults() {
	// No defaults — all values must be provided via .env or environment variables.
}
