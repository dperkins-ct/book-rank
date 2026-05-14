# BookRank ELO Rating System

## Overview
The BookRank ELO rating system provides personalized book rankings based on pairwise comparisons. Each user maintains their own set of book ratings that adjust dynamically as they compare books.

## Core Components

### 1. ELO Service (`internal/service/elo_service.go`)
- **Starting Rating**: 1500 (standard ELO starting point)
- **K-Factor**: 32 (moderate sensitivity to rating changes)
- **Rating Range**: 0-3000 (with boundary protection)
- **Core Functions**:
  - `CalculateExpectedScore()`: Predicts outcome probability
  - `CalculateNewRating()`: Updates rating after comparison
  - `CalculateELOUpdate()`: Handles full comparison workflow

### 2. Comparison Service (`internal/service/comparison_service.go`)
- **Pairwise Comparisons**: User selects preferred book between two options
- **Real-time Updates**: Ratings recalculated immediately after each comparison
- **Conflict Resolution**: Most recent comparisons weighted in chronological order
- **Onboarding Flow**: Requires 10+ books and 5+ comparisons to complete

### 3. Ranking Service (`internal/service/ranking_service.go`)
- **User-Centric Ratings**: Each user has independent book ratings
- **Statistical Analysis**: Provides ranking stats and comparisons
- **Bulk Operations**: Supports initializing multiple books at once

## API Endpoints

### Comparisons
```
POST   /api/comparisons              # Submit a pairwise comparison
GET    /api/comparisons/pending      # Get books needing comparison
GET    /api/comparisons/history      # View comparison history
GET    /api/comparisons/book/{id}    # Get comparisons for specific book
POST   /api/comparisons/recalculate  # Force rating recalculation
GET    /api/comparisons/onboarding-status # Check onboarding completion
```

### Rankings
```
GET    /api/rankings/me              # Get current user's rankings
GET    /api/rankings/user/{userId}   # Get specific user's rankings (privacy controlled)
GET    /api/rankings/top             # Get highest-rated books
GET    /api/rankings/stats           # Get ranking statistics
GET    /api/rankings/position/{id}   # Get book's position in rankings
GET    /api/rankings/compare         # Compare two books' ratings
POST   /api/rankings/initialize      # Initialize rating for new book(s)
```

## Database Models

### Comparison
```go
type Comparison struct {
    ID         uint
    UserID     uint
    BookAID    uint
    BookBID    uint
    Preference ComparisonPreference // "book_a", "book_b", "tie"
    CreatedAt  time.Time
}
```

### Ranking
```go
type Ranking struct {
    UserID    uint      // Composite primary key
    BookID    uint      // Composite primary key
    Score     int       // ELO rating (0-3000)
    UpdatedAt time.Time
}
```

## ELO Algorithm Details

### Expected Score Calculation
```
E_A = 1 / (1 + 10^((R_B - R_A)/400))
```
Where:
- `E_A` = Expected score for book A
- `R_A` = Current rating of book A  
- `R_B` = Current rating of book B

### Rating Update Formula
```
R'_A = R_A + K * (S_A - E_A)
```
Where:
- `R'_A` = New rating for book A
- `K` = K-factor (32)
- `S_A` = Actual score (1.0 for win, 0.5 for tie, 0.0 for loss)

### Comparison Outcomes
- **Book A Preferred**: A gets 1.0, B gets 0.0
- **Book B Preferred**: A gets 0.0, B gets 1.0  
- **Tie**: Both books get 0.5

## User Onboarding Flow

1. **Add Books**: User must add minimum 10 books to library
2. **Initial Comparisons**: System presents book pairs for comparison
3. **Rating Establishment**: After 5+ comparisons, onboarding is complete
4. **Ongoing Comparisons**: System suggests pending comparisons

## Usage Examples

### Submit a Comparison
```bash
curl -X POST /api/comparisons \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "book_a_id": 1,
    "book_b_id": 2, 
    "preference": "book_a"
  }'
```

### Get Pending Comparisons
```bash
curl -X GET /api/comparisons/pending?limit=5 \
  -H "Authorization: Bearer <token>"
```

### View Rankings
```bash
curl -X GET /api/rankings/me \
  -H "Authorization: Bearer <token>"
```

## Testing

Run ELO system tests:
```bash
go test ./internal/service -v -run "TestELO"
```

Test coverage includes:
- ELO calculation accuracy
- Rating boundary protection  
- Preference conversion logic
- Edge cases and error handling

## Performance Considerations

- **Indexes**: Composite indexes on `(user_id, score)` for fast ranking queries
- **Batch Operations**: Support for bulk rating initialization
- **Caching**: Consider caching frequently accessed rankings
- **Rate Limiting**: API rate limits prevent abuse

## Future Enhancements

- **Advanced K-Factors**: Dynamic K-factor based on rating confidence
- **Temporal Decay**: Reduce weight of very old comparisons
- **Genre-Specific Ratings**: Separate ratings per genre
- **Social Features**: Compare rankings with friends