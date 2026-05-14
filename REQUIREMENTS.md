# Book Ranking & Recommendation Application - Requirements

## Name
The name of the site should be BookRank

## Project Overview
A web application where users can register books they've read, rank them using a 1-10 scoring system that adjusts based on pairwise comparisons, and receive personalized book recommendations.

## Technology Stack
- **Backend**: Go
- **Frontend**: React
- **Database**: PostgreSQL
- **Hosting**: AWS
- **Scale**: ~100 users (demo application)

## Technical Requirements

### Testing Standards
- All files must have a corresponding _test.go file for unit tests
- Must use table driven unit tests like:
```go
	tests := map[string]struct {
		running  bool
		interval time.Duration
		wantErr  bool
	}
```
- Must achieve 80% code coverage

### Project Structure (Standard Go Layout)
```
/cmd
  /server          # main.go application entry point
/internal
  /api            # HTTP handlers and middleware
  /service        # Business logic layer
  /repository     # Data access layer
  /models         # Domain entities and structs
  /auth           # JWT authentication logic
/pkg              # Reusable packages
/migrations       # Database migration files
/docker-compose.yml
/Dockerfile
```

### Backend Implementation Details
- **HTTP Framework**: Standard `net/http` with gorilla/mux for routing
- **ORM**: GORM with PostgreSQL driver
- **Authentication**: JWT tokens with 24-hour expiration
- **Logging**: Standard `slog` package
- **Error Handling**: Standardized JSON error responses
- **Configuration**: Environment-based config management

### Database Configuration (GORM)
- Auto-migrate schema on application startup
- Soft deletes enabled for user data preservation
- Connection pooling: max 25 open connections, max 5 idle
- Query timeout: 30 seconds for complex operations

### ELO Rating System Implementation
- **User Onboarding**: New users must add 10 books minimum before accessing features
- **Initial Setup**: Users perform pairwise comparisons on their initial 10 books
- **Starting ELO**: New books start at 1500 rating
- **K-Factor**: 32 (moderate rating change sensitivity)
- **Conflict Resolution**: Most recent comparison weighted higher in rating calculations
- **Real-time Updates**: Ratings recalculated immediately after each comparison
- **User-Centric**: Book popularity does not influence personal ratings

### API Design Specifications
- **Architecture**: RESTful API with JSON responses
- **Pagination**: Default 20 items per page, max 100
- **Sorting**: Support for rating, date_added, alphabetical sorting
- **Filtering**: By genre, author, rating range (1-10)
- **Rate Limiting**: 100 requests per minute per user
- **Health Endpoint**: `/health` for monitoring

### Authentication Flow
- **Registration**: Username (3-20 chars), password (any non-empty)
- **Login**: Returns JWT with user ID and expiration
- **Token Refresh**: 24-hour token lifetime, sliding session
- **Password Security**: bcrypt hashing with cost factor 12

### Development & Deployment
- **Local Development**: Docker Compose with PostgreSQL container
- **Hot Reload**: Air for Go development auto-restart
- **CI/CD**: GitHub Actions for testing and Docker builds
- **Containerization**: Multi-stage Docker builds for production
- **Environment Configs**: Separate configs for dev/staging/production 

## Core Features

### 1. User Management
- Simple authentication with username and password (no complex registration flow)
- **Onboarding Process**: New users must add 10 books and perform initial pairwise comparisons before accessing main features
- User profiles displaying:
  - Reading statistics (books read, average rating, etc.)
  - Reading preferences derived from ranking patterns
  - Friend connections (simple friend system)

### 2. Book Management
- **Manual Book Registration**: Users can add books with:
  - Title (required)
  - Author (required) 
  - Genre (required)
  - Publication Date
  - Description/Summary
- **Book Information Display**:
  - All registered book details
  - Personal ranking (1-10 scale)
  - Average user ranking across platform
  - Additional metadata from free/open APIs (when available)

