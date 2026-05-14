import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { BookOpenIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline'
import { booksAPI } from '../../services/api'
import { useBooks } from '../../context/BookContext'
import Button from '../../components/common/Button'
import Input from '../../components/common/Input'
import LoadingSpinner from '../../components/common/LoadingSpinner'

function AddBookPage() {
  const navigate = useNavigate()
  const { addBook } = useBooks()
  const [loading, setLoading] = useState(false)
  const [searchResults, setSearchResults] = useState([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [foundBook, setFoundBook] = useState(null)
  const [status, setStatus] = useState('')

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
    setError
  } = useForm()

  const title = watch('title')
  const author = watch('author')

  const handleSearch = async () => {
    if (!title?.trim() || !author?.trim()) {
      setStatus('Please enter both title and author')
      return
    }

    try {
      setSearchLoading(true)
      setStatus('Searching archives...')

      // First search existing books in the archive
      const response = await booksAPI.searchBooks(`${title} ${author}`)
      const existingBooks = response.data.books || []

      // Look for exact or close matches
      const exactMatch = existingBooks.find(book =>
        book.title.toLowerCase().includes(title.toLowerCase()) &&
        book.author.toLowerCase().includes(author.toLowerCase())
      )

      if (exactMatch) {
        setFoundBook(exactMatch)
        setStatus('Book found in archive!')
        setSearchResults([])
      } else {
        setStatus('Book not found in archive, will fetch metadata automatically')
        setFoundBook(null)
        setSearchResults([])
      }
    } catch (error) {
      console.error('Search error:', error)
      setStatus('Search failed, will fetch metadata automatically')
      setFoundBook(null)
    } finally {
      setSearchLoading(false)
    }
  }

  const handleUseFoundBook = (book) => {
    setFoundBook(book)
    setValue('title', book.title)
    setValue('author', book.author)
  }

  const onSubmit = async (data) => {
    try {
      setLoading(true)
      setStatus('Processing...')

      if (foundBook) {
        // Use existing book from archive
        addBook(foundBook)
        setStatus('Added book from archive!')
        setTimeout(() => navigate('/books'), 1000)
      } else {
        // Create new book and fetch metadata automatically
        const bookData = {
          title: data.title,
          author: data.author,
          fetch_metadata: true // Always fetch metadata for new books
        }

        setStatus('Creating book and fetching metadata...')
        const response = await booksAPI.createBook(bookData)
        addBook(response.data)
        setStatus('Book created successfully!')
        setTimeout(() => navigate('/books'), 1000)
      }
    } catch (error) {
      const message = error.response?.data?.message || 'Failed to add book'
      setError('root', { message })
      setStatus('')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Add New Book</h1>
        <p className="mt-1 text-sm text-gray-500">
          Just enter the title and author. We&apos;ll handle the rest automatically.
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div className="card">
          <div className="space-y-4">
            <Input
              label="Title"
              {...register('title', {
                required: 'Title is required'
              })}
              error={errors.title?.message}
              placeholder="Enter book title"
            />

            <Input
              label="Author"
              {...register('author', {
                required: 'Author is required'
              })}
              error={errors.author?.message}
              placeholder="Enter author name"
            />

            {/* Search Button */}
            {title?.trim() && author?.trim() && (
              <div>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={handleSearch}
                  loading={searchLoading}
                  className="w-full"
                >
                  <MagnifyingGlassIcon className="h-4 w-4 mr-2" />
                  Check if book exists in archive
                </Button>
              </div>
            )}

            {/* Status Messages */}
            {status && (
              <div className={`p-3 rounded-md ${
                status.includes('found in archive') ? 'bg-green-50 text-green-700' :
                status.includes('not found') ? 'bg-blue-50 text-blue-700' :
                status.includes('fail') ? 'bg-red-50 text-red-700' :
                'bg-gray-50 text-gray-700'
              }`}>
                {status}
              </div>
            )}

            {/* Found Book Preview */}
            {foundBook && (
              <div className="border rounded-lg p-4 bg-green-50">
                <h3 className="text-sm font-medium text-green-900 mb-2">
                  Found in Archive:
                </h3>
                <div className="space-y-1">
                  <div className="font-medium">{foundBook.title}</div>
                  <div className="text-sm text-gray-600">{foundBook.author}</div>
                  {foundBook.genre && (
                    <div className="text-xs text-gray-500">{foundBook.genre}</div>
                  )}
                  {foundBook.publication_date && (
                    <div className="text-xs text-gray-500">
                      Published: {new Date(foundBook.publication_date).getFullYear()}
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Info Box */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="text-sm text-blue-700">
                <strong>How it works:</strong>
                <ul className="mt-2 list-disc list-inside space-y-1">
                  <li>We&apos;ll first check if this book already exists in our archive</li>
                  <li>If found, we&apos;ll add it to your library instantly</li>
                  <li>If not found, we&apos;ll create it and automatically fetch publication year, genre, and other metadata</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        {errors.root && (
          <div className="rounded-md bg-red-50 p-4">
            <div className="text-sm text-red-700">
              {errors.root.message}
            </div>
          </div>
        )}

        <div className="flex justify-end space-x-3">
          <Button
            type="button"
            variant="secondary"
            onClick={() => navigate('/books')}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            loading={loading}
          >
            {foundBook ? 'Add from Archive' : 'Create & Fetch Metadata'}
          </Button>
        </div>
      </form>
    </div>
  )
}

export default AddBookPage