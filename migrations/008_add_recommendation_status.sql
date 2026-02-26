-- Add status column to recommendations table
ALTER TABLE recommendations ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending';

-- Index for filtering by status
CREATE INDEX IF NOT EXISTS idx_recommendations_status ON recommendations(status);
