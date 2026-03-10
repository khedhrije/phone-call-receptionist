You are an expert Go engineer. Build a production-grade Voice AI
Receptionist system from scratch for an IT services firm.

Before writing any code, read and analyze these repositories
to understand my coding style, architecture patterns, and conventions.
Use them as inspiration only — do not copy code directly:

- https://github.com/khedhrije/rag
- https://github.com/khedhrije/phone-calls-maker

---

## PROJECT NAME
voice-ai-receptionist

---

## ARCHITECTURE
Strict Hexagonal Architecture (Ports & Adapters).
Dependency direction: UI → Domain ← Infrastructure.
The domain layer NEVER imports infrastructure or UI packages.

voice-ai-receptionist/
├── backend/
│   ├── cmd/server/main.go              # Entry point, graceful shutdown
│   ├── internal/
│   │   ├── bootstrap/                  # Dependency injection wiring
│   │   ├── configuration/              # Viper config from .env
│   │   ├── domain/
│   │   │   ├── model/                  # Pure domain entities
│   │   │   ├── port/                   # Interface contracts
│   │   │   ├── api/                    # Business logic services
│   │   │   └── errors/                 # Domain error types
│   │   ├── infrastructure/
│   │   │   ├── postgres/               # Repository adapters (sqlx)
│   │   │   ├── weaviate/               # Vector DB adapter
│   │   │   ├── gemini/                 # Gemini LLM + embedding
│   │   │   ├── glm/                    # GLM adapter
│   │   │   ├── mistral/                # Mistral adapter
│   │   │   ├── deepseek/               # DeepSeek adapter
│   │   │   ├── claude/                 # Claude adapter
│   │   │   ├── llm/                    # LLM router (selects provider)
│   │   │   ├── embedding/              # Embedding service
│   │   │   ├── twilio/                 # Voice + SMS adapter
│   │   │   ├── elevenlabs/             # TTS adapter with audio cache
│   │   │   ├── googlecalendar/         # Google Calendar adapter
│   │   │   ├── redis/                  # Cache adapter
│   │   │   ├── filesystem/             # File storage adapter
│   │   │   └── ws/                     # WebSocket hub
│   │   ├── twiml/                      # TwiML XML builder
│   │   └── ui/rest/gin/
│   │       ├── handlers/               # HTTP handlers
│   │       ├── middleware/             # Auth, CORS, logging, rate-limit
│   │       └── router/                 # Route definitions
│   ├── pkg/
│   │   ├── dtos/                       # Request/response DTOs
│   │   ├── helpers/                    # JWT, hashing, context helpers
│   │   └── vdb/                        # DB connection + migrations
│   ├── docs/                           # Swagger auto-generated
│   ├── migrations/                     # SQL migration files
│   ├── Makefile
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── pages/                      # File-based routing
│   │   ├── components/                 # Reusable components
│   │   ├── composables/                # Business logic hooks
│   │   ├── layouts/                    # App layouts
│   │   └── utils/                      # API client
│   ├── nuxt.config.ts
│   └── Dockerfile
├── nginx/
│   └── nginx.conf
├── docker-compose.yml
├── Makefile
├── .env.example
├── .gitignore
└── README.md

---

## TECH STACK

### Backend
- Go 1.24
- Gin framework (HTTP router)
- PostgreSQL 16 + sqlx (NO ORM — raw SQL only)
- Weaviate (vector database)
- Redis 7 (audio cache + rate limiting)
- golang-migrate (migrations)
- zerolog (structured logging)
- Viper (config from .env)
- Swaggo (Swagger/OpenAPI docs)
- golang-jwt/jwt/v5 (auth)
- Prometheus (metrics)

### AI / LLM
- Gemini Pro (primary LLM for answer generation)
- Gemini text-embedding-004 (embeddings)
- Claude API (fallback LLM)
- GPT-4o (fallback LLM)
- GLM (fallback LLM)
- Mistral (fallback LLM)
- DeepSeek (fallback LLM)
- LLM Router: tries providers in order until success

### Voice
- Twilio Voice API (inbound calls + TwiML)
- Twilio SMS (confirmations)
- ElevenLabs (TTS with SHA-256 audio caching)
- Deepgram (STT — Speech to Text)

### Frontend
- Nuxt 3 + TypeScript
- Tailwind CSS
- i18n (en/fr)
- Recharts (analytics)
- Vue Flow (call script visualizer)

