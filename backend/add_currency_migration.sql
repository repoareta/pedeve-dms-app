-- Migration: Add currency field to companies table
-- Date: 2025-01-XX
-- Description: Add currency field to support IDR and USD

-- Check if column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'companies' 
        AND column_name = 'currency'
    ) THEN
        ALTER TABLE companies 
        ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'IDR';
        
        RAISE NOTICE 'Column currency added to companies table';
    ELSE
        RAISE NOTICE 'Column currency already exists in companies table';
    END IF;
END $$;
