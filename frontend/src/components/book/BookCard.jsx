import React from 'react'
import { Link } from 'react-router-dom'
import {
  BookOpenIcon,
  StarIcon,
  EyeIcon,
  PencilIcon,
  TrashIcon
} from '@heroicons/react/24/outline'
import { StarIcon as StarSolidIcon } from '@heroicons/react/24/solid'
import Button from '../common/Button'
import { formatDate, formatRating, getRatingBadgeColor } from '../../utils/formatters'

function BookCard({ book, onEdit, onDelete, showActions = false }) {
  const coverImage = book.metadata?.find(m => m.additional_data?.cover_url)?.additional_data?.cover_url

  return (
    <div className="card hover:shadow-md transition-shadow">
      <div className="flex space-x-4">
        {/* Book Cover */}
        <div className="flex-shrink-0">
          {coverImage ? (
            <img
              src={coverImage}
              alt={`${book.title} cover`}
              className="h-24 w-16 object-cover rounded"
            />
          ) : (
            <div className="h-24 w-16 bg-gray-200 rounded flex items-center justify-center">
              <BookOpenIcon className="h-8 w-8 text-gray-400" />
            </div>
          )}
        </div>

        {/* Book Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <Link
                to={`/books/${book.id}`}
                className="text-lg font-semibold text-gray-900 hover:text-primary-600 line-clamp-2"
              >
                {book.title}
              </Link>
              <p className="text-sm text-gray-600 mt-1">{book.author}</p>
              {book.genre && (
                <span className="inline-block bg-gray-100 text-gray-700 text-xs px-2 py-1 rounded mt-2">
                  {book.genre}
                </span>
              )}
            </div>

            {/* Rating */}
            <div className="flex items-center space-x-2 ml-4">
              {book.average_rating ? (
                <>
                  <div className="flex items-center">
                    <StarSolidIcon className="h-4 w-4 text-yellow-400" />
                    <span className="ml-1 text-sm font-medium">
                      {formatRating(book.average_rating)}
                    </span>
                  </div>
                  <span className="text-xs text-gray-500">
                    ({book.total_ratings} rating{book.total_ratings !== 1 ? 's' : ''})
                  </span>
                </>
              ) : (
                <span className="text-xs text-gray-500">Not rated</span>
              )}
            </div>
          </div>

          {/* Description */}
          {book.description && (
            <p className="text-sm text-gray-600 mt-2 line-clamp-2">
              {book.description}
            </p>
          )}

          {/* Meta Info */}
          <div className="flex items-center justify-between mt-3">
            <div className="flex items-center space-x-4 text-xs text-gray-500">
              {book.publication_date && (
                <span>Published {formatDate(book.publication_date, 'yyyy')}</span>
              )}
              <span>Added {formatDate(book.created_at)}</span>
            </div>

            {/* Actions */}
            {showActions && (
              <div className="flex items-center space-x-2">
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => onEdit?.(book)}
                >
                  <PencilIcon className="h-4 w-4" />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => onDelete?.(book)}
                  className="text-red-600 hover:text-red-700"
                >
                  <TrashIcon className="h-4 w-4" />
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default BookCard