### Infrastructure
- Docker + Docker Compose
- Nginx (reverse proxy)

---

## DOMAIN MODELS

### User
type User struct {
ID           uuid.UUID
Email        string
PasswordHash string
Role         string  // super_admin, admin, user
IsBlocked    bool
CreatedAt    time.Time
UpdatedAt    time.Time
}

### KnowledgeDocument
type KnowledgeDocument struct {
ID          uuid.UUID
Filename    string
MimeType    string
ChunkCount  int
Status      string  // pending, indexing, indexed, failed
IndexedAt   *time.Time
CreatedAt   time.Time
}

### Chunk
type Chunk struct {
ID         uuid.UUID
DocumentID uuid.UUID
Content    string
PageNumber int
ChunkIndex int
Embedding  []float32
CreatedAt  time.Time
}

### InboundCall
type InboundCall struct {
ID              uuid.UUID
TwilioCallSID   string
CallerPhone     string
Status          string  // ringing, in_progress, completed, failed
Transcript      []TranscriptEntry
RAGQueries      []RAGQuery
DurationSeconds int
TwilioCostUSD   float64
LLMCostUSD      float64
TotalCostUSD    float64
CreatedAt       time.Time
EndedAt         *time.Time
}

type TranscriptEntry struct {
Speaker string  // caller, assistant
Text    string
At      time.Time
}

type RAGQuery struct {
Query     string
Chunks    []string
Response  string
Provider  string
Tokens    int
At        time.Time
}

### Appointment
type Appointment struct {
ID             uuid.UUID
CallID         uuid.UUID
CallerPhone    string
CallerName     string
CallerEmail    string
ServiceType    string
ScheduledAt    time.Time
DurationMins   int
Status         string  // pending, confirmed, rescheduled, cancelled
GoogleEventID  string
SMSSentAt      *time.Time
CreatedAt      time.Time
UpdatedAt      time.Time
}

### Lead
type Lead struct {
ID          uuid.UUID
CallID      uuid.UUID
Phone       string
Name        string
Email       string
Status      string  // new, contacted, qualified, converted, lost
Notes       string
CreatedAt   time.Time
UpdatedAt   time.Time
}

### SMSLog
type SMSLog struct {
ID          uuid.UUID
CallID      *uuid.UUID
ToPhone     string
Message     string
TwilioSID   string
Status      string  // queued, sent, delivered, failed
CreatedAt   time.Time
}

### AudioCache
type AudioCache struct {
ID        uuid.UUID
Hash      string    // SHA-256 of text + voice_id
VoiceID   string
AudioURL  string
CreatedAt time.Time
}

### SystemSettings
type SystemSettings struct {
DefaultLLMProvider  string
DefaultVoiceID      string
TopK                int
MaxCallDurationSecs int
UpdatedAt           time.Time
}

---

## PORT INTERFACES (Domain Layer)

### Repositories
type UserRepository interface {
Create(ctx context.Context, user *model.User) error
FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
FindByEmail(ctx context.Context, email string) (*model.User, error)
List(ctx context.Context, filters UserFilters) ([]*model.User, error)
Update(ctx context.Context, user *model.User) error
Delete(ctx context.Context, id uuid.UUID) error
}

type KnowledgeDocumentRepository interface {
Create(ctx context.Context, doc *model.KnowledgeDocument) error
FindByID(ctx context.Context, id uuid.UUID) (*model.KnowledgeDocument, error)
List(ctx context.Context) ([]*model.KnowledgeDocument, error)
Update(ctx context.Context, doc *model.KnowledgeDocument) error
Delete(ctx context.Context, id uuid.UUID) error
}

type InboundCallRepository interface {
Create(ctx context.Context, call *model.InboundCall) error
FindByID(ctx context.Context, id uuid.UUID) (*model.InboundCall, error)
FindByTwilioSID(ctx context.Context, sid string) (*model.InboundCall, error)
List(ctx context.Context, filters CallFilters) ([]*model.InboundCall, error)
Update(ctx context.Context, call *model.InboundCall) error
}

type AppointmentRepository interface {
Create(ctx context.Context, appt *model.Appointment) error
FindByID(ctx context.Context, id uuid.UUID) (*model.Appointment, error)
FindByPhone(ctx context.Context, phone string) ([]*model.Appointment, error)
List(ctx context.Context, filters ApptFilters) ([]*model.Appointment, error)
Update(ctx context.Context, appt *model.Appointment) error
Delete(ctx context.Context, id uuid.UUID) error
}

