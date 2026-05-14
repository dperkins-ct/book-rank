# BookRank Database Models

This directory contains all GORM model definitions for the BookRank application database layer.

## Overview

The database layer implements a complete ELO-based book ranking system with the following core entities:

- **Users**: Application users with authentication
- **Books**: Book catalog with metadata
- **Rankings**: ELO-based scoring for user book preferences  
- **Comparisons**: Pairwise book comparisons that drive ELO calculations
- **Friendships**: Social connections between users
- **BookMetadata**: External metadata from APIs (OpenLibrary, Google Books)

## Architecture

### Database Technology
- **Primary Database**: PostgreSQL 16
- **ORM**: GORM v1.25+
- **Connection Pool**: Max 25 open connections, Max 5 idle connections
- **Query Timeout**: 30 seconds for complex operations
- **Soft Deletes**: Enabled for user data preservation

### Key Features
- Auto-migration on application startup
- Custom performance indexes
- Composite primary keys where appropriate
- Foreign key constraints with cascading
- JSON fields for flexible metadata storage
- Proper timestamp handling with UTC
- Comprehensive validation tags

## Models

### User Model
```go
type User struct {
    ID           uint           `gorm:"primaryKey"`
    Username     string         `gorm:"unique;not null;size:20;check:length(username) >= 3"`
    PasswordHash string         `gorm:"not null;size:255"`
    CreatedAt    time.Time      `gorm:"not null"`
    UpdatedAt    time.Time      `gorm:"not null"`
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}
```

**Constraints:**
- Username: 3-20 characters, unique
- Password validation handled in service layer (8+ chars, 1 special char)
- Soft deletes enabled

### Book Model
```go
type Book struct {
    ID              uint           `gorm:"primaryKey"`
    Title           string         `gorm:"not null;size:255;index"`
    Author          string         `gorm:"not null;size:255;index"`
    Genre           string         `gorm:"not null;size:100;index"`
    PublicationDate *time.Time     `gorm:"index"`
    Description     string         `gorm:"type:text"`
    CreatedBy       uint           `gorm:"not null;index"`
    CreatedAt       time.Time      `gorm:"not null"`
    UpdatedAt       time.Time      `gorm:"not null"`
}
```

**Constraints:**
- Title and Author are required
- CreatedBy references Users.ID
- Indexes on searchable fields

### Ranking Model (ELO Scores)
```go
type Ranking struct {
    UserID    uint      `gorm:"primaryKey;not null"`
    BookID    uint      `gorm:"primaryKey;not null"`
    Score     int       `gorm:"not null;default:1500;check:score >= 0 AND score <= 3000"`
    UpdatedAt time.Time `gorm:"not null"`
}
```

**Constraints:**
- Composite primary key (UserID, BookID)
- ELO scores default to 1500
- Score range: 0-3000
- Real-time updates after comparisons

### Comparison Model
```go
type Comparison struct {
    ID         uint                 `gorm:"primaryKey"`
    UserID     uint                 `gorm:"not null;index"`
    BookAID    uint                 `gorm:"not null;index"`
    BookBID    uint                 `gorm:"not null;index"`
    Preference ComparisonPreference `gorm:"not null;type:varchar(10)"`
    CreatedAt  time.Time            `gorm:"not null;index"`
}
```

**Preferences:**
- `book_a`: User prefers Book A
- `book_b`: User prefers Book B  
- `tie`: No preference/equal ranking

**Constraints:**
- BookAID ≠ BookBID (enforced in BeforeCreate hook)
- Most recent comparison weighted higher in ELO calculations

