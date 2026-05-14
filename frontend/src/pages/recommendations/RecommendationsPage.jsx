import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  HeartIcon,
  BookOpenIcon,
  SparklesIcon,
  HandThumbUpIcon,
  HandThumbDownIcon,
  EyeSlashIcon
} from '@heroicons/react/24/outline'
import { HeartIcon as HeartSolidIcon } from '@heroicons/react/24/solid'
import { recommendationsAPI, booksAPI } from '../../services/api'
import Button from '../../components/common/Button'
import LoadingSpinner from '../../components/common/LoadingSpinner'
import BookCard from '../../components/book/BookCard'
import { formatRating, getRatingBadgeColor } from '../../utils/formatters'

function RecommendationsPage() {
  const [recommendations, setRecommendations] = useState([])
  const [genreRecommendations, setGenreRecommendations] = useState({})
  const [loading, setLoading] = useState(true)
  const [selectedGenre, setSelectedGenre] = useState('')
  const [feedbackLoading, setFeedbackLoading] = useState({})

  const genres = ['Fiction', 'Mystery', 'Science Fiction', 'Romance', 'Biography', 'History', 'Fantasy']

  useEffect(() => {
    loadRecommendations()
  }, [])

  const loadRecommendations = async () => {
    try {
      setLoading(true)
      const response = await recommendationsAPI.getPersonalRecommendations()
      setRecommendations(response.data.recommendations || [])
    } catch (error) {
      console.error('Error loading recommendations:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadGenreRecommendations = async (genre) => {
    if (genreRecommendations[genre]) return // Already loaded

    try {
      const response = await recommendationsAPI.getGenreRecommendations(genre)
      setGenreRecommendations(prev => ({
        ...prev,
        [genre]: response.data.recommendations || []
      }))
    } catch (error) {
      console.error('Error loading genre recommendations:', error)
    }
  }

  const handleGenreSelect = (genre) => {
    setSelectedGenre(genre)
    loadGenreRecommendations(genre)
  }

  const handleFeedback = async (bookId, status) => {
    try {
      setFeedbackLoading(prev => ({ ...prev, [bookId]: true }))
      await recommendationsAPI.markRecommendation(bookId, status)

      // Update recommendations list
      setRecommendations(prev =>
        prev.map(rec =>
          rec.book.id === bookId
            ? { ...rec, feedback: status }
            : rec
        )
      )
    } catch (error) {
      console.error('Error submitting feedback:', error)
    } finally {
      setFeedbackLoading(prev => ({ ...prev, [bookId]: false }))
    }
  }

  const RecommendationCard = ({ recommendation, showReason = true }) => {
    const { book, score, reason, feedback } = recommendation
    const coverImage = book.metadata?.find(m => m.additional_data?.cover_url)?.additional_data?.cover_url
    const isLoading = feedbackLoading[book.id]

    return (
      <div className="card">
        <div className="flex space-x-4">
          {/* Book Cover */}
          <div className="flex-shrink-0">
            {coverImage ? (
              <img
                src={coverImage}
                alt={`${book.title} cover`}
                className="h-32 w-20 object-cover rounded"
              />
            ) : (
              <div className="h-32 w-20 bg-gray-200 rounded flex items-center justify-center">
                <BookOpenIcon className="h-8 w-8 text-gray-400" />
              </div>
            )}
          </div>

          {/* Book Info */}
          <div className="flex-1">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <Link
                  to={`/books/${book.id}`}
                  className="text-lg font-semibold text-gray-900 hover:text-primary-600"
                >
                  {book.title}
                </Link>
                <p className="text-sm text-gray-600 mt-1">by {book.author}</p>

                {book.genre && (
                  <span className="inline-block bg-gray-100 text-gray-700 text-xs px-2 py-1 rounded mt-2">
                    {book.genre}
                  </span>
                )}

                {/* Rating */}
                {book.average_rating && (
                  <div className="flex items-center mt-2">
                    <span className={`badge ${getRatingBadgeColor(book.average_rating)}`}>
                      {formatRating(book.average_rating)}
                    </span>
                    <span className="text-xs text-gray-500 ml-2">
                      ({book.total_ratings} ratings)
                    </span>
                  </div>
                )}
              </div>

              {/* Recommendation Score */}
              <div className="text-right ml-4">
                <div className="flex items-center">
                  <SparklesIcon className="h-4 w-4 text-primary-500 mr-1" />
                  <span className="text-sm font-medium text-primary-600">
                    {Math.round(score * 100)}% match
                  </span>
                </div>
              </div>
            </div>

            {/* Description */}
            {book.description && (
              <p className="text-sm text-gray-600 mt-3 line-clamp-2">
                {book.description}
              </p>
            )}

            {/* Recommendation Reason */}
            {showReason && reason && (
              <div className="mt-3 p-3 bg-primary-50 rounded-lg">
                <p className="text-sm text-primary-700">
                  <SparklesIcon className="h-4 w-4 inline mr-1" />
                  {reason}
                </p>
              </div>
            )}

            {/* Feedback Actions */}
            <div className="flex items-center justify-between mt-4">
              <div className="flex space-x-2">
                <Button
                  size="sm"
                  variant={feedback === 'interested' ? 'primary' : 'secondary'}
                  onClick={() => handleFeedback(book.id, 'interested')}
                  disabled={isLoading}
                  loading={isLoading && feedback !== 'not_interested'}
                >
                  <HandThumbUpIcon className="h-4 w-4 mr-1" />
                  Interested
                </Button>
                <Button
                  size="sm"
                  variant={feedback === 'not_interested' ? 'danger' : 'secondary'}
                  onClick={() => handleFeedback(book.id, 'not_interested')}
                  disabled={isLoading}
                  loading={isLoading && feedback !== 'interested'}
                >
                  <HandThumbDownIcon className="h-4 w-4 mr-1" />
                  Not Interested
                </Button>
              </div>

              <Button
                as={Link}
                to={`/books/${book.id}`}
                size="sm"
                variant="ghost"
              >
                View Details
              </Button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="text-center">
        <HeartSolidIcon className="h-12 w-12 text-primary-600 mx-auto mb-4" />
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Book Recommendations</h1>
        <p className="text-gray-600 max-w-2xl mx-auto">
          Discover your next favorite book with personalized recommendations based on your reading preferences and comparisons.
        </p>
      </div>

      {/* Genre Navigation */}
      <div className="flex flex-wrap justify-center gap-2">
        <button
          onClick={() => setSelectedGenre('')}
          className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
            selectedGenre === ''
              ? 'bg-primary-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          }`}
        >
          For You
        </button>
        {genres.map((genre) => (
          <button
            key={genre}
            onClick={() => handleGenreSelect(genre)}
            className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
              selectedGenre === genre
                ? 'bg-primary-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {genre}
          </button>
        ))}
      </div>

      {/* Recommendations */}
      {loading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner size="lg" />
        </div>
      ) : selectedGenre === '' ? (
        /* Personal Recommendations */
        <div className="space-y-6">
          <h2 className="text-xl font-semibold text-gray-900">
            Recommended For You
          </h2>
          {recommendations.length > 0 ? (
            <div className="space-y-4">
              {recommendations.map((recommendation, index) => (
                <RecommendationCard
                  key={recommendation.book.id}
                  recommendation={recommendation}
                />
              ))}
            </div>
          ) : (
            <div className="text-center py-12">
              <HeartIcon className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">No recommendations yet</h3>
              <p className="mt-1 text-sm text-gray-500">
                Add more books to your library and make some comparisons to get personalized recommendations.
              </p>
              <div className="mt-6 flex justify-center space-x-4">
                <Button as={Link} to="/books/add">
                  Add Books
                </Button>
                <Button as={Link} to="/compare" variant="secondary">
                  Compare Books
                </Button>
              </div>
            </div>
          )}
        </div>
      ) : (
        /* Genre Recommendations */
        <div className="space-y-6">
          <h2 className="text-xl font-semibold text-gray-900">
            {selectedGenre} Recommendations
          </h2>
          {genreRecommendations[selectedGenre] ? (
            genreRecommendations[selectedGenre].length > 0 ? (
              <div className="space-y-4">
                {genreRecommendations[selectedGenre].map((recommendation, index) => (
                  <RecommendationCard
                    key={recommendation.book.id}
                    recommendation={recommendation}
                    showReason={false}
                  />
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <EyeSlashIcon className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">
                  No {selectedGenre.toLowerCase()} recommendations
                </h3>
                <p className="mt-1 text-sm text-gray-500">
                  Try adding some {selectedGenre.toLowerCase()} books to your library first.
                </p>
              </div>
            )
          ) : (
            <div className="flex justify-center py-8">
              <LoadingSpinner size="lg" />
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default RecommendationsPage