# BookRank API Documentation

The BookRank API provides comprehensive book management functionality with external metadata integration, advanced filtering, sorting, and pagination capabilities. The API is built using Go with Gin HTTP router and provides JWT-based authentication for secure access to all endpoints.

## Base Configuration

The API runs on `http://localhost:8080` by default and requires JWT authentication via the `Authorization` header for all endpoints except health checks and authentication endpoints. All API requests should include `Authorization: Bearer <your_jwt_token>` in the header for authenticated access.

## Health Check

The health endpoint at `GET /health` provides a simple status check that returns `{"status": "healthy", "service": "bookrank"}` when the API is running and the database connection is accessible.

## Authentication Endpoints

User registration is handled through `POST /auth/register` with a JSON payload containing username and password fields. The endpoint returns a JWT token and expiration timestamp upon successful registration. User login follows the same pattern with `POST /auth/login` using identical credentials format. The current user information can be retrieved using `GET /api/me` which returns the user ID and username when properly authenticated.

**Registration Request:**
```json
{
  "username": "johndoe",
  "password": "password123"
}
```

**Login/Registration Response:**
```json
{
  "token": "jwt_token_here",
  "expires_at": 1234567890
}
```

**Current User Response:**
```json
{
  "id": 1,
  "username": "johndoe"
}
```

## Book Management Endpoints

The primary book listing endpoint `GET /api/books` supports comprehensive filtering, sorting, and pagination through query parameters. You can filter by genre, author, search terms, and rating ranges while controlling pagination with page and limit parameters. The sort parameter accepts "title", "author", "created_at", or "rating" with ascending or descending order.

**Available Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `genre` (string): Filter by genre
- `author` (string): Filter by author
- `search` (string): Search in title, author, or description
- `min_rating` (int): Minimum average rating (1-10)
- `max_rating` (int): Maximum average rating (1-10)
- `sort` (string): Sort field ("title", "author", "created_at", "rating")
- `order` (string): Sort direction ("asc", "desc")

Creating new books through `POST /api/books` accepts title, author, genre, publication date, and description fields. The `fetch_metadata` boolean flag enables automatic retrieval of additional book information from external APIs. Individual book details are accessible through `GET /api/books/{id}` which includes creator information and external metadata when available.

**Book Creation Request:**
```json
{
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald", 
  "genre": "Classic Literature",
  "publication_date": "1925-04-10T00:00:00Z",
  "description": "A classic American novel about the Jazz Age.",
  "fetch_metadata": true
}
```

Book updates are performed using `PUT /api/books/{id}` and are restricted to the book creator. Deletion through `DELETE /api/books/{id}` implements soft deletion to preserve ranking history and relationships. Book statistics including rating distribution and averages are available through `GET /api/books/{id}/stats`.

The search functionality at `GET /api/books/search` requires a query parameter `q` and supports the same pagination options as the main book listing. External metadata refresh can be triggered using `POST /api/books/{id}/metadata` which initiates background fetching from OpenLibrary and Google Books APIs.

## External API Integration

The system automatically fetches book metadata from OpenLibrary API as the primary source and Google Books API as fallback. This integration provides cover images, ISBN numbers, publication details, and additional book information that enhances the user experience without manual data entry.

## Rate Limiting and Security

API requests are limited to 100 requests per minute per user to ensure fair usage and system stability. All inputs undergo validation and sanitization to prevent XSS attacks, SQL injection, and invalid data types. The system enforces required field validation and uses proper HTTP status codes for error responses.

## Error Handling

All error responses follow a consistent JSON format with error type and detailed message fields. Common HTTP status codes include 200 (OK), 201 (Created), 204 (No Content), 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), 404 (Not Found), 422 (Unprocessable Entity), and 500 (Internal Server Error). Validation errors provide specific field-level feedback to help with troubleshooting.

**Error Response Format:**
```json
{
  "error": "Error Type",
  "message": "Detailed error message"
}
```

## Performance Features

The API implements database connection pooling, optimized queries with proper indexing, pagination for large datasets, and asynchronous metadata fetching to ensure responsive performance. Soft deletion preserves data integrity while maintaining system responsiveness.

## Getting Started Example

To begin using the API, start the server and register a user account. Use the returned JWT token for subsequent API calls to create and manage your book collection. The system supports immediate book creation with optional metadata enrichment from external sources.

```bash
# Register user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Create book
curl -X POST http://localhost:8080/api/books \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Book","author":"Test Author","genre":"Fiction"}'

# List books
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/books
```