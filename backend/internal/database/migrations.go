package database

import (
	"database/sql"
	"fmt"
	"log"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// Migrations holds all database migrations
var Migrations = []Migration{
	{
		Version: 1,
		Name:    "create_users_table",
		Up: `
			CREATE TABLE IF NOT EXISTS users (
				id VARCHAR(8) PRIMARY KEY,
				email VARCHAR(255) UNIQUE NOT NULL,
				name VARCHAR(255) NOT NULL,
				role VARCHAR(20) NOT NULL DEFAULT 'user',
				organization_id VARCHAR(8),
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			
			CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
			CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id);
		`,
		Down: `DROP TABLE IF EXISTS users;`,
	},
	{
		Version: 2,
		Name:    "create_organizations_table",
		Up: `
			CREATE TABLE IF NOT EXISTS organizations (
				id VARCHAR(8) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				description TEXT,
				domain VARCHAR(255),
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			
			CREATE INDEX IF NOT EXISTS idx_organizations_domain ON organizations(domain);
		`,
		Down: `DROP TABLE IF EXISTS organizations;`,
	},
	{
		Version: 3,
		Name:    "create_organization_members_table",
		Up: `
			CREATE TABLE IF NOT EXISTS organization_members (
				id VARCHAR(8) PRIMARY KEY,
				organization_id VARCHAR(8) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
				user_id VARCHAR(8) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				role VARCHAR(20) NOT NULL DEFAULT 'member',
				joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				UNIQUE(organization_id, user_id)
			);
			
			CREATE INDEX IF NOT EXISTS idx_organization_members_org_id ON organization_members(organization_id);
			CREATE INDEX IF NOT EXISTS idx_organization_members_user_id ON organization_members(user_id);
		`,
		Down: `DROP TABLE IF EXISTS organization_members;`,
	},
	{
		Version: 4,
		Name:    "create_invites_table",
		Up: `
			CREATE TABLE IF NOT EXISTS invites (
				id VARCHAR(8) PRIMARY KEY,
				organization_id VARCHAR(8) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
				email VARCHAR(255) NOT NULL,
				invited_by VARCHAR(8) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				role VARCHAR(20) NOT NULL DEFAULT 'member',
				status VARCHAR(20) NOT NULL DEFAULT 'pending',
				expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				accepted_at TIMESTAMP WITH TIME ZONE
			);
			
			CREATE INDEX IF NOT EXISTS idx_invites_organization_id ON invites(organization_id);
			CREATE INDEX IF NOT EXISTS idx_invites_email ON invites(email);
			CREATE INDEX IF NOT EXISTS idx_invites_status ON invites(status);
		`,
		Down: `DROP TABLE IF EXISTS invites;`,
	},
	{
		Version: 5,
		Name:    "create_labs_table",
		Up: `
			CREATE TABLE IF NOT EXISTS labs (
				id VARCHAR(8) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'provisioning',
				owner_id VARCHAR(8) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				started_at TIMESTAMP WITH TIME ZONE NOT NULL,
				ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
				template_id VARCHAR(255),
				service_data JSONB,
				used_services TEXT[],
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			
			CREATE INDEX IF NOT EXISTS idx_labs_owner_id ON labs(owner_id);
			CREATE INDEX IF NOT EXISTS idx_labs_status ON labs(status);
			CREATE INDEX IF NOT EXISTS idx_labs_ends_at ON labs(ends_at);
		`,
		Down: `DROP TABLE IF EXISTS labs;`,
	},
	{
		Version: 6,
		Name:    "create_credentials_table",
		Up: `
			CREATE TABLE IF NOT EXISTS credentials (
				id VARCHAR(8) PRIMARY KEY,
				lab_id VARCHAR(8) NOT NULL REFERENCES labs(id) ON DELETE CASCADE,
				label VARCHAR(255) NOT NULL,
				username VARCHAR(255) NOT NULL,
				password VARCHAR(255) NOT NULL,
				url TEXT,
				expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
				notes TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			
			CREATE INDEX IF NOT EXISTS idx_credentials_lab_id ON credentials(lab_id);
			CREATE INDEX IF NOT EXISTS idx_credentials_expires_at ON credentials(expires_at);
		`,
		Down: `DROP TABLE IF EXISTS credentials;`,
	},
	{
		Version: 7,
		Name:    "create_service_configs_table",
		Up: `
			CREATE TABLE IF NOT EXISTS service_configs (
				id VARCHAR(8) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				type VARCHAR(50) NOT NULL,
				description TEXT,
				logo VARCHAR(255),
				config JSONB NOT NULL DEFAULT '{}',
				is_active BOOLEAN NOT NULL DEFAULT true,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			
			CREATE INDEX IF NOT EXISTS idx_service_configs_type ON service_configs(type);
			CREATE INDEX IF NOT EXISTS idx_service_configs_is_active ON service_configs(is_active);
		`,
		Down: `DROP TABLE IF EXISTS service_configs;`,
	},
	{
		Version: 8,
		Name:    "create_service_limits_table",
		Up: `
			CREATE TABLE IF NOT EXISTS service_limits (
				id VARCHAR(8) PRIMARY KEY,
				service_id VARCHAR(8) NOT NULL REFERENCES service_configs(id) ON DELETE CASCADE,
				max_labs INTEGER NOT NULL DEFAULT 10,
				max_duration INTEGER NOT NULL DEFAULT 480,
				is_active BOOLEAN NOT NULL DEFAULT true,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			
			CREATE INDEX IF NOT EXISTS idx_service_limits_service_id ON service_limits(service_id);
			CREATE INDEX IF NOT EXISTS idx_service_limits_is_active ON service_limits(is_active);
		`,
		Down: `DROP TABLE IF EXISTS service_limits;`,
	},
	{
		Version: 9,
		Name:    "create_migrations_table",
		Up: `
			CREATE TABLE IF NOT EXISTS migrations (
				version INTEGER PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
		`,
		Down: `DROP TABLE IF EXISTS migrations;`,
	},
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB) error {
	// First, ensure the migrations table exists
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range Migrations {
		if _, applied := appliedMigrations[migration.Version]; !applied {
			log.Printf("Running migration %d: %s", migration.Version, migration.Name)

			if err := runMigration(db, migration); err != nil {
				return fmt.Errorf("failed to run migration %d (%s): %w", migration.Version, migration.Name, err)
			}

			log.Printf("Successfully applied migration %d: %s", migration.Version, migration.Name)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	query := `SELECT version FROM migrations ORDER BY version`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// runMigration executes a single migration
func runMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute the migration
	if _, err := tx.Exec(migration.Up); err != nil {
		return err
	}

	// Record the migration as applied
	query := `INSERT INTO migrations (version, name) VALUES ($1, $2)`
	if _, err := tx.Exec(query, migration.Version, migration.Name); err != nil {
		return err
	}

	return tx.Commit()
}
