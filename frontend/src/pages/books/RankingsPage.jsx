import React, { useState, useEffect } from 'react'
import { TrophyIcon, ChartBarIcon, StarIcon } from '@heroicons/react/24/outline'
import { ratingsAPI } from '../../services/api'
import LoadingSpinner from '../../components/common/LoadingSpinner'
import { formatRating, getRatingBadgeColor, convertPersonalRankingToScore, convertELOToBookRank } from '../../utils/formatters'

function RankingsPage() {
  const [rankings, setRankings] = useState([])
  const [stats, setStats] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    loadRankings()
  }, [])

  const loadRankings = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await ratingsAPI.getMyRankings()
      setRankings(response.data.rankings || [])
      setStats(response.data.stats || null)
    } catch (error) {
      console.error('Error loading rankings:', error)
      setError('Failed to load rankings. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <div className="text-red-500 mb-4">{error}</div>
        <button
          onClick={loadRankings}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Try Again
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="text-center">
        <TrophyIcon className="h-12 w-12 text-yellow-500 mx-auto mb-4" />
        <h1 className="text-3xl font-bold text-gray-900 mb-2">My Book Rankings</h1>
        <p className="text-gray-600 max-w-2xl mx-auto">
          Your personalized book rankings based on your comparison preferences.
          Rankings are calculated using our proprietary BookRank algorithm.
        </p>
      </div>

      {/* Stats Summary */}
      {stats && (
        <div className="max-w-4xl mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div className="card text-center">
              <ChartBarIcon className="h-8 w-8 text-blue-500 mx-auto mb-2" />
              <div className="text-2xl font-bold text-gray-900">{stats.total_books || 0}</div>
              <div className="text-sm text-gray-600">Books Ranked</div>
            </div>
            <div className="card text-center">
              <StarIcon className="h-8 w-8 text-yellow-500 mx-auto mb-2" />
              <div className="text-2xl font-bold text-gray-900">
                {rankings.length > 0 ?
                  ((rankings.reduce((sum, _, index) => sum + convertPersonalRankingToScore(rankings, rankings[index]), 0) / rankings.length).toFixed(1))
                  : 'N/A'}
              </div>
              <div className="text-sm text-gray-600">Average Personal Score</div>
            </div>
            <div className="card text-center">
              <TrophyIcon className="h-8 w-8 text-green-500 mx-auto mb-2" />
              <div className="text-2xl font-bold text-gray-900">{stats.comparisons_made || 0}</div>
              <div className="text-sm text-gray-600">Comparisons Made</div>
            </div>
          </div>
        </div>
      )}

      {/* Rankings List */}
      <div className="max-w-4xl mx-auto">
        {rankings.length > 0 ? (
          <div className="card">
            <div className="space-y-4">
              <div className="flex items-center justify-between pb-4 border-b">
                <h2 className="text-xl font-semibold text-gray-900">Your Book Rankings</h2>
                <div className="text-sm text-gray-500">
                  {rankings.length} book{rankings.length !== 1 ? 's' : ''} ranked
                </div>
              </div>

              {rankings.map((ranking, index) => {
                // Calculate personal 0-10 score based on position in personal rankings
                const personalScore = convertPersonalRankingToScore(rankings, ranking)

                // Medal colors: 1st = Gold, 2nd = Silver, 3rd = Bronze, rest = Gray
                const getPositionColor = (position) => {
                  if (position === 0) return 'bg-yellow-500' // Gold
                  if (position === 1) return 'bg-gray-400'   // Silver
                  if (position === 2) return 'bg-amber-600'  // Bronze
                  return 'bg-gray-300' // Normal
                }

                return (
                  <div key={ranking.book_id} className="flex items-center justify-between py-4 border-b last:border-b-0">
                    <div className="flex items-center space-x-4">
                      <div className="flex-shrink-0">
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold text-white ${getPositionColor(index)}`}>
                          {index + 1}
                        </div>
                      </div>
                    <div className="flex-grow">
                      <h3 className="text-lg font-medium text-gray-900">
                        {ranking.book?.title || 'Unknown Title'}
                      </h3>
                      <p className="text-sm text-gray-600">
                        {ranking.book?.author || 'Unknown Author'}
                      </p>
                      {ranking.book?.genre && (
                        <span className="inline-block px-2 py-1 text-xs text-gray-600 bg-gray-100 rounded-full mt-1">
                          {ranking.book.genre}
                        </span>
                      )}
                    </div>
                  </div>

                    <div className="text-right">
                      <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getRatingBadgeColor(personalScore * 100)}`}>
                        {personalScore.toFixed(1)}/10
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        Personal Rank #{index + 1}
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        ) : (
          <div className="card text-center py-8">
            <TrophyIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No Rankings Yet</h3>
            <p className="text-gray-600 mb-4">
              Start comparing books to build your personal rankings!
            </p>
            <a
              href="/books/compare"
              className="inline-flex items-center px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
            >
              Start Comparing Books
            </a>
          </div>
        )}
      </div>
    </div>
  )
}

export default RankingsPage