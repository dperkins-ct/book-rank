import { format, formatDistanceToNow } from 'date-fns'

export function formatDate(date, formatStr = 'MMM dd, yyyy') {
  if (!date) return ''
  return format(new Date(date), formatStr)
}

export function formatRelativeTime(date) {
  if (!date) return ''
  return formatDistanceToNow(new Date(date), { addSuffix: true })
}

export function formatRating(rating) {
  if (rating == null) return 'Not rated'
  return `${rating.toFixed(1)}/10`
}

export function formatNumber(num) {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(1)}M`
  }
  if (num >= 1000) {
    return `${(num / 1000).toFixed(1)}K`
  }
  return num.toString()
}

export function truncateText(text, maxLength = 100) {
  if (!text) return ''
  if (text.length <= maxLength) return text
  return `${text.substring(0, maxLength)}...`
}

export function getRatingColor(rating) {
  if (rating >= 8) return 'text-green-600'
  if (rating >= 6) return 'text-yellow-600'
  if (rating >= 4) return 'text-orange-600'
  return 'text-red-600'
}

export function getRatingBadgeColor(rating) {
  if (rating >= 8) return 'bg-green-100 text-green-800'
  if (rating >= 6) return 'bg-yellow-100 text-yellow-800'
  if (rating >= 4) return 'bg-orange-100 text-orange-800'
  return 'bg-red-100 text-red-800'
}