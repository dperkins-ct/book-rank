# BookRank - Personal Book Rating & Recommendation System

A complete full-stack application for book ranking, comparison, and personalized recommendations built with Go backend and React frontend.

## 🌟 Overview

BookRank helps you discover your next favorite book through intelligent comparisons and personalized recommendations. Compare books side-by-side, build your personal library, and get recommendations based on your reading preferences using an advanced ELO rating system.

## ✨ Features

### 📚 Book Management
- **Personal Library**: Add, edit, and organize your books
- **Metadata Integration**: Automatic book info from OpenLibrary and Google Books
- **Advanced Search**: Filter by author, genre, rating, publication date
- **Rich Details**: Cover images, descriptions, publication info, and statistics

### ⚖️ Intelligent Comparison System
- **Side-by-Side Comparisons**: Interactive book comparison interface
- **ELO Rating System**: Mathematical ranking based on user preferences
- **Comparison History**: Track all your book preferences
- **Smart Pairing**: Intelligent book selection for optimal ranking

### 🎯 Personalized Recommendations
- **AI-Powered Suggestions**: Based on your comparison history and preferences
- **Genre-Based Discovery**: Explore books in specific categories
- **Social Recommendations**: See what your friends are reading
- **Feedback Learning**: Improve recommendations through user feedback

### 👥 Social Features
- **Friends System**: Connect with other book lovers
- **Activity Feeds**: See what your friends are reading and rating
- **Shared Rankings**: Compare book preferences with friends
- **Social Discovery**: Find books through your network

### 📱 Modern UI/UX
- **Responsive Design**: Works perfectly on desktop, tablet, and mobile
- **Intuitive Interface**: Clean, modern design focused on usability
- **Real-time Updates**: Optimistic UI with instant feedback
- **Accessibility**: ARIA labels and keyboard navigation support

## 🏗️ Technology Stack

### Backend (Go)
- **Framework**: Go 1.22+ with Gin HTTP router
- **Database**: PostgreSQL 15+ with GORM ORM
- **Authentication**: JWT with bcrypt password hashing
- **External APIs**: OpenLibrary and Google Books integration
- **Architecture**: Clean architecture with dependency injection

### Frontend (React)
- **Framework**: React 18+ with Vite build tool
- **Routing**: React Router 6 for navigation
- **Styling**: Tailwind CSS for responsive design
- **State**: React Context API for global state
- **Forms**: React Hook Form with validation
- **HTTP**: Axios for API communication

### Infrastructure
- **Containerization**: Docker and Docker Compose
- **Database**: PostgreSQL with proper indexing and migrations
- **Monitoring**: Health checks and structured logging
- **Security**: Rate limiting, input validation, CORS

## 🚀 Quick Start

### Prerequisites
- **Go 1.22+**
- **Node.js 16+** and npm
- **PostgreSQL 15+**
- **Docker & Docker Compose** (optional but recommended)

### Method 1: Docker Compose (Recommended)

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd book-rank
   ```

2. **Set up environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start all services:**
   ```bash
   docker-compose up -d
   ```

4. **Access the application:**
   - **Frontend**: http://localhost:3000
   - **Backend API**: http://localhost:8080
   - **Database**: localhost:5432

### Method 2: Manual Setup

#### Backend Setup

1. **Start PostgreSQL database:**
   ```bash
   docker-compose up postgres -d
   ```

2. **Run database migrations:**
   ```bash
   make migrate-up
   ```

3. **Start the Go server:**
   ```bash
   go run cmd/server/main.go
   ```
   
   Server will be available at `http://localhost:8080`

#### Frontend Setup

1. **Navigate to frontend directory:**
   ```bash
   cd frontend
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Set up environment:**
   ```bash
   cp .env.example .env
   # Configure VITE_API_URL=http://localhost:8080
   ```

4. **Start development server:**
   ```bash
   npm run dev
   ```
   
   Application will be available at `http://localhost:3000`

## 📁 Project Structure

```
book-rank/
├── cmd/                    # Application entry points
│   └── server/            # Main server application
├── internal/              # Private application code
│   ├── api/              # HTTP handlers and middleware
│   ├── auth/             # Authentication service
│   ├── books/            # Book management service
│   ├── ratings/          # ELO rating system
│   ├── recommendations/  # Recommendation engine
│   └── config/          # Configuration management
├── frontend/             # React frontend application
│   ├── src/
│   │   ├── components/   # Reusable React components
│   │   ├── pages/        # Page components
│   │   ├── context/      # React Context providers
│   │   ├── services/     # API service functions
│   │   └── utils/        # Helper functions
│   ├── public/          # Static assets
│   └── package.json     # Frontend dependencies
├── migrations/          # Database migrations
├── pkg/                # Public packages
├── scripts/            # Build and deployment scripts
├── docker-compose.yml  # Development environment
├── Dockerfile         # Application container
├── Makefile          # Development commands
└── README.md         # This file
```

## 🔧 Configuration

### Backend Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=bookrank
DB_PASSWORD=secure_password
DB_NAME=bookrank

# Authentication (REQUIRED - Generate strong secret!)
JWT_SECRET=your-super-secret-jwt-key-make-it-long-and-random
TOKEN_EXPIRATION=24h
BCRYPT_COST=12

