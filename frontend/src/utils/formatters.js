import { format, formatDistanceToNow } from 'date-fns'

export function formatDate(date, formatStr = 'MMM dd, yyyy') {
  if (!date) return ''
  return format(new Date(date), formatStr)
}

export function formatRelativeTime(date) {
  if (!date) return ''
  return formatDistanceToNow(new Date(date), { addSuffix: true })
}

// Convert user's personal ranking position to 0-10 scale
export function convertPersonalRankingToScore(rankings, currentRanking) {
  if (!rankings || rankings.length <= 1) {
    return 5.0 // Default middle score for single book
  }

  // Find position of current book in user's rankings (sorted by ELO DESC)
  const position = rankings.findIndex(r => r.book_id === currentRanking.book_id)
  if (position === -1) return 5.0

  // Convert position to 0-10 scale (0-based position, inverted so higher is better)
  const totalBooks = rankings.length
  const score = 10.0 - (position / (totalBooks - 1)) * 10.0

  return Math.round(score * 10) / 10
}

// Convert ELO rating to BookRank score (0-10 scale) - DEPRECATED but kept for compatibility
export function convertELOToBookRank(eloRating) {
  const minELO = 800
  const maxELO = 2200

  if (eloRating <= minELO) return 0.0
  if (eloRating >= maxELO) return 10.0

  const ratio = (eloRating - minELO) / (maxELO - minELO)
  const bookRankScore = ratio * 10.0

  return Math.round(bookRankScore * 10) / 10
}

export function formatRating(eloRating) {
  if (eloRating == null) return 'Not rated'
  const bookRankScore = convertELOToBookRank(eloRating)
  return `${bookRankScore.toFixed(1)}/10`
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

export function getRatingColor(eloRating) {
  const rating = convertELOToBookRank(eloRating)
  if (rating >= 8) return 'text-green-600'
  if (rating >= 6) return 'text-yellow-600'
  if (rating >= 4) return 'text-orange-600'
  return 'text-red-600'
}

export function getRatingBadgeColor(eloRating) {
  const rating = convertELOToBookRank(eloRating)
  if (rating >= 9) return 'bg-purple-100 text-purple-800' // Masterpiece
  if (rating >= 8) return 'bg-green-100 text-green-800'   // Excellent
  if (rating >= 7) return 'bg-blue-100 text-blue-800'     // Great
  if (rating >= 6) return 'bg-teal-100 text-teal-800'     // Good
  if (rating >= 5) return 'bg-yellow-100 text-yellow-800' // Average
  if (rating >= 4) return 'bg-orange-100 text-orange-800' // Below Average
  return 'bg-red-100 text-red-800' // Poor/Bad/Awful
}

export function getRatingLabel(eloRating) {
  const rating = convertELOToBookRank(eloRating)
  if (rating >= 9) return 'Masterpiece'
  if (rating >= 8) return 'Excellent'
  if (rating >= 7) return 'Great'
  if (rating >= 6) return 'Good'
  if (rating >= 5) return 'Average'
  if (rating >= 4) return 'Below Average'
  if (rating >= 3) return 'Poor'
  if (rating >= 2) return 'Bad'
  if (rating >= 1) return 'Awful'
  return 'Unrated'
}