-- Tool Marketplace Migration
-- Creates tables for tool marketplace functionality

-- Tool Marketplace table
CREATE TABLE IF NOT EXISTS tool_marketplace (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    developer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    tags JSONB DEFAULT '[]'::jsonb,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    latest_version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    function_schema JSONB NOT NULL,
    source_code TEXT NOT NULL,
    documentation TEXT DEFAULT '',
    examples JSONB DEFAULT '[]'::jsonb,
    dependencies JSONB DEFAULT '[]'::jsonb,
    requirements JSONB DEFAULT '{}'::jsonb,
    is_public BOOLEAN DEFAULT false,
    is_verified BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    download_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.0,
    rating_count INTEGER DEFAULT 0,
    security_status VARCHAR(50) DEFAULT 'pending',
    validation_hash VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE NULL,
    
    CONSTRAINT tool_marketplace_name_unique UNIQUE(developer_id, name),
    CONSTRAINT tool_marketplace_rating_check CHECK (rating >= 0.0 AND rating <= 5.0),
    CONSTRAINT tool_marketplace_download_count_check CHECK (download_count >= 0),
    CONSTRAINT tool_marketplace_rating_count_check CHECK (rating_count >= 0)
);

-- Tool Versions table
CREATE TABLE IF NOT EXISTS tool_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID NOT NULL REFERENCES tool_marketplace(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    changelog TEXT NOT NULL,
    function_schema JSONB NOT NULL,
    source_code TEXT NOT NULL,
    dependencies JSONB DEFAULT '[]'::jsonb,
    is_stable BOOLEAN DEFAULT false,
    security_status VARCHAR(50) DEFAULT 'pending',
    validation_hash VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT tool_versions_unique UNIQUE(tool_id, version)
);

-- Tool Installations table
CREATE TABLE IF NOT EXISTS tool_installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tool_id UUID NOT NULL REFERENCES tool_marketplace(id) ON DELETE CASCADE,
    version_id UUID NOT NULL REFERENCES tool_versions(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'installed',
    configuration JSONB DEFAULT '{}'::jsonb,
    installed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE NULL,
    
    CONSTRAINT tool_installations_unique UNIQUE(user_id, tool_id)
);

-- Tool Reviews table
CREATE TABLE IF NOT EXISTS tool_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID NOT NULL REFERENCES tool_marketplace(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255) NOT NULL,
    comment TEXT NOT NULL,
    is_public BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT tool_reviews_unique UNIQUE(tool_id, user_id)
);