### 3. Ranking System
- **1-10 Scoring Scale**: Each book receives a numerical score
- **Comparison-Based Adjustments**: 
  - Users can compare two books (Book A vs Book B)
  - System adjusts scores based on comparative rankings
  - Algorithm maintains consistency across all user's book rankings
- **Personal vs Community Rankings**: Display both individual and platform average scores

### 4. Recommendation Engine
Primary recommendation drivers (in order of priority):
1. **Collaborative Filtering**: Recommendations based on users with similar ranking patterns
2. **Genre/Author Preferences**: Suggestions based on highly-ranked books in user's preferred categories
3. **External Metadata**: Utilize book metadata and similarity data from open APIs

### 5. Social Features
- **Simple Friends System**:
  - Add/remove friends by username
  - View friends' book rankings and recommendations
  - No complex social features (no posts, comments, etc.)

## Technical Requirements

### Backend (Go)
- RESTful API design
- User authentication and session management
- Book CRUD operations
- Ranking algorithm implementation
- Recommendation engine
- Friend management system
- Integration with external book APIs (OpenLibrary, Google Books API, etc.)

### Frontend (React)
- Responsive web design
- User dashboard with reading statistics
- Book registration and management interface
- Ranking comparison interface
- Recommendation display
- Friends management
- Clean, intuitive UI for book comparisons

### Database Schema (PostgreSQL)
Key entities:
- **Users**: id, username, password_hash, created_at, updated_at
- **Books**: id, title, author, genre, publication_date, description, created_by, created_at
- **Rankings**: user_id, book_id, score, updated_at
- **Comparisons**: user_id, book_a_id, book_b_id, preference, created_at
- **Friendships**: user_id, friend_id, created_at
- **Book_Metadata**: book_id, external_id, source, additional_data (JSON)

### External Integrations
- **OpenLibrary API**: For additional book metadata and cover images
- **Google Books API**: Fallback for book information and descriptions
- **Book recommendation data sources**: For similarity and popularity metrics

## User Stories

### Core User Flows
1. **New User Registration**: Simple username/password signup
2. **Onboarding Setup**: 
   - Add 10 books with manual entry and form validation
   - Perform pairwise comparisons on initial book set
   - System generates initial ELO ratings from comparisons
3. **Add Additional Books**: Manual entry of new book details
4. **Compare Books**: Side-by-side comparison interface with preference selection
5. **View Recommendations**: Personalized book suggestions with reasoning
6. **Add Friends**: Search and connect with other users
7. **Browse Rankings**: View personal library with scores and community averages

### Advanced Features
- **Bulk Import**: CSV upload for users with existing book lists
- **Reading Goals**: Set and track annual reading targets
- **Genre Analytics**: Insights into reading patterns and preferences

## Performance & Scalability
- Target: Support 100 concurrent users
- Response time: <500ms for most operations
- Database optimization for ranking calculations
- Caching layer for recommendation engine results

## Security Considerations
- Password hashing (bcrypt)
- Input validation and sanitization
- Rate limiting on API endpoints
- HTTPS encryption
- Basic CSRF protection

## Deployment Architecture (AWS)
- **Application**: EC2 instance or ECS containers
- **Database**: RDS PostgreSQL
- **Static Assets**: S3 + CloudFront
- **Load Balancing**: Application Load Balancer (if needed)
- **Domain & SSL**: Route 53 + Certificate Manager

## Success Metrics
- User engagement: Books added per user, comparisons made per session
- Recommendation quality: Click-through rate on recommended books
- User retention: Weekly/monthly active users
- System performance: API response times, database query efficiency

## Future Enhancements (Out of Scope for MVP)
- Mobile application (React Native)
- Book clubs and group features
- Advanced analytics and insights
- Integration with reading platforms (Kindle, Goodreads)
- Machine learning improvements to recommendation engine
- Book lending/sharing features among friends