-- Migration: Add indexes for optimized comparison queries
-- Date: 2026-05-14
-- Purpose: Improve performance of pending comparisons and comparison lookups

-- Index on books.created_by for efficient user book filtering
CREATE INDEX IF NOT EXISTS idx_books_created_by ON books(created_by);

-- Composite index on comparisons for efficient duplicate checking
CREATE INDEX IF NOT EXISTS idx_comparisons_user_books ON comparisons(user_id, book_a_id, book_b_id);

-- Index on comparisons.user_id for user-specific queries
CREATE INDEX IF NOT EXISTS idx_comparisons_user_id ON comparisons(user_id);

-- Index on comparisons created_at for history ordering
CREATE INDEX IF NOT EXISTS idx_comparisons_created_at ON comparisons(created_at DESC);

-- Composite index for book-specific comparison queries
CREATE INDEX IF NOT EXISTS idx_comparisons_book_lookups ON comparisons(user_id, book_a_id);
CREATE INDEX IF NOT EXISTS idx_comparisons_book_lookups_alt ON comparisons(user_id, book_b_id);

-- Index on rankings for user-specific queries
CREATE INDEX IF NOT EXISTS idx_rankings_user_book ON rankings(user_id, book_id);
CREATE INDEX IF NOT EXISTS idx_rankings_user_rating ON rankings(user_id, current_rating DESC);