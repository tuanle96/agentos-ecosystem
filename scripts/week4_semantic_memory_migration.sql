-- Week 4: Semantic Memory System Migration
-- Advanced Memory System with Vector Storage and Consolidation

-- ===================================
-- SEMANTIC MEMORY TABLES
-- ===================================

-- Semantic memory storage with vector embeddings
CREATE TABLE IF NOT EXISTS semantic_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    embedding vector(1536), -- OpenAI embedding dimension
    concepts TEXT[], -- Array of concept tags
    importance FLOAT DEFAULT 0.5 CHECK (importance >= 0 AND importance <= 1),
    framework VARCHAR(50), -- langchain, swarms, crewai, autogen, universal
    source_type VARCHAR(50) DEFAULT 'user_input', -- user_input, consolidation, system
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Memory consolidation tracking
CREATE TABLE IF NOT EXISTS memory_consolidations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework VARCHAR(50),
    episodic_count INTEGER DEFAULT 0,
    semantic_count INTEGER DEFAULT 0,
    consolidation_score FLOAT DEFAULT 0.0,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    error_message TEXT,
    metadata JSONB DEFAULT '{}'
);

-- Cross-framework memory links
CREATE TABLE IF NOT EXISTS memory_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_memory_id UUID REFERENCES semantic_memories(id) ON DELETE CASCADE,
    target_memory_id UUID REFERENCES semantic_memories(id) ON DELETE CASCADE,
    link_type VARCHAR(50) DEFAULT 'related', -- related, causal, temporal, conceptual
    strength FLOAT DEFAULT 1.0 CHECK (strength >= 0 AND strength <= 1),
    created_at TIMESTAMP DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- ===================================
-- PERFORMANCE INDEXES
-- ===================================

-- Vector similarity search index (IVFFlat for approximate nearest neighbor)
CREATE INDEX IF NOT EXISTS idx_semantic_memories_embedding 
ON semantic_memories USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);

-- Framework and concept queries
CREATE INDEX IF NOT EXISTS idx_semantic_memories_framework ON semantic_memories(framework);
CREATE INDEX IF NOT EXISTS idx_semantic_memories_concepts ON semantic_memories USING GIN(concepts);
CREATE INDEX IF NOT EXISTS idx_semantic_memories_importance ON semantic_memories(importance DESC);
CREATE INDEX IF NOT EXISTS idx_semantic_memories_created_at ON semantic_memories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_semantic_memories_source_type ON semantic_memories(source_type);

-- Consolidation indexes
CREATE INDEX IF NOT EXISTS idx_memory_consolidations_framework ON memory_consolidations(framework);
CREATE INDEX IF NOT EXISTS idx_memory_consolidations_status ON memory_consolidations(status);
CREATE INDEX IF NOT EXISTS idx_memory_consolidations_started_at ON memory_consolidations(started_at DESC);

-- Memory links indexes
CREATE INDEX IF NOT EXISTS idx_memory_links_source ON memory_links(source_memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_links_target ON memory_links(target_memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_links_type ON memory_links(link_type);
CREATE INDEX IF NOT EXISTS idx_memory_links_strength ON memory_links(strength DESC);

-- ===================================
-- TRIGGERS FOR UPDATED_AT
-- ===================================

-- Update semantic_memories updated_at trigger
CREATE TRIGGER update_semantic_memories_updated_at 
BEFORE UPDATE ON semantic_memories 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ===================================
-- SAMPLE SEMANTIC MEMORIES FOR TESTING
-- ===================================

-- Insert sample semantic memories for testing
INSERT INTO semantic_memories (content, concepts, importance, framework, source_type) VALUES
('Machine learning is a subset of artificial intelligence that focuses on algorithms that can learn from data', 
 ARRAY['machine learning', 'artificial intelligence', 'algorithms', 'data'], 
 0.9, 'universal', 'system'),
 
('Natural language processing enables computers to understand and generate human language',
 ARRAY['nlp', 'natural language processing', 'computers', 'human language'],
 0.8, 'langchain', 'system'),
 
('Vector databases store and retrieve high-dimensional vectors efficiently for similarity search',
 ARRAY['vector database', 'embeddings', 'similarity search', 'high-dimensional'],
 0.7, 'universal', 'system'),
 
('Swarm intelligence involves collective behavior of decentralized systems',
 ARRAY['swarm intelligence', 'collective behavior', 'decentralized', 'systems'],
 0.6, 'swarms', 'system'),
 
('CrewAI enables role-based multi-agent workflows for complex task execution',
 ARRAY['crewai', 'multi-agent', 'workflows', 'role-based', 'task execution'],
 0.8, 'crewai', 'system');

-- ===================================
-- MEMORY CONSOLIDATION FUNCTIONS
-- ===================================

-- Function to calculate memory importance based on access patterns
CREATE OR REPLACE FUNCTION calculate_memory_importance(
    memory_id UUID,
    access_count INTEGER DEFAULT 1,
    recency_days INTEGER DEFAULT 1
) RETURNS FLOAT AS $$
DECLARE
    base_importance FLOAT;
    access_boost FLOAT;
    recency_boost FLOAT;
    final_importance FLOAT;
BEGIN
    -- Get current importance
    SELECT importance INTO base_importance 
    FROM semantic_memories 
    WHERE id = memory_id;
    
    -- Calculate access frequency boost (logarithmic)
    access_boost := LEAST(LOG(access_count + 1) * 0.1, 0.3);
    
    -- Calculate recency boost (exponential decay)
    recency_boost := EXP(-recency_days / 30.0) * 0.2;
    
    -- Calculate final importance (capped at 1.0)
    final_importance := LEAST(base_importance + access_boost + recency_boost, 1.0);
    
    -- Update the memory importance
    UPDATE semantic_memories 
    SET importance = final_importance, updated_at = NOW()
    WHERE id = memory_id;
    
    RETURN final_importance;
END;
$$ LANGUAGE plpgsql;

-- Function to find similar memories using vector similarity
CREATE OR REPLACE FUNCTION find_similar_memories(
    query_embedding vector(1536),
    similarity_threshold FLOAT DEFAULT 0.7,
    result_limit INTEGER DEFAULT 10
) RETURNS TABLE(
    memory_id UUID,
    content TEXT,
    similarity_score FLOAT,
    concepts TEXT[],
    framework VARCHAR(50)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        sm.id,
        sm.content,
        1 - (sm.embedding <=> query_embedding) as similarity,
        sm.concepts,
        sm.framework
    FROM semantic_memories sm
    WHERE 1 - (sm.embedding <=> query_embedding) >= similarity_threshold
    ORDER BY sm.embedding <=> query_embedding
    LIMIT result_limit;
END;
$$ LANGUAGE plpgsql;

COMMIT;