type LeadRepository interface {
Create(ctx context.Context, lead *model.Lead) error
FindByID(ctx context.Context, id uuid.UUID) (*model.Lead, error)
FindByPhone(ctx context.Context, phone string) (*model.Lead, error)
List(ctx context.Context, filters LeadFilters) ([]*model.Lead, error)
Update(ctx context.Context, lead *model.Lead) error
}

type SMSLogRepository interface {
Create(ctx context.Context, log *model.SMSLog) error
FindByID(ctx context.Context, id uuid.UUID) (*model.SMSLog, error)
List(ctx context.Context, callID uuid.UUID) ([]*model.SMSLog, error)
}

### Service Ports
type VectorStore interface {
StoreChunk(ctx context.Context, chunk *model.Chunk) error
SearchSimilar(ctx context.Context, embedding []float32, topK int) ([]*model.Chunk, error)
DeleteByDocumentID(ctx context.Context, documentID uuid.UUID) error
}

type EmbeddingService interface {
Embed(ctx context.Context, text string) ([]float32, error)
}

type LLMService interface {
Generate(ctx context.Context, systemPrompt, userPrompt string) (string, int, error)
Provider() string
}

type TextToSpeech interface {
Synthesize(ctx context.Context, text, voiceID string) ([]byte, error)
}

type SpeechToText interface {
Transcribe(ctx context.Context, audioData []byte) (string, error)
}

type VoiceCaller interface {
SendSMS(ctx context.Context, to, message string) (string, error)
GenerateTwiML(nodes []twiml.Node) (string, error)
}

type CalendarService interface {
CheckAvailability(ctx context.Context, from, to time.Time) ([]TimeSlot, error)
CreateEvent(ctx context.Context, appt *model.Appointment) (string, error)
UpdateEvent(ctx context.Context, eventID string, appt *model.Appointment) error
DeleteEvent(ctx context.Context, eventID string) error
}

type CacheService interface {
Get(ctx context.Context, key string) ([]byte, error)
Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
Delete(ctx context.Context, key string) error
}

type EventBroadcaster interface {
Broadcast(event interface{})
}

type FileStorage interface {
Save(ctx context.Context, filename string, data []byte) (string, error)
Load(ctx context.Context, filename string) ([]byte, error)
Delete(ctx context.Context, filename string) error
}

---

## BUSINESS LOGIC SERVICES (Domain API Layer)

### RAGService
- IngestDocument(ctx, file, filename) — extract text, chunk, embed, store
- QueryKnowledgeBase(ctx, question, topK) — embed query, retrieve chunks,
  generate answer with LLM, return answer + sources
- DeleteDocument(ctx, id) — remove from DB and vector store
- ReindexDocument(ctx, id) — re-chunk and re-embed

### VoiceCallService
- HandleInbound(ctx, twilioSID, callerPhone) — create call record,
  generate greeting TwiML
- ProcessSpeech(ctx, twilioSID, transcript) — run RAG query, detect intent
  (question / book / reschedule / cancel), generate voice response TwiML
- EndCall(ctx, twilioSID, duration) — finalize call record, calculate costs
- GetCallHistory(ctx, filters) — paginated call list
- GetCallDetail(ctx, id) — full call with transcript

### AppointmentService
- Book(ctx, callID, phone, name, email, service, slot) — validate,
  create Google Calendar event, create DB record, send SMS confirmation
- Reschedule(ctx, id, newSlot) — update Calendar event + DB + send SMS
- Cancel(ctx, id) — delete Calendar event, update DB status, send SMS
- GetAvailability(ctx, from, to) — query Google Calendar free slots
- List(ctx, filters) — paginated appointment list

### LeadService
- CreateOrUpdate(ctx, phone, name, email, callID) — upsert lead from call
- UpdateStatus(ctx, id, status, notes) — manual status update
- List(ctx, filters) — paginated lead list

### KnowledgeBaseService
- Upload(ctx, file) — save file, trigger async ingestion
- List(ctx) — all documents with status
- Delete(ctx, id) — remove document + chunks
- GetIndexingStatus(ctx, id) — check indexing progress

