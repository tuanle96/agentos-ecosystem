-- AgentOS MVP Database Schema
-- Week 1: Core tables for MVP implementation

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

-- ===================================
-- CORE TABLES FOR MVP
-- ===================================

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Agents table
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    capabilities JSONB DEFAULT '[]',
    personality JSONB DEFAULT '{}',
    config JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    framework_preference VARCHAR(50) DEFAULT 'auto', -- 'langchain', 'swarms', 'auto'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Tools table
CREATE TABLE tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(100),
    function_schema JSONB NOT NULL,
    implementation_code TEXT,
    is_active BOOLEAN DEFAULT true,
    version VARCHAR(20) DEFAULT '1.0.0',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Executions table
CREATE TABLE executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    input_text TEXT NOT NULL,
    output_text TEXT,
    framework_used VARCHAR(50),
    tools_used JSONB DEFAULT '[]',
    execution_time_ms INTEGER,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Memories table
CREATE TABLE memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    memory_type VARCHAR(50) NOT NULL, -- 'working', 'episodic', 'semantic'
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    embedding vector(1536), -- OpenAI embedding dimension
    importance_score FLOAT DEFAULT 0.5,
    created_at TIMESTAMP DEFAULT NOW(),
    accessed_at TIMESTAMP DEFAULT NOW()
);

-- Sessions table (for working memory)
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_data JSONB DEFAULT '{}',
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ===================================
-- INDEXES FOR PERFORMANCE
-- ===================================

-- Users indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);

-- Agents indexes
CREATE INDEX idx_agents_user_id ON agents(user_id);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_created_at ON agents(created_at);

-- Tools indexes
CREATE INDEX idx_tools_name ON tools(name);
CREATE INDEX idx_tools_category ON tools(category);
CREATE INDEX idx_tools_active ON tools(is_active);

-- Executions indexes
CREATE INDEX idx_executions_agent_id ON executions(agent_id);
CREATE INDEX idx_executions_user_id ON executions(user_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_started_at ON executions(started_at);

-- Memories indexes
CREATE INDEX idx_memories_agent_id ON memories(agent_id);
CREATE INDEX idx_memories_type ON memories(memory_type);
CREATE INDEX idx_memories_created_at ON memories(created_at);
CREATE INDEX idx_memories_importance ON memories(importance_score);

-- Vector similarity search index
CREATE INDEX idx_memories_embedding ON memories USING ivfflat (embedding vector_cosine_ops);

-- Sessions indexes
CREATE INDEX idx_sessions_agent_id ON sessions(agent_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- ===================================
-- TRIGGERS FOR UPDATED_AT
-- ===================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tools_updated_at BEFORE UPDATE ON tools FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ===================================
-- SAMPLE DATA FOR MVP TESTING
-- ===================================

-- Insert sample tools
INSERT INTO tools (name, description, category, function_schema) VALUES
('web_search', 'Search the web for information using DuckDuckGo', 'search', 
 '{"type": "object", "properties": {"query": {"type": "string"}, "max_results": {"type": "integer", "default": 5}}}'),
('file_operations', 'Read and write files safely', 'file', 
 '{"type": "object", "properties": {"operation": {"type": "string", "enum": ["read", "write"]}, "file_path": {"type": "string"}, "content": {"type": "string"}}}'),
('api_calls', 'Make HTTP API calls', 'network', 
 '{"type": "object", "properties": {"url": {"type": "string"}, "method": {"type": "string", "default": "GET"}, "headers": {"type": "object"}, "data": {"type": "object"}}}'),
('calculations', 'Perform mathematical calculations', 'math', 
 '{"type": "object", "properties": {"expression": {"type": "string"}}}'),
('text_processing', 'Process and analyze text content', 'text', 
 '{"type": "object", "properties": {"text": {"type": "string"}, "operation": {"type": "string", "enum": ["summarize", "analyze", "extract"]}}}');

-- Create a test user
INSERT INTO users (email, password_hash, first_name, last_name) VALUES
('test@agentos.com', crypt('testpassword', gen_salt('bf')), 'Test', 'User');

COMMIT;
