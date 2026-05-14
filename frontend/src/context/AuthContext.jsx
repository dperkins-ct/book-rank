import React, { createContext, useContext, useReducer, useEffect } from 'react'
import { authAPI } from '../services/api'

const AuthContext = createContext()

const initialState = {
  user: null,
  token: localStorage.getItem('token'),
  loading: true,
  error: null
}

function authReducer(state, action) {
  switch (action.type) {
    case 'LOGIN_START':
    case 'REGISTER_START':
      return { ...state, loading: true, error: null }

    case 'LOGIN_SUCCESS':
    case 'REGISTER_SUCCESS':
      localStorage.setItem('token', action.payload.token)
      return {
        ...state,
        user: action.payload.user,
        token: action.payload.token,
        loading: false,
        error: null
      }

    case 'LOGIN_ERROR':
    case 'REGISTER_ERROR':
      return {
        ...state,
        loading: false,
        error: action.payload
      }

    case 'LOGOUT':
      localStorage.removeItem('token')
      return {
        ...state,
        user: null,
        token: null,
        loading: false,
        error: null
      }

    case 'SET_USER':
      return {
        ...state,
        user: action.payload,
        loading: false
      }

    case 'SET_LOADING':
      return {
        ...state,
        loading: action.payload
      }

    default:
      return state
  }
}

export function AuthProvider({ children }) {
  const [state, dispatch] = useReducer(authReducer, initialState)

  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem('token')
      if (token) {
        try {
          const response = await authAPI.getCurrentUser()
          dispatch({ type: 'SET_USER', payload: response.data })
        } catch (error) {
          localStorage.removeItem('token')
          dispatch({ type: 'LOGOUT' })
        }
      } else {
        dispatch({ type: 'SET_LOADING', payload: false })
      }
    }

    initializeAuth()
  }, [])

  const login = async (credentials) => {
    dispatch({ type: 'LOGIN_START' })
    try {
      const response = await authAPI.login(credentials)
      const { token: tokenData } = response.data

      // Store token immediately so it's available for the next API call
      localStorage.setItem('token', tokenData.token)

      // Get user data
      const userResponse = await authAPI.getCurrentUser()

      dispatch({
        type: 'LOGIN_SUCCESS',
        payload: { token: tokenData.token, user: userResponse.data }
      })
    } catch (error) {
      // Clean up token if user fetch fails
      localStorage.removeItem('token')
      dispatch({
        type: 'LOGIN_ERROR',
        payload: error.response?.data?.message || 'Login failed'
      })
      throw error
    }
  }

  const register = async (credentials) => {
    dispatch({ type: 'REGISTER_START' })
    try {
      const response = await authAPI.register(credentials)
      const { token: tokenData } = response.data

      // Get user data
      const userResponse = await authAPI.getCurrentUser()

      dispatch({
        type: 'REGISTER_SUCCESS',
        payload: { token: tokenData.token, user: userResponse.data }
      })
    } catch (error) {
      dispatch({
        type: 'REGISTER_ERROR',
        payload: error.response?.data?.message || 'Registration failed'
      })
      throw error
    }
  }

  const logout = () => {
    dispatch({ type: 'LOGOUT' })
  }

  const value = {
    ...state,
    login,
    register,
    logout,
    isAuthenticated: !!state.token
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}