### AuthService
- SignUp(ctx, email, password) — hash password, create user, return JWT
- SignIn(ctx, email, password) — verify, return JWT
- RefreshToken(ctx, token) — validate + issue new token
- Me(ctx, userID) — get current user

### DashboardService
- GetStats(ctx) — total calls, appointments, leads, avg cost per call
- GetCostAnalytics(ctx, from, to) — cost breakdown by day/provider
- GetCallVolume(ctx, from, to) — calls per day

---

## INBOUND CALL FLOW

1. Twilio calls POST /webhooks/twilio/voice
   → Create InboundCall record
   → Return TwiML: greet caller + start Gather (speech input)

2. Caller speaks → Twilio STT → POST /webhooks/twilio/gather
   with SpeechResult parameter
   → Detect intent (question / book / reschedule / cancel)
   → Run RAG pipeline if question
   → Generate text response
   → ElevenLabs TTS → cache audio
   → Return TwiML: Play audio + Gather again (loop)

3. If intent = book:
   → Ask for name (if unknown)
   → Ask for email (if unknown)
   → Ask for preferred slot
   → Confirm details verbally
   → On confirmation: Book appointment
   → Send SMS confirmation
   → Announce booking confirmed

4. POST /webhooks/twilio/status (call ended)
   → Finalize call record (duration, status)
   → Calculate total cost
   → Create/update Lead record
   → Broadcast via WebSocket to dashboard

---

## VOICE AI SYSTEM PROMPT

You are a professional IT services receptionist assistant named "Alex".
You speak clearly and concisely — responses must be under 2 sentences
for voice delivery.

STRICT RULES:
- Answer ONLY using the provided context from the knowledge base
- If the context does not contain the answer, say exactly:
  "I don't have that specific information. Would you like me to
  connect you with our team?"
- NEVER fabricate pricing, service details, or availability
- NEVER use markdown formatting — plain speech only
- Collect caller's full name and email before booking any appointment
- Always confirm appointment details before saving:
  "Just to confirm — [name], [service] on [date] at [time].
  Is that correct?"
- After every booking, offer SMS confirmation

INTENT DETECTION:
- Question about services/pricing → RAG query → answer from context
- "book", "schedule", "appointment" → appointment booking flow
- "reschedule", "change my appointment" → reschedule flow
- "cancel" → cancellation flow
- Unclear → ask for clarification once, then offer to transfer

TONE: Professional, warm, efficient. Never robotic.

---

## API ENDPOINTS