# External APIs (Optional)
OPENLIBRARY_API_URL=https://openlibrary.org
GOOGLE_BOOKS_API_KEY=your-google-books-api-key

# Security & Performance
RATE_LIMIT_PER_MINUTE=100
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Logging
LOG_LEVEL=info
ENVIRONMENT=development
```

### Frontend Environment Variables

```bash
# API Configuration
VITE_API_URL=http://localhost:8080

# Environment
NODE_ENV=development
```

## 📚 API Documentation

### Authentication Endpoints

```http
POST /auth/register          # Create new user account
POST /auth/login             # User login
POST /auth/refresh           # Refresh JWT token
GET  /api/me                 # Get current user info
```

### Book Management

```http
GET    /api/books            # List books with filters and pagination
POST   /api/books            # Create new book
GET    /api/books/{id}       # Get book details
PUT    /api/books/{id}       # Update book (owner only)
DELETE /api/books/{id}       # Delete book (owner only)
GET    /api/books/{id}/stats # Get book rating statistics
```

### Rating & Comparison System

```http
POST /api/comparisons        # Submit book comparison
GET  /api/comparisons        # Get comparison history
GET  /api/rankings           # Get book rankings (ELO-based)
```

### Recommendations

```http
GET  /api/recommendations           # Get personalized recommendations
GET  /api/recommendations/genre     # Get genre-based recommendations
POST /api/recommendations/{id}/feedback # Mark recommendation as interested/not
```

### Social Features

```http
GET    /api/friends              # Get friends list
POST   /api/friends              # Add friend by username
DELETE /api/friends/{id}         # Remove friend
GET    /api/friends/{id}/activity # Get friend activity
```

Full API documentation available at `/docs` when server is running.

## 🧪 Testing

### Backend Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/auth/... -v
```

### Frontend Tests

```bash
cd frontend

# Run tests
npm test

# Run with coverage
npm run test:coverage

# Run linting
npm run lint
```

## 🚀 Deployment

### Production Build

#### Backend
```bash
# Build binary
go build -o bookrank cmd/server/main.go

# Run migrations
./bookrank migrate

# Start server
./bookrank
```

#### Frontend
```bash
cd frontend

# Build for production
npm run build

# Deploy dist/ folder to your web server
```

### Docker Deployment

```bash
# Build and push images
docker build -t bookrank-api .
docker build -f frontend/Dockerfile -t bookrank-frontend ./frontend

# Deploy with docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

## 🔒 Security Features

### Authentication & Authorization
- JWT-based authentication with secure token handling
- Password hashing using bcrypt with configurable cost
- Rate limiting (100 requests/minute per user)
- Token refresh for sliding sessions

### Data Protection
- Input validation and sanitization
- SQL injection protection via ORM
- CORS configuration for cross-origin requests
- Environment-based configuration management

### Best Practices
- Structured logging for audit trails
- Health checks for monitoring
- Error handling with proper HTTP status codes
- Secure headers and middleware

## 📈 Performance Features

### Backend Optimizations
- Database connection pooling
- Proper database indexing
- Pagination for large datasets
- Caching for frequently accessed data
- Asynchronous external API calls

### Frontend Optimizations
- Code splitting with React Router
- Optimistic UI updates
- Image lazy loading and optimization
- Bundle optimization with Vite
- Responsive images and assets

## 🛠️ Development Commands

### Makefile Commands

```bash
make help                    # Show available commands
make build                   # Build the application
make run                     # Run the application
make test                    # Run tests
make migrate-up              # Run database migrations
make migrate-down            # Rollback migrations
make docker-build            # Build Docker images
make docker-up               # Start with Docker Compose
make clean                   # Clean build artifacts
```

### Database Commands

```bash
# Create migration
migrate create -ext sql -dir migrations -seq add_books_table

# Run migrations
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up

# Rollback migrations
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" down
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit changes**: `git commit -m 'Add amazing feature'`
4. **Push to branch**: `git push origin feature/amazing-feature`
5. **Open a Pull Request**

### Development Guidelines

- **Backend**: Follow Go best practices and maintain test coverage >80%
- **Frontend**: Use TypeScript-style prop validation and accessibility standards
- **Database**: Use migrations for all schema changes
- **Testing**: Write comprehensive tests for new features
- **Documentation**: Update README and API docs for changes

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

- **Documentation**: Check the `/docs` endpoint when running locally
- **Issues**: Create an issue on GitHub for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions and community

## 🎯 Roadmap

### Phase 1 (Current) ✅
- [x] Core book management system
- [x] JWT authentication
- [x] ELO rating system
- [x] Basic recommendations
- [x] React frontend with responsive design

### Phase 2 (In Progress) 🚧
- [ ] Advanced recommendation algorithms
- [ ] Social features enhancement
- [ ] Reading progress tracking
- [ ] Book import from external services
- [ ] Mobile app (React Native)

### Phase 3 (Planned) 📋
- [ ] Machine learning recommendations
- [ ] Reading challenges and goals
- [ ] Book club features
- [ ] Advanced analytics dashboard
- [ ] API rate limiting tiers
- [ ] Multi-language support

---

**Built with ❤️ for book lovers everywhere**