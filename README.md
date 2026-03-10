# Phone Call Receptionist

Voice AI Receptionist for IT services firms. Handles inbound calls with speech recognition, RAG-powered knowledge base answers, appointment booking, and lead capture.

## Architecture

- **Backend**: Go 1.24, Gin, Hexagonal Architecture (Ports & Adapters)
- **Frontend**: Nuxt 3, Vue 3, Tailwind CSS
- **Database**: PostgreSQL 16, Redis 7, Weaviate (vector search)
- **Voice**: Twilio (telephony), ElevenLabs (TTS), Deepgram (STT)
- **LLM**: Multi-provider with failover (Gemini, Claude, OpenAI, GLM, Mistral, DeepSeek)
- **Calendar**: Google Calendar integration for appointment booking

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.24+ (for local development)
- Node.js 20+ (for frontend development)

### 1. Configure environment

```bash
cp .env.example .env
# Edit .env with your API keys
```

### 2. Start infrastructure

```bash
make infra-up
```

This starts PostgreSQL, Redis, and Weaviate containers.

### 3. Run migrations

```bash
cd backend && make migrate-up
```

### 4. Seed database

```bash
cd backend && make db-seed
```

Creates a default admin user (`admin@example.com` / `admin123`) and 10 IT FAQ knowledge documents.

### 5. Run backend

```bash
cd backend && make run
```

Backend starts at `http://localhost:8080`.

### 6. Run frontend

```bash
cd frontend && npm install && npm run dev
```

Frontend starts at `http://localhost:3000`.

### Full stack with Docker

```bash
make dev
```

Starts all services (backend, frontend, postgres, redis, weaviate, nginx).

## API Documentation

Swagger UI is available at `http://localhost:8080/swagger/index.html` after starting the backend.

Generate/update swagger docs:

```bash
cd backend && make swagger
```

## Project Structure

```
phone-call-receptionist/
├── backend/
│   ├── cmd/
│   │   ├── server/         # Main server entry point
│   │   └── seed/           # Database seeding tool
│   ├── internal/
│   │   ├── bootstrap/      # Dependency injection wiring
│   │   ├── configuration/  # Environment-based config
│   │   ├── domain/
│   │   │   ├── model/      # Pure domain entities
│   │   │   ├── port/       # Interface contracts
│   │   │   ├── api/        # Business logic services
│   │   │   └── errors/     # Domain error types
│   │   ├── infrastructure/ # External adapters (postgres, redis, llm, etc.)
│   │   └── ui/rest/gin/    # HTTP handlers, middleware, router
│   ├── pkg/                # Shared packages (DTOs, helpers, database)
│   └── migrations/         # SQL migration files
├── frontend/src/           # Nuxt 3 application
├── nginx/                  # Reverse proxy config
├── docker-compose.yml
└── Makefile
```

## Key Features

- **Inbound call handling**: Twilio webhooks receive calls, Gather speech input
- **RAG knowledge base**: Upload documents, chunk and embed them, search with LLM-generated answers
- **Appointment booking**: Multi-turn conversation state machine, Google Calendar integration
- **Lead capture**: Automatic lead creation from caller information
- **Real-time updates**: WebSocket events for call status, new leads
- **Multi-LLM failover**: Ordered provider list with automatic fallback
- **Admin dashboard**: Call history, cost analytics, lead management

## Twilio Webhook Setup

Configure your Twilio phone number webhooks:

| Event | URL | Method |
|-------|-----|--------|
| Voice | `https://your-domain/api/webhooks/twilio/voice` | POST |
| Status | `https://your-domain/api/webhooks/twilio/status` | POST |

## Environment Variables

See `.env.example` for all available configuration options. At minimum you need:

- `DB_*` — PostgreSQL connection (defaults work with docker-compose)
- `JWT_SECRET` — Secret key for JWT token signing
- `GEMINI_API_KEY` or at least one LLM provider key
- `TWILIO_*` — Twilio credentials for voice/SMS
- `ELEVENLABS_API_KEY` — Text-to-speech
