import React, { useState, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import {
  BookOpenIcon,
  ScaleIcon,
  ArrowPathIcon,
  TrophyIcon,
  ChartBarIcon
} from '@heroicons/react/24/outline'
import { booksAPI, ratingsAPI } from '../../services/api'
import { useBooks } from '../../context/BookContext'
import Button from '../../components/common/Button'
import LoadingSpinner from '../../components/common/LoadingSpinner'
import { formatRating, getRatingBadgeColor } from '../../utils/formatters'

function ComparisonPage() {
  const location = useLocation()
  const { comparison, startComparison, endComparison } = useBooks()

  const [loading, setLoading] = useState(true)
  const [comparing, setComparing] = useState(false)
  const [comparisonHistory, setComparisonHistory] = useState([])

  useEffect(() => {
    loadRandomBookPair()
    loadComparisonHistory()
  }, [])

  const loadRandomBookPair = async (showLoading = true) => {
    try {
      if (showLoading) {
        setLoading(true)
      }
      const response = await ratingsAPI.getRandomBookPair()
      const bookPair = response.data.book_pair
      if (bookPair) {
        startComparison(bookPair.book_a, bookPair.book_b)
      } else {
        // No more book pairs available
        startComparison(null, null)
      }
    } catch (error) {
      console.error('Error loading random book pair:', error)
      startComparison(null, null)
    } finally {
      if (showLoading) {
        setLoading(false)
      }
    }
  }

  const loadComparisonHistory = async () => {
    try {
      const response = await ratingsAPI.getComparisons({ limit: 10 })
      setComparisonHistory(response.data.comparisons || [])
    } catch (error) {
      console.error('Error loading comparison history:', error)
    }
  }


  const handleComparison = async (winner) => {
    if (!comparison.bookA || !comparison.bookB || comparing) return

    try {
      setComparing(true)

      // Show winner immediately for better UX
      endComparison(winner)

      // Submit comparison and update history
      await Promise.all([
        ratingsAPI.submitComparison(
          comparison.bookA.id,
          comparison.bookB.id,
          winner.id
        ),
        loadComparisonHistory()
      ])

      // Short delay to show the winner state, then load next pair
      setTimeout(() => {
        loadRandomBookPair(false) // Don't show loading spinner for smoother transition
      }, 800)
    } catch (error) {
      console.error('Error submitting comparison:', error)
      // Reset on error
      startComparison(comparison.bookA, comparison.bookB)
    } finally {
      setComparing(false)
    }
  }

  const selectNextComparison = () => {
    loadRandomBookPair(true) // Show loading when manually skipping
  }

  const resetComparison = () => {
    startComparison(null, null)
  }

  const BookCard = ({ book, label, onSelect, isWinner = false, showSelectButton = false }) => {
    const coverImage = book?.metadata?.find(m => m.additional_data?.cover_url)?.additional_data?.cover_url

    return (
      <div className={`card cursor-pointer transition-all duration-300 ${
        isWinner ? 'ring-2 ring-green-500 bg-green-50' : 'hover:shadow-lg'
      } ${showSelectButton ? 'opacity-50' : ''}`}>
        <div className="text-center">
          <div className="mb-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">{label}</h3>

            {book ? (
              <div className="space-y-4">
                {/* Book Cover */}
                <div className="flex justify-center">
                  {coverImage ? (
                    <img
                      src={coverImage}
                      alt={`${book.title} cover`}
                      className="h-32 w-20 object-cover rounded shadow-sm"
                    />
                  ) : (
                    <div className="h-32 w-20 bg-gray-200 rounded flex items-center justify-center">
                      <BookOpenIcon className="h-8 w-8 text-gray-400" />
                    </div>
                  )}
                </div>

                {/* Book Info */}
                <div>
                  <h4 className="font-semibold text-gray-900 text-center mb-1">
                    {book.title}
                  </h4>
                  <p className="text-sm text-gray-600 text-center mb-2">
                    by {book.author}
                  </p>
                  {book.genre && (
                    <span className="badge badge-secondary">{book.genre}</span>
                  )}
                </div>

                {/* Rating */}
                {book.average_rating && (
                  <div className="flex justify-center">
                    <span className={`badge ${getRatingBadgeColor(book.average_rating)}`}>
                      {formatRating(book.average_rating)}
                    </span>
                  </div>
                )}

                {/* Description Preview */}
                {book.description && (
                  <p className="text-xs text-gray-500 line-clamp-3 text-center">
                    {book.description}
                  </p>
                )}

                {/* Action Button */}
                {!comparing && !isWinner && comparison.bookA && comparison.bookB && (
                  <Button
                    onClick={() => handleComparison(book)}
                    className="w-full"
                    variant="primary"
                  >
                    I Prefer This Book
                  </Button>
                )}
              </div>
            ) : (
              <div className="py-8">
                <BookOpenIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
                <Button onClick={onSelect} variant="secondary">
                  Select Book
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="text-center">
        <ScaleIcon className="h-12 w-12 text-primary-600 mx-auto mb-4" />
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Book Comparison</h1>
        <p className="text-gray-600 max-w-2xl mx-auto">
          We'll show you two random books to compare. Choose which book you prefer, or skip pairs
          you don't want to compare. Your preferences help us build a personalized ranking system
          and create better recommendations for you.
        </p>
      </div>

      {/* Comparison Interface */}
      <div className="max-w-4xl mx-auto">
        {comparison.bookA && comparison.bookB ? (
          <div className="relative">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <BookCard
                book={comparison.bookA}
                label="Book A"
                isWinner={comparison.lastWinner === comparison.bookA}
              />

              <BookCard
                book={comparison.bookB}
                label="Book B"
                isWinner={comparison.lastWinner === comparison.bookB}
              />
            </div>

            {/* VS symbol positioned absolutely within the relative container */}
            <div className="hidden md:flex absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-10 pointer-events-none">
              <div className="bg-white rounded-full p-4 border-2 border-gray-300 shadow-lg">
                <span className="text-lg font-bold text-gray-700">VS</span>
              </div>
            </div>

            {/* VS symbol for mobile - shown between cards */}
            <div className="md:hidden flex justify-center py-4">
              <div className="bg-white rounded-full p-3 border-2 border-gray-200 shadow-md">
                <span className="text-xl font-bold text-gray-600">VS</span>
              </div>
            </div>
          </div>
        ) : loading ? (
          <div className="flex flex-col items-center justify-center py-8">
            <LoadingSpinner size="lg" />
            <p className="text-gray-600 mt-4">Finding your next book pair...</p>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-8">
            <div className="text-center text-gray-500">
              <p className="text-lg mb-4">No more book pairs available for comparison!</p>
              <p className="text-sm">Add more books to continue comparing.</p>
            </div>
          </div>
        )}

        {/* Controls */}
        {comparison.bookA && comparison.bookB && (
          <div className="flex justify-center space-x-4 mt-8">
            <Button
              onClick={selectNextComparison}
              variant="secondary"
              disabled={comparing}
            >
              <ArrowPathIcon className="h-4 w-4 mr-2" />
              Skip This Pair
            </Button>
            <Button
              onClick={resetComparison}
              variant="outline"
              disabled={comparing}
            >
              Reset
            </Button>
          </div>
        )}

        {comparing && (
          <div className="flex flex-col items-center justify-center mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
            <LoadingSpinner size="sm" />
            <p className="text-sm text-blue-600 mt-2 font-medium">Recording your choice...</p>
          </div>
        )}
      </div>

      {/* Recent Comparisons */}
      {comparisonHistory.length > 0 && (
        <div className="max-w-4xl mx-auto">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">
            <ChartBarIcon className="h-5 w-5 inline mr-2" />
            Recent Comparisons
          </h2>
          <div className="card">
            <div className="space-y-3">
              {comparisonHistory.slice(0, 5).map((comp, index) => {
                // Determine winner and loser based on preference
                const winner = comp.preference === 'book_a' ? comp.book_a :
                              comp.preference === 'book_b' ? comp.book_b : null
                const loser = comp.preference === 'book_a' ? comp.book_b :
                             comp.preference === 'book_b' ? comp.book_a : null

                return (
                  <div key={comp.id || index} className="flex items-center justify-between py-2 border-b last:border-b-0">
                    <div className="text-sm">
                      {comp.preference === 'tie' ? (
                        <>
                          <span className="font-medium">{comp.book_a?.title}</span>
                          <span className="text-gray-500"> tied with </span>
                          <span className="font-medium">{comp.book_b?.title}</span>
                        </>
                      ) : winner && loser ? (
                        <>
                          <span className="font-medium">{winner.title}</span>
                          <span className="text-gray-500"> won against </span>
                          <span>{loser.title}</span>
                        </>
                      ) : (
                        <span className="text-gray-500">Invalid comparison data</span>
                      )}
                    </div>
                    <span className="text-xs text-gray-400">
                      {new Date(comp.created_at).toLocaleDateString()}
                    </span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      )}

    </div>
  )
}

export default ComparisonPage