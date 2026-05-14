# BookRank - Personal Book Rating & Recommendation System

BookRank is a complete full-stack application for book ranking, comparison, and personalized recommendations built with Go backend and React frontend. The system helps you discover your next favorite book through intelligent comparisons and personalized recommendations using an advanced ELO rating system to build your personal library and get tailored suggestions based on your reading preferences.

## Running the Application Locally

To run BookRank locally, you need Go 1.22+, Node.js 16+, PostgreSQL 15+, and Docker with Docker Compose installed on your system. The quickest way to get started is using Docker Compose which handles all service dependencies automatically. First clone the repository and navigate to the project directory, then copy the example environment file with `cp .env.example .env` and edit the configuration as needed. Start all services using `docker-compose up -d` and the application will be available at http://localhost:3000 for the frontend and http://localhost:8080 for the backend API.

For manual setup without Docker, start PostgreSQL using `docker-compose up postgres -d`, then run database migrations with `make db-migrate`. Start the Go backend server by running `go run cmd/server/main.go` which will serve the API on port 8080. For the frontend, navigate to the frontend directory, install dependencies with `npm install`, copy the frontend environment file with `cp .env.example .env` and configure `VITE_API_URL=http://localhost:8080`, then start the development server with `npm run dev` to access the application on port 3000.

The system also provides a convenient `make start` command that automatically handles the complete setup process including starting Colima (if needed), database services, installing frontend dependencies, and launching both backend and frontend servers simultaneously.

## Configuration

The application uses environment variables for configuration with separate files for backend and frontend settings. The backend configuration in `.env` requires database connection details including DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, and DB_NAME for PostgreSQL connectivity. Authentication settings include JWT_SECRET which must be a strong random string, TOKEN_EXPIRATION for session duration, and BCRYPT_COST for password hashing strength. Optional external API keys can be configured for enhanced book metadata including Google Books API integration.

The frontend configuration uses a simple `.env` file with VITE_API_URL pointing to the backend service location. Additional server configuration includes rate limiting set to 100 requests per minute, CORS settings for cross-origin requests, and logging configuration with adjustable log levels and output formats. The system supports multiple environments including development, staging, and production with corresponding configuration files.

## Contributing

Contributors should fork the repository and create feature branches following the naming convention `feature/description-of-change`. The project maintains high code quality standards with Go best practices for backend development and modern React patterns for frontend work. All database schema changes must use proper migrations, and comprehensive tests are required for new features with a target coverage above 80%. Pull requests should include updated documentation for any API changes and follow the established commit message conventions used throughout the project history.

Development setup includes automated tooling with `make deps` installing required linters and migration tools, while `make setup` provides a complete development environment initialization. The project uses golangci-lint for Go code quality and standard npm tooling for frontend development including ESLint and Prettier for consistent code formatting.

## License

This project is licensed under the MIT License which permits unrestricted use, modification, and distribution. See the LICENSE file in the repository root for complete license terms and conditions.