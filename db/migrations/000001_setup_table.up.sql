-- Create 'urls' table for storing shortened URLs
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,  -- Auto-incrementing primary key
    short_code VARCHAR(10) UNIQUE NOT NULL,  -- Shortened URL code (max 10 chars)
    long_url TEXT UNIQUE NOT NULL,  -- Original long URL
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- Timestamp when URL was shortened
);

-- Add an index on short_code for faster lookups
CREATE INDEX idx_short_code ON urls(short_code);
CREATE INDEX idx_long_url ON urls(long_url);

-- Add an index on created_at to optimize queries based on time
CREATE INDEX idx_created_at ON urls(created_at);

-- Add comments for documentation
COMMENT ON TABLE urls IS 'Table to store shortened URLs and their mappings to long URLs';
COMMENT ON COLUMN urls.id IS 'Primary key - unique identifier for each URL';
COMMENT ON COLUMN urls.short_code IS 'Unique short URL code generated for each long URL';
COMMENT ON COLUMN urls.long_url IS 'Original long URL that is mapped to the short_code';
COMMENT ON COLUMN urls.created_at IS 'Timestamp indicating when the URL was shortened';
