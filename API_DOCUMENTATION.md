# BookRank API Documentation

## Overview

The BookRank API provides comprehensive book management functionality with external metadata integration, advanced filtering, sorting, and pagination capabilities.

## Base URL

```
http://localhost:8080
```

## Authentication

All API endpoints (except health and auth endpoints) require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <your_jwt_token>
```

## Health Check

### GET /health

Check if the API is running and database is accessible.

**Response:**
```json
{
  "status": "healthy",
  "service": "bookrank"
}
```

## Authentication Endpoints

### POST /auth/register

Register a new user account.

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "jwt_token_here",
  "expires_at": 1234567890
}
```

### POST /auth/login

Login with existing credentials.

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "jwt_token_here",
  "expires_at": 1234567890
}
```

### GET /api/me

Get current user information.

**Response:**
```json
{
  "id": 1,
  "username": "johndoe"
}
```

## Book Management Endpoints

### GET /api/books

List books with filtering, sorting, and pagination.

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `genre` (string): Filter by genre
- `author` (string): Filter by author
- `search` (string): Search in title, author, or description
- `min_rating` (int): Minimum average rating (1-10)
- `max_rating` (int): Maximum average rating (1-10)
- `sort` (string): Sort field ("title", "author", "created_at", "rating")
- `order` (string): Sort direction ("asc", "desc")

**Response:**
```json
{
  "books": [
    {
      "id": 1,
      "title": "The Great Gatsby",
      "author": "F. Scott Fitzgerald",
      "genre": "Classic Literature",
      "publication_date": "1925-04-10T00:00:00Z",
      "description": "A classic American novel...",
      "created_by": 1,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z",
      "creator": {
        "id": 1,
        "username": "johndoe",
        "email": ""
      },
      "average_rating": 8.5,
      "total_ratings": 150
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

### POST /api/books

Create a new book.

**Request Body:**
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

**Response:**
```json
{
  "id": 1,
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald",
  "genre": "Classic Literature",
  "publication_date": "1925-04-10T00:00:00Z",
  "description": "A classic American novel about the Jazz Age.",
  "created_by": 1,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### GET /api/books/{id}

Get details for a specific book.

**Response:**
```json
{
  "id": 1,
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald",
  "genre": "Classic Literature",
  "publication_date": "1925-04-10T00:00:00Z",
  "description": "A classic American novel about the Jazz Age.",
  "created_by": 1,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "creator": {
    "id": 1,
    "username": "johndoe",
    "email": ""
  },
  "metadata": [
    {
      "id": 1,
      "source": "openlibrary",
      "external_id": "/works/OL468516W",
      "additional_data": {
        "cover_url": "https://covers.openlibrary.org/b/id/123456-L.jpg",
        "isbn_13": "9780743273565"
      },
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### PUT /api/books/{id}

Update a book (only the creator can update).

**Request Body:**
```json
{
  "title": "The Great Gatsby - Updated",
  "description": "An updated description..."
}
```

**Response:** Same as GET /api/books/{id}

### DELETE /api/books/{id}

Soft delete a book (only the creator can delete).

**Response:** 204 No Content

### GET /api/books/{id}/stats

Get statistics for a specific book.

**Response:**
```json
{
  "book_id": 1,
  "total_ratings": 150,
  "average_rating": 8.5,
  "highest_rating": 10,
  "lowest_rating": 3,
  "rating_distribution": {
    "1": 2,
    "2": 5,
    "3": 8,
    "4": 12,
    "5": 18,
    "6": 25,
    "7": 30,
    "8": 28,
    "9": 15,
    "10": 7
  }
}
```

### POST /api/books/{id}/metadata

Refresh external metadata for a book.

**Response:**
```json
{
  "message": "Metadata refresh initiated successfully"
}
```

### GET /api/books/search

Search for books by title or author.

**Query Parameters:**
- `q` (string, required): Search query
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)

**Response:**
```json
{
  "books": [
    {
      "id": 1,
      "title": "The Great Gatsby",
      "author": "F. Scott Fitzgerald",
      "genre": "Classic Literature",
      "average_rating": 8.5,
      "total_ratings": 150,
      "cover_url": "https://covers.openlibrary.org/b/id/123456-L.jpg",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1,
  "query": "gatsby"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error Type",
  "message": "Detailed error message"
}
```

### Common HTTP Status Codes

- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **204 No Content**: Request successful, no content returned
- **400 Bad Request**: Invalid request data
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Access denied
- **404 Not Found**: Resource not found
- **422 Unprocessable Entity**: Validation errors
- **500 Internal Server Error**: Server error

### Validation Errors

```json
{
  "error": "Validation failed",
  "message": "validation error: title is required"
}
```

## Features

### External API Integration

The system automatically fetches book metadata from:
1. **OpenLibrary API** (primary source)
2. **Google Books API** (fallback)

Metadata includes:
- Cover images
- ISBN numbers
- Publication details
- Additional book information

### Rate Limiting

API requests are limited to **100 requests per minute per user**.

### Input Validation

All inputs are validated and sanitized to prevent:
- XSS attacks
- SQL injection
- Invalid data types
- Required field violations

### Soft Deletion

Books are soft-deleted to preserve ranking history and relationships.

### Performance Features

- Database connection pooling
- Optimized queries with proper indexing
- Pagination for large datasets
- Async metadata fetching

## Getting Started

1. **Start the server:**
   ```bash
   ./bookrank
   ```

2. **Register a user:**
   ```bash
   curl -X POST http://localhost:8080/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"password123"}'
   ```

3. **Create a book:**
   ```bash
   curl -X POST http://localhost:8080/api/books \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"title":"Test Book","author":"Test Author","genre":"Fiction"}'
   ```

4. **List books:**
   ```bash
   curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/books
   ```