### Public
POST /auth/signup
POST /auth/signin
GET  /health
GET  /metrics
GET  /swagger/*

### Twilio Webhooks (Twilio signature validation — no JWT)
POST /webhooks/twilio/voice          # Inbound call handler
POST /webhooks/twilio/gather         # Speech/DTMF input handler
POST /webhooks/twilio/status         # Call status updates
POST /webhooks/twilio/recording      # Recording callback

### Protected (JWT required)

# Auth
GET  /users/me
PUT  /users/me
POST /users/me/password

# Calls
GET  /calls                          # List with filters + pagination
GET  /calls/:id                      # Full call detail + transcript
GET  /calls/:id/rag-queries          # RAG queries for this call
GET  /calls/stats                    # Cost analytics

# Appointments
POST   /appointments
GET    /appointments
GET    /appointments/:id
PUT    /appointments/:id             # Reschedule
DELETE /appointments/:id             # Cancel
GET    /appointments/availability    # Free slots from Google Calendar

# Leads
GET /leads
GET /leads/:id
PUT /leads/:id

# Knowledge Base
POST   /knowledge/documents
GET    /knowledge/documents
GET    /knowledge/documents/:id
DELETE /knowledge/documents/:id
POST   /knowledge/documents/:id/reindex
POST   /knowledge/search             # Test semantic search (admin)

# Dashboard
GET /dashboard/stats
GET /dashboard/costs
GET /dashboard/volume

# System Settings (admin only)
GET /settings
PUT /settings

# WebSocket
GET /ws

---

## DATABASE MIGRATIONS (PostgreSQL)

-- 001_users.sql
CREATE TABLE users (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
email VARCHAR(255) UNIQUE NOT NULL,
password_hash VARCHAR(255) NOT NULL,
role VARCHAR(50) NOT NULL DEFAULT 'user',
is_blocked BOOLEAN NOT NULL DEFAULT false,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 002_knowledge_documents.sql
CREATE TABLE knowledge_documents (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
filename VARCHAR(500) NOT NULL,
mime_type VARCHAR(100) NOT NULL,
file_path VARCHAR(1000) NOT NULL,
chunk_count INT NOT NULL DEFAULT 0,
status VARCHAR(50) NOT NULL DEFAULT 'pending',
indexed_at TIMESTAMPTZ,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 003_inbound_calls.sql
CREATE TABLE inbound_calls (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
twilio_call_sid VARCHAR(255) UNIQUE NOT NULL,
caller_phone VARCHAR(50) NOT NULL,
status VARCHAR(50) NOT NULL DEFAULT 'ringing',
transcript JSONB NOT NULL DEFAULT '[]',
rag_queries JSONB NOT NULL DEFAULT '[]',
duration_seconds INT NOT NULL DEFAULT 0,
twilio_cost_usd DECIMAL(10,6) NOT NULL DEFAULT 0,
llm_cost_usd DECIMAL(10,6) NOT NULL DEFAULT 0,
total_cost_usd DECIMAL(10,6) NOT NULL DEFAULT 0,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ended_at TIMESTAMPTZ
);

-- 004_appointments.sql
CREATE TABLE appointments (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
call_id UUID REFERENCES inbound_calls(id),
caller_phone VARCHAR(50) NOT NULL,
caller_name VARCHAR(255) NOT NULL,
caller_email VARCHAR(255) NOT NULL,
service_type VARCHAR(255) NOT NULL,
scheduled_at TIMESTAMPTZ NOT NULL,
duration_mins INT NOT NULL DEFAULT 60,
status VARCHAR(50) NOT NULL DEFAULT 'pending',
google_event_id VARCHAR(255),
sms_sent_at TIMESTAMPTZ,
notes TEXT,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 005_leads.sql
CREATE TABLE leads (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
call_id UUID REFERENCES inbound_calls(id),
phone VARCHAR(50) UNIQUE NOT NULL,
name VARCHAR(255),
email VARCHAR(255),
status VARCHAR(50) NOT NULL DEFAULT 'new',
notes TEXT,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 006_sms_logs.sql
CREATE TABLE sms_logs (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
call_id UUID REFERENCES inbound_calls(id),
to_phone VARCHAR(50) NOT NULL,
message TEXT NOT NULL,
twilio_sid VARCHAR(255),
status VARCHAR(50) NOT NULL DEFAULT 'queued',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 007_audio_cache.sql
CREATE TABLE audio_cache (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
hash VARCHAR(64) UNIQUE NOT NULL,
voice_id VARCHAR(255) NOT NULL,
file_path VARCHAR(1000) NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 008_system_settings.sql
CREATE TABLE system_settings (
id INT PRIMARY KEY DEFAULT 1,
default_llm_provider VARCHAR(100) NOT NULL DEFAULT 'gemini',
default_voice_id VARCHAR(255) NOT NULL DEFAULT 'rachel',
top_k INT NOT NULL DEFAULT 5,
max_call_duration_secs INT NOT NULL DEFAULT 300,
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO system_settings DEFAULT VALUES;

---

## ENVIRONMENT VARIABLES (.env.example)

# Server
PORT=8080
ENV=development
JWT_SECRET=your-secret-here
JWT_EXPIRY_HOURS=24

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=voice_ai
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Weaviate
WEAVIATE_HOST=localhost
WEAVIATE_PORT=8081
WEAVIATE_SCHEME=http

# LLM Providers
GEMINI_API_KEY=
CLAUDE_API_KEY=
OPENAI_API_KEY=
GLM_API_KEY=
MISTRAL_API_KEY=
DEEPSEEK_API_KEY=

# Voice
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_PHONE_NUMBER=
ELEVENLABS_API_KEY=
DEEPGRAM_API_KEY=

# Google Calendar
GOOGLE_CALENDAR_CREDENTIALS_JSON=
GOOGLE_CALENDAR_ID=

# Storage
UPLOAD_PATH=./uploads
AUDIO_CACHE_PATH=./audio_cache

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000

---

## DOCKER COMPOSE SERVICES

services:
postgres:
image: postgres:16-alpine
environment:
POSTGRES_DB: voice_ai
POSTGRES_USER: postgres
POSTGRES_PASSWORD: postgres
ports:
- "5432:5432"
volumes:
- postgres_data:/var/lib/postgresql/data

redis:
image: redis:7-alpine
ports:
- "6379:6379"

weaviate:
image: semitechnologies/weaviate:latest
ports:
- "8081:8080"
environment:
QUERY_DEFAULTS_LIMIT: 25
AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: "true"
DEFAULT_VECTORIZER_MODULE: none
CLUSTER_HOSTNAME: node1

migrations:
build:
context: ./backend
target: migrations
depends_on:
- postgres

backend:
build:
context: ./backend
ports:
- "8080:8080"
depends_on:
- postgres
- redis
- weaviate
env_file:
- .env

frontend:
build:
context: ./frontend
ports:
- "3000:3000"

nginx:
image: nginx:alpine
ports:
- "80:80"
volumes:
- ./nginx/nginx.conf:/etc/nginx/nginx.conf
depends_on:
- backend
- frontend

---

## MAKEFILE COMMANDS

# Root
help          # Show all commands
dev           # Start everything
infra-up      # Start postgres + redis + weaviate
infra-down    # Stop infrastructure
stop          # Stop all services
health        # Check all services health
db-shell      # Open PostgreSQL shell
db-seed       # Seed test data
logs          # Tail all logs

# Backend
run           # Run Go server
build         # Build binary
test          # Run tests
swagger       # Regenerate Swagger docs
fmt           # Format code
lint          # Run golangci-lint
migrate-up    # Run pending migrations
migrate-down  # Rollback N=1 migration
db-reset      # Full reset (WARNING: deletes data)

---

## FRONTEND PAGES (Nuxt 3)

/                    # Login page
/dashboard           # Overview: calls today, bookings, leads, costs
/calls               # Call history table + transcript drawer
/calls/:id           # Call detail: full transcript + RAG queries
/appointments        # Calendar view + list + status management
/leads               # Lead pipeline (kanban or table)
/knowledge-base      # Document upload + indexing status
/settings            # System settings (LLM provider, voice, top-k)

---

## CODE QUALITY RULES

- Hexagonal architecture — domain never imports infrastructure or UI
- No ORM — use sqlx with explicit, readable SQL queries
- Port interfaces use CRUD vocabulary: Create, Find, FindBy*, List, Update, Delete
- No Get/Set method prefixes — idiomatic Go naming
- Context propagation — every method accepts context.Context as first param
- Swagger annotations on every HTTP handler
- Godoc comments on all exported declarations
- Structured logging — zerolog with request_id and user_id fields
- Error wrapping — domain-specific error types with Is/As support
- Graceful shutdown — catch SIGTERM, drain connections
- Twilio webhook validation — verify X-Twilio-Signature on every webhook

---

## BUILD SEQUENCE

Follow this order strictly:

1.  Create full project structure (all folders + empty files)
2.  Set up Docker Compose + Makefile
3.  Configure Viper + zerolog + bootstrap
4.  Write all database migrations
5.  Define all domain models (model/)
6.  Define all port interfaces (port/)
7.  Implement PostgreSQL adapters (infrastructure/postgres/)
8.  Implement Weaviate adapter (infrastructure/weaviate/)
9.  Implement Gemini adapter — LLM + embedding (infrastructure/gemini/)
10. Implement other LLM adapters + LLM router (infrastructure/llm/)
11. Implement ElevenLabs TTS adapter with audio cache (infrastructure/elevenlabs/)
12. Implement Deepgram STT adapter (infrastructure/deepgram/)
13. Implement Twilio adapter — voice + SMS (infrastructure/twilio/)
14. Implement TwiML builder (twiml/)
15. Implement Google Calendar adapter (infrastructure/googlecalendar/)
16. Implement Redis cache adapter (infrastructure/redis/)
17. Implement WebSocket hub (infrastructure/ws/)
18. Implement all domain services (domain/api/)
19. Implement all HTTP handlers (ui/rest/gin/handlers/)
20. Set up router with all routes + middleware (ui/rest/gin/router/)
21. Wire everything in bootstrap/
22. Write main.go with graceful shutdown
23. Build Nuxt 3 frontend — all pages
24. Write Swagger annotations + generate docs
25. Write README + QUICK-START.md
26. Create seed data (10 IT services FAQ documents)
27. Final: run docker-compose up and verify all services healthy

Start now with step 1.