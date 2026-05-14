package database

import (
	"fmt"
	"log/slog"

	"bookrank/internal/models"
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID          string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// MigrationRecord tracks applied migrations
type MigrationRecord struct {
	ID          string `gorm:"primaryKey"`
	AppliedAt   int64  `gorm:"not null"`
	Description string `gorm:"not null"`
}

// TableName specifies the table name for migration records
func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// RunMigrations runs all pending migrations
func RunMigrations(db *gorm.DB) error {
	// Create migrations table if it doesn't exist
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations := getMigrations()

	for _, migration := range migrations {
		var record MigrationRecord
		result := db.First(&record, "id = ?", migration.ID)

		if result.Error == gorm.ErrRecordNotFound {
			slog.Info("Running migration", "id", migration.ID, "description", migration.Description)

			if err := migration.Up(db); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", migration.ID, err)
			}

			// Record the migration
			record = MigrationRecord{
				ID:          migration.ID,
				AppliedAt:   db.NowFunc().Unix(),
				Description: migration.Description,
			}

			if err := db.Create(&record).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
			}

			slog.Info("Migration completed", "id", migration.ID)
		}
	}

	return nil
}

// RollbackMigration rolls back a specific migration
func RollbackMigration(db *gorm.DB, migrationID string) error {
	migrations := getMigrations()

	var targetMigration *Migration
	for i := range migrations {
		if migrations[i].ID == migrationID {
			targetMigration = &migrations[i]
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %s not found", migrationID)
	}

	var record MigrationRecord
	if err := db.First(&record, "id = ?", migrationID).Error; err != nil {
		return fmt.Errorf("migration %s not applied: %w", migrationID, err)
	}

	slog.Info("Rolling back migration", "id", migrationID, "description", targetMigration.Description)

	if err := targetMigration.Down(db); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", migrationID, err)
	}

	if err := db.Delete(&record).Error; err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", migrationID, err)
	}

	slog.Info("Migration rolled back", "id", migrationID)
	return nil
}

// getMigrations returns all available migrations
func getMigrations() []Migration {
	return []Migration{
		{
			ID:          "001_initial_schema",
			Description: "Create initial database schema",
			Up: func(db *gorm.DB) error {
				return models.AutoMigrate(db)
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(models.AllModels()...)
			},
		},
		{
			ID:          "002_create_indexes",
			Description: "Create performance indexes",
			Up: func(db *gorm.DB) error {
				return models.CreateIndexes(db)
			},
			Down: func(db *gorm.DB) error {
				// Drop custom indexes
				indexes := []string{
					"idx_rankings_user_score",
					"idx_rankings_book_score",
					"idx_comparisons_user_created",
					"idx_comparisons_books",
					"idx_friendships_unique",
					"idx_book_metadata_source_external",
					"idx_books_title_author",
				}

				for _, index := range indexes {
					if err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index)).Error; err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			ID:          "003_make_genre_optional",
			Description: "Make book genre field optional for metadata fetching",
			Up: func(db *gorm.DB) error {
				// Remove NOT NULL constraint from genre column
				return db.Exec("ALTER TABLE books ALTER COLUMN genre DROP NOT NULL").Error
			},
			Down: func(db *gorm.DB) error {
				// Add back NOT NULL constraint (this will fail if there are NULL values)
				return db.Exec("ALTER TABLE books ALTER COLUMN genre SET NOT NULL").Error
			},
		},
	}
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(db *gorm.DB) ([]map[string]interface{}, error) {
	migrations := getMigrations()
	status := make([]map[string]interface{}, len(migrations))

	for i, migration := range migrations {
		var record MigrationRecord
		applied := db.First(&record, "id = ?", migration.ID).Error == nil

		status[i] = map[string]interface{}{
			"id":          migration.ID,
			"description": migration.Description,
			"applied":     applied,
		}

		if applied {
			status[i]["applied_at"] = record.AppliedAt
		}
	}

	return status, nil
}