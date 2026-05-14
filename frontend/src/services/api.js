import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || ''

// Create axios instance
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  register: (credentials) => api.post('/auth/register', credentials),
  getCurrentUser: () => api.get('/api/me'),
}

// Books API
export const booksAPI = {
  getBooks: (params = {}) => {
    const searchParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== null && value !== undefined && value !== '') {
        searchParams.append(key, value)
      }
    })
    return api.get(`/api/books?${searchParams.toString()}`)
  },

  getBook: (id) => api.get(`/api/books/${id}`),

  createBook: (bookData) => api.post('/api/books', bookData),

  updateBook: (id, bookData) => api.put(`/api/books/${id}`, bookData),

  deleteBook: (id) => api.delete(`/api/books/${id}`),

  searchBooks: (query, params = {}) => {
    const searchParams = new URLSearchParams({ q: query, ...params })
    return api.get(`/api/books/search?${searchParams.toString()}`)
  },

  getBookStats: (id) => api.get(`/api/books/${id}/stats`),

  refreshMetadata: (id) => api.post(`/api/books/${id}/metadata`),
}

// Ratings API (based on the ELO system from the backend)
export const ratingsAPI = {
  submitComparison: (bookAId, bookBId, winnerId) =>
    api.post('/api/comparisons', {
      book_a_id: bookAId,
      book_b_id: bookBId,
      winner_id: winnerId
    }),

  getComparisons: (params = {}) => {
    const searchParams = new URLSearchParams(params)
    return api.get(`/api/comparisons?${searchParams.toString()}`)
  },

  getRankings: (params = {}) => {
    const searchParams = new URLSearchParams(params)
    return api.get(`/api/rankings?${searchParams.toString()}`)
  },
}

// Recommendations API
export const recommendationsAPI = {
  getPersonalRecommendations: (params = {}) => {
    const searchParams = new URLSearchParams(params)
    return api.get(`/api/recommendations?${searchParams.toString()}`)
  },

  getGenreRecommendations: (genre, params = {}) => {
    const searchParams = new URLSearchParams({ genre, ...params })
    return api.get(`/api/recommendations/genre?${searchParams.toString()}`)
  },

  markRecommendation: (bookId, status) =>
    api.post(`/api/recommendations/${bookId}/feedback`, { status }),
}

// Friends API (assuming this exists based on requirements)
export const friendsAPI = {
  getFriends: () => api.get('/api/friends'),

  addFriend: (username) => api.post('/api/friends', { username }),

  removeFriend: (friendId) => api.delete(`/api/friends/${friendId}`),

  getFriendActivity: (friendId, params = {}) => {
    const searchParams = new URLSearchParams(params)
    return api.get(`/api/friends/${friendId}/activity?${searchParams.toString()}`)
  },

  getFriendRankings: (friendId, params = {}) => {
    const searchParams = new URLSearchParams(params)
    return api.get(`/api/friends/${friendId}/rankings?${searchParams.toString()}`)
  },
}

// Dashboard API
export const dashboardAPI = {
  getStats: () => api.get('/api/dashboard/stats'),

  getRecentActivity: (limit = 10) =>
    api.get(`/api/dashboard/activity?limit=${limit}`),

  getProgress: () => api.get('/api/dashboard/progress'),
}

// Health check
export const healthAPI = {
  check: () => axios.get(`${API_BASE_URL}/health`),
}

export default api