### Friendship Model
```go
type Friendship struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null;index"`
    FriendID  uint      `gorm:"not null;index"`
    CreatedAt time.Time `gorm:"not null;index"`
}
```

**Constraints:**
- UserID ≠ FriendID (enforced in BeforeCreate hook)
- Unique constraint on (UserID, FriendID)
- Bidirectional relationships supported

### BookMetadata Model
```go
type BookMetadata struct {
    ID             uint           `gorm:"primaryKey"`
    BookID         uint           `gorm:"not null;index"`
    ExternalID     string         `gorm:"size:255;index"`
    Source         MetadataSource `gorm:"not null;type:varchar(50);index"`
    AdditionalData JSON           `gorm:"type:jsonb"`
    CreatedAt      time.Time      `gorm:"not null"`
    UpdatedAt      time.Time      `gorm:"not null"`
}
```

**Sources:**
- `openlibrary`: OpenLibrary API
- `googlebooks`: Google Books API
- `manual`: Manually entered metadata

**JSON Data Examples:**
```json
{
  "isbn": "9780123456789",
  "cover_url": "https://example.com/cover.jpg",
  "page_count": 250,
  "categories": ["Fiction", "Adventure"],
  "ratings": {
    "average": 4.5,
    "count": 123
  }
}
```

## Relationships

### User Relationships
- `Books`: One-to-many (books created by user)
- `Rankings`: One-to-many (user's book scores)
- `Comparisons`: One-to-many (user's comparisons)
- `Friendships`: One-to-many (initiated friendships)
- `FriendOf`: One-to-many (received friendships)

### Book Relationships
- `Creator`: Many-to-one (User)
- `Rankings`: One-to-many (user scores for book)
- `Metadata`: One-to-many (external metadata)
- `ComparisonsAsA`/`ComparisonsAsB`: One-to-many (comparisons involving book)

## Performance Optimizations

### Indexes
```sql
-- User rankings by score
CREATE INDEX idx_rankings_user_score ON rankings(user_id, score DESC);

-- Book rankings by score  
CREATE INDEX idx_rankings_book_score ON rankings(book_id, score DESC);

-- User comparison history
CREATE INDEX idx_comparisons_user_created ON comparisons(user_id, created_at DESC);

-- Book comparison lookup
CREATE INDEX idx_comparisons_books ON comparisons(book_a_id, book_b_id);

-- Unique friendships
CREATE UNIQUE INDEX idx_friendships_unique ON friendships(user_id, friend_id);

-- Metadata source lookup
CREATE UNIQUE INDEX idx_book_metadata_source_external 
ON book_metadata(source, external_id) WHERE external_id IS NOT NULL;

-- Book search
CREATE INDEX idx_books_title_author ON books(title, author);
```

### Connection Pool
- Max Open Connections: 25
- Max Idle Connections: 5  
- Connection Max Lifetime: 1 hour
- Connection Max Idle Time: 10 minutes
- Query Timeout: 30 seconds

## Usage

### Basic Setup
```go
import (
    "bookrank/internal/database"
    "bookrank/internal/models"
)

// Initialize database
config := database.NewConfig()
db, err := database.Initialize(config)
if err != nil {
    log.Fatal(err)
}

// Auto-migrate schemas
if err := models.AutoMigrate(db); err != nil {
    log.Fatal(err)
}
```

### Creating Records
```go
// Create user
user := models.User{
    Username:     "bookworm",
    PasswordHash: "hashed_password",
}
db.Create(&user)

// Create book
book := models.Book{
    Title:     "The Great Gatsby",
    Author:    "F. Scott Fitzgerald", 
    Genre:     "Fiction",
    CreatedBy: user.ID,
}
db.Create(&book)

// Create ranking (starts at ELO 1500)
ranking := models.Ranking{
    UserID: user.ID,
    BookID: book.ID,
}
db.Create(&ranking)
```

### Querying with Relationships
```go
// Get user with books
var user models.User
db.Preload("Books").First(&user, userID)

// Get book with rankings
var book models.Book  
db.Preload("Rankings.User").First(&book, bookID)

// Get user's top ranked books
var rankings []models.Ranking
db.Preload("Book").Where("user_id = ?", userID).
   Order("score DESC").Limit(10).Find(&rankings)
```

## Testing

All models include comprehensive unit tests with table-driven test patterns:

```bash
# Run all model tests
go test ./internal/models/...

# Run with coverage
go test -cover ./internal/models/...

# Run specific model tests  
go test ./internal/models/ -run TestUserModel
```

### Test Database
Tests use SQLite in-memory database for fast, isolated testing without external dependencies.

## Migration

The database supports versioned migrations:

```go
// Run all pending migrations
database.RunMigrations(db)

// Check migration status
status, err := database.GetMigrationStatus(db)

// Rollback specific migration
database.RollbackMigration(db, "001_initial_schema")
```

## Environment Configuration

See `.env.example` for all configuration options:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=bookrank

# Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

## Development

### Docker Setup
```bash
# Start PostgreSQL only
docker-compose up postgres-dev

# Start full stack
docker-compose --profile full up

# Start with adminer for database management
docker-compose up postgres-dev adminer
```

### Local Development
```bash
# Install dependencies
go mod tidy

# Run tests
go test ./internal/models/...

# Run with PostgreSQL
export DB_HOST=localhost
go run cmd/server/main.go
```