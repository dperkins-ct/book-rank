import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  BookOpenIcon,
  PlusIcon,
  ScaleIcon,
  HeartIcon,
  ChartBarIcon,
  TrophyIcon
} from '@heroicons/react/24/outline'
import { dashboardAPI, booksAPI, ratingsAPI } from '../../services/api'
import LoadingSpinner from '../../components/common/LoadingSpinner'
import Button from '../../components/common/Button'
import { formatNumber, formatRating, getRatingBadgeColor } from '../../utils/formatters'

function DashboardPage() {
  const [stats, setStats] = useState(null)
  const [recentBooks, setRecentBooks] = useState([])
  const [topRanked, setTopRanked] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      setLoading(true)

      // Load data sources with individual error handling
      const booksData = await booksAPI.getBooks({ limit: 5, sort: 'created_at', order: 'desc' })
        .then(response => response.data)
        .catch(error => {
          console.warn('Books API not available:', error.response?.status)
          return { books: [], total: 0 }
        })

      const rankingsData = await ratingsAPI.getRankings({ limit: 5 })
        .then(response => response.data)
        .catch(error => {
          console.warn('Rankings API not available:', error.response?.status)
          return { books: [] }
        })

      setRecentBooks(booksData.books || [])
      setTopRanked(rankingsData.books || [])

      // Calculate basic stats from the books data
      const totalBooks = booksData.total || 0
      const books = booksData.books || []
      const avgRating = books.length > 0
        ? books.reduce((sum, book) => sum + (book.average_rating || 0), 0) / books.length
        : 0

      setStats({
        totalBooks,
        totalComparisons: 0, // This would come from comparisons API
        avgRating: avgRating || 0,
        booksRead: totalBooks // Assuming all books in library are "read"
      })

    } catch (error) {
      console.error('Error loading dashboard data:', error)
      // Set default empty state
      setRecentBooks([])
      setTopRanked([])
      setStats({
        totalBooks: 0,
        totalComparisons: 0,
        avgRating: 0,
        booksRead: 0
      })
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  const quickActions = [
    {
      name: 'Add Book',
      href: '/books/add',
      icon: PlusIcon,
      color: 'bg-blue-500 hover:bg-blue-600'
    },
    {
      name: 'Compare Books',
      href: '/compare',
      icon: ScaleIcon,
      color: 'bg-purple-500 hover:bg-purple-600'
    },
    {
      name: 'Get Recommendations',
      href: '/recommendations',
      icon: HeartIcon,
      color: 'bg-pink-500 hover:bg-pink-600'
    }
  ]

  const statCards = [
    {
      name: 'Books in Library',
      value: formatNumber(stats?.totalBooks || 0),
      icon: BookOpenIcon,
      color: 'text-blue-600 bg-blue-100'
    },
    {
      name: 'Average Rating',
      value: formatRating(stats?.avgRating || 0),
      icon: ChartBarIcon,
      color: 'text-green-600 bg-green-100'
    },
    {
      name: 'Comparisons Made',
      value: formatNumber(stats?.totalComparisons || 0),
      icon: ScaleIcon,
      color: 'text-purple-600 bg-purple-100'
    },
    {
      name: 'Books Read',
      value: formatNumber(stats?.booksRead || 0),
      icon: TrophyIcon,
      color: 'text-yellow-600 bg-yellow-100'
    }
  ]

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="bg-gradient-to-r from-primary-600 to-primary-700 rounded-lg p-6 text-white">
        <h1 className="text-2xl font-bold mb-2">Welcome to BookRank!</h1>
        <p className="text-primary-100">
          Discover your next favorite book through intelligent comparisons and personalized recommendations.
        </p>
      </div>

      {/* Quick Actions */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          {quickActions.map((action) => (
            <Link
              key={action.name}
              to={action.href}
              className={`${action.color} text-white rounded-lg p-4 flex items-center space-x-3 transition-colors`}
            >
              <action.icon className="h-6 w-6" />
              <span className="font-medium">{action.name}</span>
            </Link>
          ))}
        </div>
      </div>

      {/* Stats */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Your Stats</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {statCards.map((stat) => (
            <div key={stat.name} className="card">
              <div className="flex items-center">
                <div className={`${stat.color} rounded-lg p-2`}>
                  <stat.icon className="h-6 w-6" />
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-500">{stat.name}</p>
                  <p className="text-2xl font-semibold text-gray-900">{stat.value}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Recent Books */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Recent Books</h2>
            <Link
              to="/books"
              className="text-sm text-primary-600 hover:text-primary-500"
            >
              View all
            </Link>
          </div>
          <div className="card">
            {recentBooks.length > 0 ? (
              <div className="space-y-4">
                {recentBooks.map((book) => (
                  <div key={book.id} className="flex items-center space-x-4">
                    <div className="flex-shrink-0">
                      <div className="h-12 w-8 bg-gray-200 rounded flex items-center justify-center">
                        <BookOpenIcon className="h-6 w-6 text-gray-400" />
                      </div>
                    </div>
                    <div className="flex-1 min-w-0">
                      <Link
                        to={`/books/${book.id}`}
                        className="text-sm font-medium text-gray-900 hover:text-primary-600"
                      >
                        {book.title}
                      </Link>
                      <p className="text-sm text-gray-500">{book.author}</p>
                    </div>
                    {book.average_rating && (
                      <span className={`badge ${getRatingBadgeColor(book.average_rating)}`}>
                        {formatRating(book.average_rating)}
                      </span>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-4">
                <BookOpenIcon className="h-8 w-8 text-gray-400 mx-auto mb-2" />
                <p className="text-gray-500">No books yet</p>
                <Button
                  as={Link}
                  to="/books/add"
                  className="mt-2"
                  size="sm"
                >
                  Add your first book
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Top Ranked Books */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Top Ranked</h2>
            <Link
              to="/books?sort=rating&order=desc"
              className="text-sm text-primary-600 hover:text-primary-500"
            >
              View rankings
            </Link>
          </div>
          <div className="card">
            {topRanked.length > 0 ? (
              <div className="space-y-4">
                {topRanked.map((book, index) => (
                  <div key={book.id} className="flex items-center space-x-4">
                    <div className="flex-shrink-0 w-8 text-center">
                      <span className="text-sm font-bold text-gray-900">
                        #{index + 1}
                      </span>
                    </div>
                    <div className="flex-shrink-0">
                      <div className="h-12 w-8 bg-gray-200 rounded flex items-center justify-center">
                        <BookOpenIcon className="h-6 w-6 text-gray-400" />
                      </div>
                    </div>
                    <div className="flex-1 min-w-0">
                      <Link
                        to={`/books/${book.id}`}
                        className="text-sm font-medium text-gray-900 hover:text-primary-600"
                      >
                        {book.title}
                      </Link>
                      <p className="text-sm text-gray-500">{book.author}</p>
                    </div>
                    <span className={`badge ${getRatingBadgeColor(book.average_rating || 0)}`}>
                      {formatRating(book.average_rating || 0)}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-4">
                <TrophyIcon className="h-8 w-8 text-gray-400 mx-auto mb-2" />
                <p className="text-gray-500">No rankings yet</p>
                <Button
                  as={Link}
                  to="/compare"
                  className="mt-2"
                  size="sm"
                >
                  Start comparing books
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default DashboardPage