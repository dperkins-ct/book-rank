import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { BookProvider } from './context/BookContext'
import { ToastProvider } from './components/common/Toast'
import ErrorBoundary from './components/common/ErrorBoundary'
import ProtectedRoute from './components/auth/ProtectedRoute'
import Layout from './components/layout/Layout'

// Pages
import LoginPage from './pages/auth/LoginPage'
import RegisterPage from './pages/auth/RegisterPage'
import DashboardPage from './pages/dashboard/DashboardPage'
import BookLibraryPage from './pages/books/BookLibraryPage'
import BookDetailsPage from './pages/books/BookDetailsPage'
import AddBookPage from './pages/books/AddBookPage'
import ComparisonPage from './pages/books/ComparisonPage'
import RankingsPage from './pages/books/RankingsPage'
import RecommendationsPage from './pages/recommendations/RecommendationsPage'
import FriendsPage from './pages/friends/FriendsPage'

function App() {
  return (
    <ErrorBoundary>
      <ToastProvider>
        <AuthProvider>
          <BookProvider>
            <Routes>
              {/* Public routes */}
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />

              {/* Protected routes */}
              <Route element={<ProtectedRoute />}>
                <Route element={<Layout />}>
                  <Route path="/dashboard" element={<DashboardPage />} />
                  <Route path="/books" element={<BookLibraryPage />} />
                  <Route path="/books/:id" element={<BookDetailsPage />} />
                  <Route path="/books/add" element={<AddBookPage />} />
                  <Route path="/compare" element={<ComparisonPage />} />
                  <Route path="/rankings" element={<RankingsPage />} />
                  <Route path="/recommendations" element={<RecommendationsPage />} />
                  <Route path="/friends" element={<FriendsPage />} />
                </Route>
              </Route>

              {/* Default redirect */}
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </BookProvider>
        </AuthProvider>
      </ToastProvider>
    </ErrorBoundary>
  )
}

export default App