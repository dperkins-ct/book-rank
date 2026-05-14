import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  MagnifyingGlassIcon,
  FunnelIcon,
  PlusIcon,
  AdjustmentsHorizontalIcon,
  BookOpenIcon
} from '@heroicons/react/24/outline'
import { useBooks } from '../../context/BookContext'
import { booksAPI } from '../../services/api'
import BookCard from '../../components/book/BookCard'
import Button from '../../components/common/Button'
import Input from '../../components/common/Input'
import LoadingSpinner from '../../components/common/LoadingSpinner'

function BookLibraryPage() {
  const {
    books,
    loading,
    pagination,
    filters,
    setLoading,
    setBooks,
    setFilters
  } = useBooks()

  const [showFilters, setShowFilters] = useState(false)

  useEffect(() => {
    loadBooks()
  }, [filters, pagination.page])

  const loadBooks = async () => {
    try {
      setLoading(true)
      const params = {
        page: pagination.page,
        limit: pagination.pageSize,
        ...filters
      }

      // Remove empty values
      Object.keys(params).forEach(key => {
        if (params[key] === '' || params[key] === null || params[key] === undefined) {
          delete params[key]
        }
      })

      const response = await booksAPI.getBooks(params)
      setBooks(response.data)
    } catch (error) {
      console.error('Error loading books:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = (e) => {
    setFilters({ search: e.target.value })
  }

  const handleFilterChange = (key, value) => {
    setFilters({ [key]: value })
  }

  const handleSortChange = (sort, order) => {
    setFilters({ sort, order })
  }


  const sortOptions = [
    { value: 'created_at|desc', label: 'Recently Added' },
    { value: 'created_at|asc', label: 'Oldest First' },
    { value: 'title|asc', label: 'Title A-Z' },
    { value: 'title|desc', label: 'Title Z-A' },
    { value: 'author|asc', label: 'Author A-Z' },
    { value: 'rating|desc', label: 'Highest Rated' },
    { value: 'rating|asc', label: 'Lowest Rated' },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Book Archive</h1>
          <p className="mt-1 text-sm text-gray-500">
            Browse {pagination.total} book{pagination.total !== 1 ? 's' : ''} in our collection
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <Button as={Link} to="/books/add">
            <PlusIcon className="h-4 w-4 mr-2" />
            Add Book
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex-1">
          <div className="relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search books by title, author, or description..."
              value={filters.search}
              onChange={handleSearch}
              className="input pl-10"
            />
          </div>
        </div>

        <div className="flex gap-2">
          <Button
            variant="secondary"
            onClick={() => setShowFilters(!showFilters)}
          >
            <FunnelIcon className="h-4 w-4 mr-2" />
            Filters
          </Button>

          <select
            value={`${filters.sort}|${filters.order}`}
            onChange={(e) => {
              const [sort, order] = e.target.value.split('|')
              handleSortChange(sort, order)
            }}
            className="input"
          >
            {sortOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Advanced Filters */}
      {showFilters && (
        <div className="card">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Input
              label="Author"
              value={filters.author}
              onChange={(e) => handleFilterChange('author', e.target.value)}
              placeholder="Filter by author"
            />
            <Input
              label="Genre"
              value={filters.genre}
              onChange={(e) => handleFilterChange('genre', e.target.value)}
              placeholder="Filter by genre"
            />
            <div className="grid grid-cols-2 gap-2">
              <Input
                label="Min Rating"
                type="number"
                min="1"
                max="10"
                value={filters.min_rating || ''}
                onChange={(e) => handleFilterChange('min_rating', e.target.value)}
                placeholder="1"
              />
              <Input
                label="Max Rating"
                type="number"
                min="1"
                max="10"
                value={filters.max_rating || ''}
                onChange={(e) => handleFilterChange('max_rating', e.target.value)}
                placeholder="10"
              />
            </div>
          </div>
        </div>
      )}

      {/* Books Grid */}
      {loading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner size="lg" />
        </div>
      ) : books.length > 0 ? (
        <div className="space-y-4">
          {books.map((book) => (
            <BookCard
              key={book.id}
              book={book}
              showActions={false}
            />
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <BookOpenIcon className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No books found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {filters.search || filters.author || filters.genre
              ? "Try adjusting your search criteria"
              : "Get started by adding your first book."}
          </p>
          <div className="mt-6">
            <Button as={Link} to="/books/add">
              <PlusIcon className="h-4 w-4 mr-2" />
              Add Book
            </Button>
          </div>
        </div>
      )}

      {/* Pagination */}
      {pagination.totalPages > 1 && (
        <div className="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6">
          <div className="flex flex-1 justify-between sm:hidden">
            <Button
              variant="secondary"
              disabled={pagination.page === 1}
              onClick={() => handleFilterChange('page', pagination.page - 1)}
            >
              Previous
            </Button>
            <Button
              variant="secondary"
              disabled={pagination.page === pagination.totalPages}
              onClick={() => handleFilterChange('page', pagination.page + 1)}
            >
              Next
            </Button>
          </div>
          <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
            <div>
              <p className="text-sm text-gray-700">
                Showing{' '}
                <span className="font-medium">
                  {(pagination.page - 1) * pagination.pageSize + 1}
                </span>{' '}
                to{' '}
                <span className="font-medium">
                  {Math.min(pagination.page * pagination.pageSize, pagination.total)}
                </span>{' '}
                of{' '}
                <span className="font-medium">{pagination.total}</span>{' '}
                results
              </p>
            </div>
            <div>
              <nav className="isolate inline-flex -space-x-px rounded-md shadow-sm">
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={pagination.page === 1}
                  onClick={() => handleFilterChange('page', pagination.page - 1)}
                >
                  Previous
                </Button>
                {/* Page numbers would go here */}
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={pagination.page === pagination.totalPages}
                  onClick={() => handleFilterChange('page', pagination.page + 1)}
                >
                  Next
                </Button>
              </nav>
            </div>
          </div>
        </div>
      )}

    </div>
  )
}

export default BookLibraryPage