import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  UserGroupIcon,
  PlusIcon,
  UserCircleIcon,
  BookOpenIcon,
  TrophyIcon,
  MagnifyingGlassIcon,
  XMarkIcon
} from '@heroicons/react/24/outline'
import { useForm } from 'react-hook-form'
import { friendsAPI } from '../../services/api'
import Button from '../../components/common/Button'
import Input from '../../components/common/Input'
import LoadingSpinner from '../../components/common/LoadingSpinner'
import Modal from '../../components/common/Modal'
import { formatRelativeTime, formatRating, getRatingBadgeColor } from '../../utils/formatters'

function FriendsPage() {
  const [friends, setFriends] = useState([])
  const [friendActivity, setFriendActivity] = useState({})
  const [friendRankings, setFriendRankings] = useState({})
  const [loading, setLoading] = useState(true)
  const [showAddFriend, setShowAddFriend] = useState(false)
  const [expandedFriend, setExpandedFriend] = useState(null)

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setError
  } = useForm()

  useEffect(() => {
    loadFriends()
  }, [])

  const loadFriends = async () => {
    try {
      setLoading(true)
      const response = await friendsAPI.getFriends()
      setFriends(response.data.friends || [])
    } catch (error) {
      console.error('Error loading friends:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadFriendActivity = async (friendId) => {
    if (friendActivity[friendId]) return // Already loaded

    try {
      const response = await friendsAPI.getFriendActivity(friendId, { limit: 10 })
      setFriendActivity(prev => ({
        ...prev,
        [friendId]: response.data.activities || []
      }))
    } catch (error) {
      console.error('Error loading friend activity:', error)
    }
  }

  const loadFriendRankings = async (friendId) => {
    if (friendRankings[friendId]) return // Already loaded

    try {
      const response = await friendsAPI.getFriendRankings(friendId, { limit: 5 })
      setFriendRankings(prev => ({
        ...prev,
        [friendId]: response.data.rankings || []
      }))
    } catch (error) {
      console.error('Error loading friend rankings:', error)
    }
  }

  const handleAddFriend = async (data) => {
    try {
      await friendsAPI.addFriend(data.username)
      await loadFriends()
      setShowAddFriend(false)
      reset()
    } catch (error) {
      const message = error.response?.data?.message || 'Failed to add friend'
      setError('username', { message })
    }
  }

  const handleRemoveFriend = async (friendId) => {
    try {
      await friendsAPI.removeFriend(friendId)
      setFriends(prev => prev.filter(friend => friend.id !== friendId))
    } catch (error) {
      console.error('Error removing friend:', error)
    }
  }

  const toggleFriendDetails = async (friend) => {
    if (expandedFriend === friend.id) {
      setExpandedFriend(null)
    } else {
      setExpandedFriend(friend.id)
      await Promise.all([
        loadFriendActivity(friend.id),
        loadFriendRankings(friend.id)
      ])
    }
  }

  const FriendCard = ({ friend }) => {
    const isExpanded = expandedFriend === friend.id
    const activity = friendActivity[friend.id] || []
    const rankings = friendRankings[friend.id] || []

    return (
      <div className="card">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <UserCircleIcon className="h-12 w-12 text-gray-400" />
            <div>
              <h3 className="text-lg font-semibold text-gray-900">
                {friend.username}
              </h3>
              <p className="text-sm text-gray-500">
                {friend.totalBooks || 0} books • {friend.totalComparisons || 0} comparisons
              </p>
            </div>
          </div>

          <div className="flex space-x-2">
            <Button
              size="sm"
              variant="secondary"
              onClick={() => toggleFriendDetails(friend)}
            >
              {isExpanded ? 'Hide Details' : 'View Activity'}
            </Button>
            <Button
              size="sm"
              variant="danger"
              onClick={() => handleRemoveFriend(friend.id)}
            >
              <XMarkIcon className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {isExpanded && (
          <div className="mt-6 pt-6 border-t">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Recent Activity */}
              <div>
                <h4 className="font-medium text-gray-900 mb-3">Recent Activity</h4>
                {activity.length > 0 ? (
                  <div className="space-y-2">
                    {activity.map((item, index) => (
                      <div key={index} className="text-sm p-2 bg-gray-50 rounded">
                        <p className="text-gray-700">{item.description}</p>
                        <p className="text-xs text-gray-500 mt-1">
                          {formatRelativeTime(item.created_at)}
                        </p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">No recent activity</p>
                )}
              </div>

              {/* Top Rankings */}
              <div>
                <h4 className="font-medium text-gray-900 mb-3">Top Ranked Books</h4>
                {rankings.length > 0 ? (
                  <div className="space-y-2">
                    {rankings.map((book, index) => (
                      <div key={book.id} className="flex items-center space-x-3 text-sm">
                        <span className="font-bold text-gray-900 w-6">
                          #{index + 1}
                        </span>
                        <div className="flex-1">
                          <p className="font-medium text-gray-900">{book.title}</p>
                          <p className="text-gray-600">{book.author}</p>
                        </div>
                        {book.rating && (
                          <span className={`badge ${getRatingBadgeColor(book.rating)}`}>
                            {formatRating(book.rating)}
                          </span>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">No rankings yet</p>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Friends</h1>
          <p className="mt-1 text-sm text-gray-500">
            Connect with other readers and see what they're enjoying
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <Button onClick={() => setShowAddFriend(true)}>
            <PlusIcon className="h-4 w-4 mr-2" />
            Add Friend
          </Button>
        </div>
      </div>

      {/* Friends List */}
      {loading ? (
        <div className="flex justify-center py-8">
          <LoadingSpinner size="lg" />
        </div>
      ) : friends.length > 0 ? (
        <div className="space-y-4">
          {friends.map((friend) => (
            <FriendCard key={friend.id} friend={friend} />
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <UserGroupIcon className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No friends yet</h3>
          <p className="mt-1 text-sm text-gray-500">
            Add friends to see their book rankings and get social recommendations.
          </p>
          <div className="mt-6">
            <Button onClick={() => setShowAddFriend(true)}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Add Your First Friend
            </Button>
          </div>
        </div>
      )}

      {/* Add Friend Modal */}
      <Modal
        isOpen={showAddFriend}
        onClose={() => {
          setShowAddFriend(false)
          reset()
        }}
        title="Add Friend"
      >
        <form onSubmit={handleSubmit(handleAddFriend)} className="space-y-4">
          <Input
            label="Username"
            placeholder="Enter your friend's username"
            {...register('username', {
              required: 'Username is required',
              minLength: {
                value: 3,
                message: 'Username must be at least 3 characters'
              }
            })}
            error={errors.username?.message}
          />

          <div className="flex justify-end space-x-3">
            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                setShowAddFriend(false)
                reset()
              }}
            >
              Cancel
            </Button>
            <Button type="submit">
              Add Friend
            </Button>
          </div>
        </form>
      </Modal>

      {/* Social Features Info */}
      <div className="bg-primary-50 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-primary-900 mb-2">
          Social Features Coming Soon
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-primary-700">
          <div className="flex items-start space-x-2">
            <BookOpenIcon className="h-5 w-5 text-primary-600 mt-0.5" />
            <div>
              <p className="font-medium">Reading Challenges</p>
              <p>Compete with friends in monthly reading goals</p>
            </div>
          </div>
          <div className="flex items-start space-x-2">
            <TrophyIcon className="h-5 w-5 text-primary-600 mt-0.5" />
            <div>
              <p className="font-medium">Leaderboards</p>
              <p>See who has the best book recommendations</p>
            </div>
          </div>
          <div className="flex items-start space-x-2">
            <MagnifyingGlassIcon className="h-5 w-5 text-primary-600 mt-0.5" />
            <div>
              <p className="font-medium">Social Discovery</p>
              <p>Find books based on what your friends love</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default FriendsPage