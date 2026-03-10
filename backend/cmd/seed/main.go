// Package main is the entry point for the database seeding tool.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/configuration"
	"phone-call-receptionist/backend/pkg/helpers"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Msg("Seeding database...")

	cfg := configuration.Config
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := seedAdminUser(ctx, db, &logger); err != nil {
		logger.Fatal().Err(err).Msg("Failed to seed admin user")
	}

	if err := seedKnowledgeDocuments(ctx, db, &logger); err != nil {
		logger.Fatal().Err(err).Msg("Failed to seed knowledge documents")
	}

	logger.Info().Msg("Seeding completed successfully")
}

func seedAdminUser(ctx context.Context, db *sqlx.DB, logger *zerolog.Logger) error {
	var count int
	if err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE role = 'super_admin'"); err != nil {
		return fmt.Errorf("failed to check admin user: %w", err)
	}
	if count > 0 {
		logger.Info().Msg("Admin user already exists, skipping")
		return nil
	}

	hash, err := helpers.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4)",
		uuid.New().String(), "admin@example.com", hash, "super_admin",
	)
	if err != nil {
		return fmt.Errorf("failed to insert admin user: %w", err)
	}

	logger.Info().Msg("Created default admin user (admin@example.com / admin123)")
	return nil
}

func seedKnowledgeDocuments(ctx context.Context, db *sqlx.DB, logger *zerolog.Logger) error {
	var count int
	if err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM knowledge_documents"); err != nil {
		return fmt.Errorf("failed to check knowledge documents: %w", err)
	}
	if count > 0 {
		logger.Info().Msg("Knowledge documents already exist, skipping")
		return nil
	}

	docs := []struct {
		filename string
		content  string
	}{
		{"business-hours.txt", "Our office is open Monday through Friday, 8:00 AM to 6:00 PM Eastern Time. We are closed on weekends and major holidays including New Year's Day, Memorial Day, Independence Day, Labor Day, Thanksgiving, and Christmas."},
		{"services-overview.txt", "We provide comprehensive IT services including: managed IT support, network infrastructure design and maintenance, cybersecurity solutions, cloud migration and management, data backup and disaster recovery, VoIP phone systems, and IT consulting for small to mid-size businesses."},
		{"pricing-plans.txt", "We offer three support plans: Basic ($499/month) covers remote helpdesk support for up to 10 users. Professional ($999/month) includes on-site support, network monitoring, and covers up to 25 users. Enterprise (custom pricing) provides 24/7 support, dedicated account manager, and unlimited users."},
		{"emergency-support.txt", "For critical IT emergencies outside business hours, call our emergency hotline. Critical issues include: complete network outages, active security breaches, server failures affecting all users, and ransomware attacks. Response time for critical issues is within 1 hour."},
		{"onboarding-process.txt", "New client onboarding takes 2-3 business days. It includes: initial network assessment, inventory of all hardware and software assets, security audit, setup of remote monitoring tools, creation of documentation, and a kickoff meeting with your dedicated support team."},
		{"security-services.txt", "Our cybersecurity services include: endpoint protection deployment, firewall management, email security and spam filtering, security awareness training for employees, vulnerability assessments, penetration testing, and compliance audits for HIPAA, PCI-DSS, and SOC 2."},
		{"cloud-services.txt", "We specialize in Microsoft 365 and Google Workspace migrations, Azure and AWS cloud infrastructure, hybrid cloud solutions, and cloud backup services. Migration projects typically take 1-4 weeks depending on the size of the organization."},
		{"backup-recovery.txt", "Our backup solution includes automated daily backups with 30-day retention, off-site replication, quarterly disaster recovery testing, and guaranteed recovery time objectives (RTO) of 4 hours for critical systems. We support both on-premises and cloud-based backup targets."},
		{"hardware-procurement.txt", "We handle hardware procurement for our managed clients at competitive prices. This includes workstations, laptops, servers, networking equipment, printers, and peripherals. We offer lifecycle management, warranty tracking, and recycling of decommissioned equipment."},
		{"sla-terms.txt", "Our Service Level Agreement guarantees: 99.9% network uptime for managed clients, 15-minute response time for critical tickets, 1-hour response for high priority, 4-hour response for medium, and next-business-day for low priority. Monthly SLA compliance reports are provided to all clients."},
	}

	for _, doc := range docs {
		_, err := db.ExecContext(ctx,
			"INSERT INTO knowledge_documents (id, filename, mime_type, file_path, chunk_count, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())",
			uuid.New().String(), doc.filename, "text/plain", "/seeds/"+doc.filename, 1, "pending",
		)
		if err != nil {
			return fmt.Errorf("failed to insert document %s: %w", doc.filename, err)
		}
		logger.Info().Str("filename", doc.filename).Msg("Seeded knowledge document")
	}

	// Write seed files to disk
	if err := os.MkdirAll("seeds", 0755); err != nil {
		return fmt.Errorf("failed to create seeds directory: %w", err)
	}
	for _, doc := range docs {
		if err := os.WriteFile("seeds/"+doc.filename, []byte(doc.content), 0644); err != nil {
			return fmt.Errorf("failed to write seed file %s: %w", doc.filename, err)
		}
	}

	logger.Info().Int("count", len(docs)).Msg("Seeded knowledge documents")
	return nil
}
