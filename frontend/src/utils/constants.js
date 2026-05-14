// API Constants
export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/auth/login',
    REGISTER: '/auth/register',
    ME: '/api/me',
  },
  BOOKS: {
    LIST: '/api/books',
    CREATE: '/api/books',
    DETAILS: (id) => `/api/books/${id}`,
    UPDATE: (id) => `/api/books/${id}`,
    DELETE: (id) => `/api/books/${id}`,
    SEARCH: '/api/books/search',
    STATS: (id) => `/api/books/${id}/stats`,
    METADATA: (id) => `/api/books/${id}/metadata`,
  },
  RATINGS: {
    COMPARISONS: '/api/comparisons',
    RANKINGS: '/api/rankings',
  },
  RECOMMENDATIONS: {
    PERSONAL: '/api/recommendations',
    GENRE: '/api/recommendations/genre',
    FEEDBACK: (id) => `/api/recommendations/${id}/feedback`,
  },
  FRIENDS: {
    LIST: '/api/friends',
    ADD: '/api/friends',
    REMOVE: (id) => `/api/friends/${id}`,
    ACTIVITY: (id) => `/api/friends/${id}/activity`,
    RANKINGS: (id) => `/api/friends/${id}/rankings`,
  },
  DASHBOARD: {
    STATS: '/api/dashboard/stats',
    ACTIVITY: '/api/dashboard/activity',
    PROGRESS: '/api/dashboard/progress',
  },
}

// UI Constants
export const RATING_SCALE = {
  MIN: 1,
  MAX: 10,
}

export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
}

export const SORT_OPTIONS = {
  BOOKS: [
    { value: 'created_at|desc', label: 'Recently Added' },
    { value: 'created_at|asc', label: 'Oldest First' },
    { value: 'title|asc', label: 'Title A-Z' },
    { value: 'title|desc', label: 'Title Z-A' },
    { value: 'author|asc', label: 'Author A-Z' },
    { value: 'rating|desc', label: 'Highest Rated' },
    { value: 'rating|asc', label: 'Lowest Rated' },
  ],
}

export const GENRES = [
  'Fiction',
  'Non-Fiction',
  'Mystery',
  'Science Fiction',
  'Fantasy',
  'Romance',
  'Biography',
  'History',
  'Self-Help',
  'Business',
  'Health',
  'Travel',
  'Cooking',
  'Art',
  'Religion',
  'Philosophy',
  'Poetry',
  'Drama',
  'Adventure',
  'Humor',
]

export const BOOK_STATUS = {
  WANT_TO_READ: 'want_to_read',
  CURRENTLY_READING: 'currently_reading',
  READ: 'read',
}

export const RECOMMENDATION_STATUS = {
  INTERESTED: 'interested',
  NOT_INTERESTED: 'not_interested',
  ADDED_TO_LIBRARY: 'added_to_library',
}

// Error Messages
export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Network error. Please check your connection and try again.',
  AUTH_REQUIRED: 'Please log in to continue.',
  UNAUTHORIZED: 'You are not authorized to perform this action.',
  NOT_FOUND: 'The requested resource was not found.',
  VALIDATION_ERROR: 'Please check your input and try again.',
  SERVER_ERROR: 'Server error. Please try again later.',
}

// Local Storage Keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'token',
  USER_PREFERENCES: 'user_preferences',
  COMPARISON_HISTORY: 'comparison_history',
}

// App Configuration
export const APP_CONFIG = {
  NAME: 'BookRank',
  DESCRIPTION: 'Discover your next favorite book through intelligent comparisons and personalized recommendations',
  VERSION: '1.0.0',
  AUTHOR: 'BookRank Team',
}

// Feature Flags
export const FEATURES = {
  SOCIAL_FEATURES: true,
  ADVANCED_RECOMMENDATIONS: true,
  EXPORT_DATA: false,
  DARK_MODE: false,
}