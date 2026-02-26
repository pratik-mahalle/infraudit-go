-- Add configuration column to resources table for drift detection
-- This column stores the full JSON configuration of cloud resources
ALTER TABLE resources ADD COLUMN configuration TEXT DEFAULT '';
