-- Week 2 Database Migration: Tool Executions Table
-- AgentOS MVP - Week 2 Enhancement

-- Create tool_executions table for tracking tool execution history
CREATE TABLE IF NOT EXISTS tool_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    tool_name VARCHAR(100) NOT NULL,
    request_data JSONB NOT NULL,
    response_data JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    execution_time DECIMAL(10,6) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tool_executions_user_id ON tool_executions(user_id);
CREATE INDEX IF NOT EXISTS idx_tool_executions_agent_id ON tool_executions(agent_id);
CREATE INDEX IF NOT EXISTS idx_tool_executions_tool_name ON tool_executions(tool_name);
CREATE INDEX IF NOT EXISTS idx_tool_executions_status ON tool_executions(status);
CREATE INDEX IF NOT EXISTS idx_tool_executions_created_at ON tool_executions(created_at);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_tool_executions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_tool_executions_updated_at
    BEFORE UPDATE ON tool_executions
    FOR EACH ROW
    EXECUTE FUNCTION update_tool_executions_updated_at();

-- Add sample tool execution data for testing
INSERT INTO tool_executions (user_id, tool_name, request_data, response_data, status, execution_time)
SELECT 
    u.id,
    'web_search',
    '{"query": "test search", "max_results": 5}',
    '{"query": "test search", "results": ["result1", "result2"], "count": 2}',
    'completed',
    0.125
FROM users u
LIMIT 1;

-- Create working_memory_sessions table for Redis backup
CREATE TABLE IF NOT EXISTS working_memory_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    session_id UUID NOT NULL,
    memory_data JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for working memory sessions
CREATE INDEX IF NOT EXISTS idx_working_memory_agent_id ON working_memory_sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_working_memory_session_id ON working_memory_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_working_memory_expires_at ON working_memory_sessions(expires_at);

-- Create trigger for working memory sessions updated_at
CREATE TRIGGER trigger_working_memory_sessions_updated_at
    BEFORE UPDATE ON working_memory_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_tool_executions_updated_at();

-- Update memories table to support different memory types
ALTER TABLE memories 
ADD COLUMN IF NOT EXISTS memory_type VARCHAR(20) DEFAULT 'episodic',
ADD COLUMN IF NOT EXISTS session_id UUID,
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;

-- Create index for memory types
CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(memory_type);
CREATE INDEX IF NOT EXISTS idx_memories_session_id ON memories(session_id);

-- Add agent capabilities tracking table
CREATE TABLE IF NOT EXISTS agent_capabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    capability_name VARCHAR(100) NOT NULL,
    capability_config JSONB DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    resource_cost INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(agent_id, capability_name)
);

-- Create indexes for agent capabilities
CREATE INDEX IF NOT EXISTS idx_agent_capabilities_agent_id ON agent_capabilities(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_capabilities_name ON agent_capabilities(capability_name);
CREATE INDEX IF NOT EXISTS idx_agent_capabilities_enabled ON agent_capabilities(is_enabled);

-- Create trigger for agent capabilities updated_at
CREATE TRIGGER trigger_agent_capabilities_updated_at
    BEFORE UPDATE ON agent_capabilities
    FOR EACH ROW
    EXECUTE FUNCTION update_tool_executions_updated_at();

-- Framework integration tracking
CREATE TABLE IF NOT EXISTS framework_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    framework_name VARCHAR(50) NOT NULL,
    framework_config JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for framework integrations
CREATE INDEX IF NOT EXISTS idx_framework_integrations_agent_id ON framework_integrations(agent_id);
CREATE INDEX IF NOT EXISTS idx_framework_integrations_framework ON framework_integrations(framework_name);
CREATE INDEX IF NOT EXISTS idx_framework_integrations_status ON framework_integrations(status);

-- Create trigger for framework integrations updated_at
CREATE TRIGGER trigger_framework_integrations_updated_at
    BEFORE UPDATE ON framework_integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_tool_executions_updated_at();

-- Performance monitoring table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type VARCHAR(50) NOT NULL, -- api_response, tool_execution, memory_operation
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(10,6) NOT NULL,
    metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance metrics
CREATE INDEX IF NOT EXISTS idx_performance_metrics_type ON performance_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_name ON performance_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_recorded_at ON performance_metrics(recorded_at);

-- Insert sample performance metrics
INSERT INTO performance_metrics (metric_type, metric_name, metric_value, metadata)
VALUES 
    ('api_response', 'agent_creation', 0.001, '{"endpoint": "/api/v1/agents", "method": "POST"}'),
    ('api_response', 'agent_list', 0.001, '{"endpoint": "/api/v1/agents", "method": "GET"}'),
    ('tool_execution', 'web_search', 0.125, '{"tool": "web_search", "status": "completed"}'),
    ('memory_operation', 'working_memory_get', 0.002, '{"operation": "get", "cache": "redis"}');

-- Create view for agent statistics
CREATE OR REPLACE VIEW agent_statistics AS
SELECT 
    a.id as agent_id,
    a.name,
    a.status,
    a.framework_preference,
    COUNT(DISTINCT ac.capability_name) as capabilities_count,
    COUNT(DISTINCT te.id) as tool_executions_count,
    COUNT(DISTINCT e.id) as total_executions_count,
    MAX(e.created_at) as last_execution_at,
    a.created_at,
    a.updated_at
FROM agents a
LEFT JOIN agent_capabilities ac ON a.id = ac.agent_id AND ac.is_enabled = true
LEFT JOIN tool_executions te ON a.id = te.agent_id
LEFT JOIN executions e ON a.id = e.agent_id
GROUP BY a.id, a.name, a.status, a.framework_preference, a.created_at, a.updated_at;

-- Create view for performance dashboard
CREATE OR REPLACE VIEW performance_dashboard AS
SELECT 
    metric_type,
    metric_name,
    COUNT(*) as measurement_count,
    AVG(metric_value) as avg_value,
    MIN(metric_value) as min_value,
    MAX(metric_value) as max_value,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY metric_value) as p95_value,
    MAX(recorded_at) as last_recorded_at
FROM performance_metrics
WHERE recorded_at >= NOW() - INTERVAL '24 hours'
GROUP BY metric_type, metric_name
ORDER BY metric_type, avg_value DESC;

-- Week 2 Migration Complete
-- Summary of changes:
-- 1. Added tool_executions table for tracking tool execution history
-- 2. Added working_memory_sessions table for Redis backup
-- 3. Enhanced memories table with memory types and sessions
-- 4. Added agent_capabilities table for capability tracking
-- 5. Added framework_integrations table for framework management
-- 6. Added performance_metrics table for monitoring
-- 7. Created views for agent statistics and performance dashboard
-- 8. Added appropriate indexes and triggers for all new tables

COMMENT ON TABLE tool_executions IS 'Week 2: Tool execution history and results';
COMMENT ON TABLE working_memory_sessions IS 'Week 2: Working memory session backup';
COMMENT ON TABLE agent_capabilities IS 'Week 2: Agent capability tracking and management';
COMMENT ON TABLE framework_integrations IS 'Week 2: AI framework integration status';
COMMENT ON TABLE performance_metrics IS 'Week 2: Performance monitoring and metrics';
COMMENT ON VIEW agent_statistics IS 'Week 2: Comprehensive agent statistics view';
COMMENT ON VIEW performance_dashboard IS 'Week 2: Performance monitoring dashboard';
