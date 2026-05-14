import React, { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  BookOpenIcon,
  StarIcon,
  CalendarIcon,
  TagIcon,
  PencilIcon,
  TrashIcon,
  ArrowLeftIcon,
  ChartBarIcon
} from '@heroicons/react/24/outline'
import { StarIcon as StarSolidIcon } from '@heroicons/react/24/solid'
import { booksAPI } from '../../services/api'
import { useBooks } from '../../context/BookContext'
import Button from '../../components/common/Button'
import LoadingSpinner from '../../components/common/LoadingSpinner'
import Modal from '../../components/common/Modal'
import { formatDate, formatRating, getRatingBadgeColor } from '../../utils/formatters'

function BookDetailsPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { deleteBook } = useBooks()
  const [book, setBook] = useState(null)
  const [stats, setStats] = useState(null)
  const [loading, setLoading] = useState(true)
  const [showDeleteModal, setShowDeleteModal] = useState(false)

  useEffect(() => {
    loadBookDetails()
  }, [id])

  const loadBookDetails = async () => {
    try {
      setLoading(true)
      const [bookResponse, statsResponse] = await Promise.all([
        booksAPI.getBook(id),
        booksAPI.getBookStats(id).catch(() => ({ data: null })) // Stats might not exist
      ])

      setBook(bookResponse.data)
      setStats(statsResponse.data)
    } catch (error) {
      console.error('Error loading book details:', error)
      if (error.response?.status === 404) {
        navigate('/books')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    try {
      await booksAPI.deleteBook(book.id)
      deleteBook(book.id)
      navigate('/books')
    } catch (error) {
      console.error('Error deleting book:', error)
    }
  }

  const handleRefreshMetadata = async () => {
    try {
      await booksAPI.refreshMetadata(book.id)
      // Reload book details to get updated metadata
      await loadBookDetails()
    } catch (error) {
      console.error('Error refreshing metadata:', error)
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (!book) {
    return (
      <div className="text-center py-8">
        <h2 className="text-lg font-medium text-gray-900">Book not found</h2>
        <p className="text-gray-500 mt-1">The book you&apos;re looking for doesn&apos;t exist.</p>
        <Button
          as={Link}
          to="/books"
          className="mt-4"
        >
          Back to Library
        </Button>
      </div>
    )
  }

  const coverImage = book.metadata?.find(m => m.additional_data?.cover_url)?.additional_data?.cover_url
  const isbn = book.metadata?.find(m => m.additional_data?.isbn_13)?.additional_data?.isbn_13

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center space-x-4">
        <Button
          variant="ghost"
          onClick={() => navigate('/books')}
        >
          <ArrowLeftIcon className="h-4 w-4 mr-2" />
          Back to Library
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Content */}
        <div className="lg:col-span-2">
          <div className="card">
            <div className="flex space-x-6">
              {/* Book Cover */}
              <div className="flex-shrink-0">
                {coverImage ? (
                  <img
                    src={coverImage}
                    alt={`${book.title} cover`}
                    className="h-48 w-32 object-cover rounded-lg shadow-sm"
                  />
                ) : (
                  <div className="h-48 w-32 bg-gray-200 rounded-lg flex items-center justify-center">
                    <BookOpenIcon className="h-16 w-16 text-gray-400" />
                  </div>
                )}
              </div>

              {/* Book Info */}
              <div className="flex-1">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h1 className="text-2xl font-bold text-gray-900 mb-2">
                      {book.title}
                    </h1>
                    <p className="text-lg text-gray-600 mb-3">by {book.author}</p>

                    {/* Rating */}
                    <div className="flex items-center space-x-4 mb-4">
                      {book.average_rating ? (
                        <>
                          <div className="flex items-center">
                            <StarSolidIcon className="h-5 w-5 text-yellow-400 mr-1" />
                            <span className="text-lg font-semibold">
                              {formatRating(book.average_rating)}
                            </span>
                          </div>
                          <span className="text-sm text-gray-500">
                            ({book.total_ratings} rating{book.total_ratings !== 1 ? 's' : ''})
                          </span>
                        </>
                      ) : (
                        <span className="text-gray-500">Not yet rated</span>
                      )}
                    </div>

                    {/* Metadata */}
                    <div className="space-y-2">
                      {book.genre && (
                        <div className="flex items-center text-sm text-gray-600">
                          <TagIcon className="h-4 w-4 mr-2" />
                          <span className="badge badge-secondary">{book.genre}</span>
                        </div>
                      )}
                      {book.publication_date && (
                        <div className="flex items-center text-sm text-gray-600">
                          <CalendarIcon className="h-4 w-4 mr-2" />
                          Published {formatDate(book.publication_date, 'yyyy')}
                        </div>
                      )}
                      {isbn && (
                        <div className="flex items-center text-sm text-gray-600">
                          <span className="font-medium mr-2">ISBN:</span>
                          {isbn}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex space-x-2">
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => navigate(`/books/${book.id}/edit`)}
                    >
                      <PencilIcon className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={() => setShowDeleteModal(true)}
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            {/* Description */}
            {book.description && (
              <div className="mt-6 pt-6 border-t">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Description</h3>
                <p className="text-gray-700 leading-relaxed">{book.description}</p>
              </div>
            )}

            {/* Quick Actions */}
            <div className="mt-6 pt-6 border-t">
              <div className="flex flex-wrap gap-3">
                <Button
                  as={Link}
                  to="/compare"
                  state={{ preselectedBook: book }}
                  variant="secondary"
                >
                  Compare with Other Books
                </Button>
                <Button
                  variant="secondary"
                  onClick={handleRefreshMetadata}
                >
                  Refresh Metadata
                </Button>
              </div>
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Rating Statistics */}
          {stats && (
            <div className="card">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                <ChartBarIcon className="h-5 w-5 inline mr-2" />
                Rating Statistics
              </h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Total Ratings:</span>
                  <span className="font-medium">{stats.total_ratings}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Average:</span>
                  <span className="font-medium">{formatRating(stats.average_rating)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Highest:</span>
                  <span className="font-medium">{stats.highest_rating}/10</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Lowest:</span>
                  <span className="font-medium">{stats.lowest_rating}/10</span>
                </div>
              </div>

              {/* Rating Distribution */}
              {stats.rating_distribution && (
                <div className="mt-4 pt-4 border-t">
                  <h4 className="text-sm font-medium text-gray-900 mb-3">Rating Distribution</h4>
                  <div className="space-y-2">
                    {Object.entries(stats.rating_distribution)
                      .sort((a, b) => parseInt(b[0]) - parseInt(a[0]))
                      .map(([rating, count]) => (
                        <div key={rating} className="flex items-center space-x-2">
                          <span className="text-xs w-6">{rating}★</span>
                          <div className="flex-1 bg-gray-200 rounded-full h-2">
                            <div
                              className="bg-primary-600 h-2 rounded-full"
                              style={{
                                width: `${(count / stats.total_ratings) * 100}%`
                              }}
                            />
                          </div>
                          <span className="text-xs text-gray-500 w-8">{count}</span>
                        </div>
                      ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Book Details */}
          <div className="card">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Details</h3>
            <div className="space-y-3">
              <div>
                <span className="text-sm text-gray-600">Added by:</span>
                <p className="font-medium">{book.creator?.username}</p>
              </div>
              <div>
                <span className="text-sm text-gray-600">Date added:</span>
                <p className="font-medium">{formatDate(book.created_at)}</p>
              </div>
              {book.updated_at !== book.created_at && (
                <div>
                  <span className="text-sm text-gray-600">Last updated:</span>
                  <p className="font-medium">{formatDate(book.updated_at)}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Delete Book"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-500">
            Are you sure you want to delete "{book.title}"? This will also remove all associated ratings and comparisons. This action cannot be undone.
          </p>
          <div className="flex justify-end space-x-3">
            <Button
              variant="secondary"
              onClick={() => setShowDeleteModal(false)}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleDelete}
            >
              Delete Book
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default BookDetailsPage