-- Tool Usage Statistics table
CREATE TABLE IF NOT EXISTS tool_usage_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID NOT NULL REFERENCES tool_marketplace(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    execution_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    total_time_ms BIGINT DEFAULT 0,
    average_time_ms DECIMAL(10,2) DEFAULT 0.0,
    last_executed_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT tool_usage_stats_unique UNIQUE(tool_id, user_id),
    CONSTRAINT tool_usage_stats_counts_check CHECK (
        execution_count >= 0 AND 
        success_count >= 0 AND 
        error_count >= 0 AND
        success_count + error_count <= execution_count
    )
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_developer_id ON tool_marketplace(developer_id);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_category ON tool_marketplace(category);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_is_public ON tool_marketplace(is_public);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_is_verified ON tool_marketplace(is_verified);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_is_active ON tool_marketplace(is_active);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_rating ON tool_marketplace(rating DESC);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_download_count ON tool_marketplace(download_count DESC);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_created_at ON tool_marketplace(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_updated_at ON tool_marketplace(updated_at DESC);

-- Full-text search index for tool search
CREATE INDEX IF NOT EXISTS idx_tool_marketplace_search ON tool_marketplace 
USING gin(to_tsvector('english', name || ' ' || display_name || ' ' || description));

-- Tool versions indexes
CREATE INDEX IF NOT EXISTS idx_tool_versions_tool_id ON tool_versions(tool_id);
CREATE INDEX IF NOT EXISTS idx_tool_versions_created_at ON tool_versions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tool_versions_is_stable ON tool_versions(is_stable);

-- Tool installations indexes
CREATE INDEX IF NOT EXISTS idx_tool_installations_user_id ON tool_installations(user_id);
CREATE INDEX IF NOT EXISTS idx_tool_installations_tool_id ON tool_installations(tool_id);
CREATE INDEX IF NOT EXISTS idx_tool_installations_status ON tool_installations(status);
CREATE INDEX IF NOT EXISTS idx_tool_installations_installed_at ON tool_installations(installed_at DESC);

-- Tool reviews indexes
CREATE INDEX IF NOT EXISTS idx_tool_reviews_tool_id ON tool_reviews(tool_id);
CREATE INDEX IF NOT EXISTS idx_tool_reviews_user_id ON tool_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_tool_reviews_rating ON tool_reviews(rating DESC);
CREATE INDEX IF NOT EXISTS idx_tool_reviews_created_at ON tool_reviews(created_at DESC);

-- Tool usage stats indexes
CREATE INDEX IF NOT EXISTS idx_tool_usage_stats_tool_id ON tool_usage_stats(tool_id);
CREATE INDEX IF NOT EXISTS idx_tool_usage_stats_user_id ON tool_usage_stats(user_id);
CREATE INDEX IF NOT EXISTS idx_tool_usage_stats_execution_count ON tool_usage_stats(execution_count DESC);

-- Functions for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for automatic timestamp updates
CREATE TRIGGER update_tool_marketplace_updated_at 
    BEFORE UPDATE ON tool_marketplace 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tool_installations_updated_at 
    BEFORE UPDATE ON tool_installations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tool_reviews_updated_at 
    BEFORE UPDATE ON tool_reviews 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tool_usage_stats_updated_at 
    BEFORE UPDATE ON tool_usage_stats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update tool rating when reviews are added/updated
CREATE OR REPLACE FUNCTION update_tool_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tool_marketplace 
    SET 
        rating = (
            SELECT COALESCE(AVG(rating::DECIMAL), 0.0) 
            FROM tool_reviews 
            WHERE tool_id = COALESCE(NEW.tool_id, OLD.tool_id) AND is_public = true
        ),
        rating_count = (
            SELECT COUNT(*) 
            FROM tool_reviews 
            WHERE tool_id = COALESCE(NEW.tool_id, OLD.tool_id) AND is_public = true
        )
    WHERE id = COALESCE(NEW.tool_id, OLD.tool_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ language 'plpgsql';

-- Trigger to automatically update tool rating
CREATE TRIGGER update_tool_rating_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tool_reviews
    FOR EACH ROW EXECUTE FUNCTION update_tool_rating();

-- Insert sample tool categories
INSERT INTO tool_marketplace (
    developer_id, name, display_name, description, category, tags,
    function_schema, source_code, documentation, is_public, is_verified
) VALUES 
(
    (SELECT id FROM users LIMIT 1),
    'web_scraper',
    'Web Scraper Tool',
    'A powerful web scraping tool that can extract data from websites',
    'web',
    '["scraping", "web", "data", "extraction"]'::jsonb,
    '{
        "name": "web_scraper",
        "description": "Scrape data from websites",
        "parameters": {
            "type": "object",
            "properties": {
                "url": {"type": "string", "description": "URL to scrape"},
                "selector": {"type": "string", "description": "CSS selector for data extraction"}
            },
            "required": ["url"]
        }
    }'::jsonb,
    'def web_scraper(url, selector=None):
    """Scrape data from a website"""
    import requests
    from bs4 import BeautifulSoup
    
    response = requests.get(url)
    soup = BeautifulSoup(response.content, "html.parser")
    
    if selector:
        elements = soup.select(selector)
        return [elem.get_text().strip() for elem in elements]
    else:
        return soup.get_text().strip()',
    '# Web Scraper Tool

This tool allows you to scrape data from websites using CSS selectors.

## Usage

```python
result = web_scraper("https://example.com", "h1")
```

## Parameters

- `url`: The URL to scrape (required)
- `selector`: CSS selector for data extraction (optional)',
    true,
    true
) ON CONFLICT (developer_id, name) DO NOTHING;

-- Create initial version for sample tool
INSERT INTO tool_versions (
    tool_id, version, changelog, function_schema, source_code, is_stable
) 
SELECT 
    id, 
    '1.0.0', 
    'Initial release of web scraper tool',
    function_schema,
    source_code,
    true
FROM tool_marketplace 
WHERE name = 'web_scraper'
ON CONFLICT (tool_id, version) DO NOTHING;
