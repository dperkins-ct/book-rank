import React, { createContext, useContext, useReducer } from 'react'

const BookContext = createContext()

const initialState = {
  books: [],
  currentBook: null,
  searchResults: [],
  loading: false,
  error: null,
  pagination: {
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0
  },
  filters: {
    search: '',
    genre: '',
    author: '',
    minRating: null,
    maxRating: null,
    sort: 'created_at',
    order: 'desc'
  },
  comparison: {
    bookA: null,
    bookB: null,
    isComparing: false,
    history: []
  }
}

function bookReducer(state, action) {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, loading: action.payload }

    case 'SET_ERROR':
      return { ...state, error: action.payload, loading: false }

    case 'SET_BOOKS':
      return {
        ...state,
        books: action.payload.books,
        pagination: {
          page: action.payload.page,
          pageSize: action.payload.page_size,
          total: action.payload.total,
          totalPages: action.payload.total_pages
        },
        loading: false
      }

    case 'ADD_BOOK':
      return {
        ...state,
        books: [action.payload, ...state.books],
        pagination: {
          ...state.pagination,
          total: state.pagination.total + 1
        }
      }

    case 'UPDATE_BOOK':
      return {
        ...state,
        books: state.books.map(book =>
          book.id === action.payload.id ? action.payload : book
        ),
        currentBook: state.currentBook?.id === action.payload.id ? action.payload : state.currentBook
      }

    case 'DELETE_BOOK':
      return {
        ...state,
        books: state.books.filter(book => book.id !== action.payload),
        pagination: {
          ...state.pagination,
          total: Math.max(0, state.pagination.total - 1)
        }
      }

    case 'SET_CURRENT_BOOK':
      return { ...state, currentBook: action.payload }

    case 'SET_SEARCH_RESULTS':
      return { ...state, searchResults: action.payload }

    case 'SET_FILTERS':
      return { ...state, filters: { ...state.filters, ...action.payload } }

    case 'RESET_FILTERS':
      return { ...state, filters: initialState.filters }

    case 'START_COMPARISON':
      return {
        ...state,
        comparison: {
          ...state.comparison,
          bookA: action.payload.bookA,
          bookB: action.payload.bookB,
          isComparing: true
        }
      }

    case 'END_COMPARISON':
      return {
        ...state,
        comparison: {
          ...state.comparison,
          isComparing: false,
          history: [
            ...state.comparison.history,
            {
              bookA: state.comparison.bookA,
              bookB: state.comparison.bookB,
              winner: action.payload.winner,
              timestamp: new Date()
            }
          ]
        }
      }

    case 'CLEAR_COMPARISON':
      return {
        ...state,
        comparison: initialState.comparison
      }

    default:
      return state
  }
}

export function BookProvider({ children }) {
  const [state, dispatch] = useReducer(bookReducer, initialState)

  const setLoading = (loading) => {
    dispatch({ type: 'SET_LOADING', payload: loading })
  }

  const setError = (error) => {
    dispatch({ type: 'SET_ERROR', payload: error })
  }

  const setBooks = (booksData) => {
    dispatch({ type: 'SET_BOOKS', payload: booksData })
  }

  const addBook = (book) => {
    dispatch({ type: 'ADD_BOOK', payload: book })
  }

  const updateBook = (book) => {
    dispatch({ type: 'UPDATE_BOOK', payload: book })
  }

  const deleteBook = (bookId) => {
    dispatch({ type: 'DELETE_BOOK', payload: bookId })
  }

  const setCurrentBook = (book) => {
    dispatch({ type: 'SET_CURRENT_BOOK', payload: book })
  }

  const setSearchResults = (results) => {
    dispatch({ type: 'SET_SEARCH_RESULTS', payload: results })
  }

  const setFilters = (filters) => {
    dispatch({ type: 'SET_FILTERS', payload: filters })
  }

  const resetFilters = () => {
    dispatch({ type: 'RESET_FILTERS' })
  }

  const startComparison = (bookA, bookB) => {
    dispatch({ type: 'START_COMPARISON', payload: { bookA, bookB } })
  }

  const endComparison = (winner) => {
    dispatch({ type: 'END_COMPARISON', payload: { winner } })
  }

  const clearComparison = () => {
    dispatch({ type: 'CLEAR_COMPARISON' })
  }

  const value = {
    ...state,
    setLoading,
    setError,
    setBooks,
    addBook,
    updateBook,
    deleteBook,
    setCurrentBook,
    setSearchResults,
    setFilters,
    resetFilters,
    startComparison,
    endComparison,
    clearComparison
  }

  return (
    <BookContext.Provider value={value}>
      {children}
    </BookContext.Provider>
  )
}

export function useBooks() {
  const context = useContext(BookContext)
  if (!context) {
    throw new Error('useBooks must be used within a BookProvider')
  }
  return context
}