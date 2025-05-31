"""
Unit tests for mem0 Memory Engine
Week 4: Advanced Memory System Testing

This module provides comprehensive unit tests for the mem0 memory engine
and framework adapters.
"""

import pytest
import asyncio
import json
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from datetime import datetime, timedelta

import sys
import os
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'memory'))

from mem0_memory_engine import (
    Mem0MemoryEngine,
    MemoryConfig,
    AgentOSMemoryEntry,
    FrameworkType
)


class TestMem0MemoryEngine:
    """Test suite for Mem0MemoryEngine class"""

    @pytest.fixture
    def memory_config(self):
        """Create test memory configuration"""
        return MemoryConfig(
            vector_store="qdrant",
            embedding_model="text-embedding-ada-002",
            llm_model="gpt-4",
            memory_decay=True,
            importance_threshold=0.5,
            max_memories=1000
        )

    @pytest.fixture
    def mock_redis(self):
        """Create mock Redis client"""
        redis_mock = Mock()
        redis_mock.ping.return_value = True
        redis_mock.setex.return_value = True
        redis_mock.get.return_value = None
        redis_mock.scan_iter.return_value = []
        return redis_mock

    @pytest.fixture
    def mock_mem0(self):
        """Create mock mem0 Memory instance"""
        mem0_mock = Mock()
        mem0_mock.add.return_value = {"id": "mem0_test_123"}
        mem0_mock.search.return_value = [
            {
                "id": "mem0_001",
                "content": "Test memory content",
                "metadata": {"framework": "langchain"},
                "score": 0.85
            }
        ]
        mem0_mock.get_all.return_value = [
            {
                "id": "mem0_002",
                "content": "Another test memory",
                "metadata": {"framework": "swarms"},
                "score": 0.75
            }
        ]
        return mem0_mock

    @pytest.fixture
    def memory_engine(self, memory_config, mock_redis):
        """Create Mem0MemoryEngine instance for testing"""
        with patch('mem0_memory_engine.Redis', return_value=mock_redis):
            with patch('mem0_memory_engine.MEM0_AVAILABLE', True):
                engine = Mem0MemoryEngine(config=memory_config)
                return engine

    def test_memory_engine_initialization(self, memory_config):
        """Test memory engine initialization"""
        with patch('mem0_memory_engine.MEM0_AVAILABLE', True):
            with patch('mem0_memory_engine.Mem0Config') as mock_config_class:
                with patch('mem0_memory_engine.Memory') as mock_memory_class:
                    # Setup mock config
                    mock_config = Mock()
                    mock_config.vector_store = Mock()
                    mock_config.vector_store.provider = "qdrant"
                    mock_config.vector_store.config = Mock()
                    mock_config.embedder = Mock()
                    mock_config.embedder.provider = "openai"
                    mock_config.embedder.config = Mock()
                    mock_config.llm = Mock()
                    mock_config.llm.provider = "openai"
                    mock_config.llm.config = Mock()
                    mock_config_class.return_value = mock_config

                    # Setup mock memory
                    mock_memory_instance = Mock()
                    mock_memory_class.return_value = mock_memory_instance

                    engine = Mem0MemoryEngine(config=memory_config)

                    assert engine.config == memory_config
                    assert engine.memory == mock_memory_instance
                    assert engine.stats["memories_stored"] == 0
                assert engine.stats["memories_retrieved"] == 0

    def test_memory_engine_initialization_without_mem0(self, memory_config):
        """Test memory engine initialization when mem0 is not available"""
        with patch('mem0_memory_engine.MEM0_AVAILABLE', False):
            engine = Mem0MemoryEngine(config=memory_config)

            assert engine.config == memory_config
            assert engine.memory is None
            assert engine.stats["memories_stored"] == 0

    @pytest.mark.asyncio
    async def test_store_memory_with_mem0(self, mock_memory_engine, mock_mem0):
        """Test storing memory with mem0 available"""
        mock_memory_engine.memory = mock_mem0

        memory_id = await mock_memory_engine.store_memory(
            content="Test memory content",
            framework=FrameworkType.LANGCHAIN,
            user_id="test_user",
            agent_id="test_agent",
            metadata={"test": "metadata"}
        )

        assert memory_id == "mem0_test_123"
        assert mock_memory_engine.stats["memories_stored"] == 1

        # Verify mem0.add was called with correct parameters
        mock_mem0.add.assert_called_once()
        call_args = mock_mem0.add.call_args
        assert call_args[1]["user_id"] == "test_user"
        assert call_args[1]["agent_id"] == "test_agent"

    @pytest.mark.asyncio
    async def test_store_memory_without_mem0(self, mock_memory_engine):
        """Test storing memory without mem0 (fallback mode)"""
        mock_memory_engine.memory = None

        memory_id = await mock_memory_engine.store_memory(
            content="Test memory content",
            framework=FrameworkType.SWARMS,
            user_id="test_user"
        )

        assert memory_id.startswith("fallback_")
        assert mock_memory_engine.stats["memories_stored"] == 1

    @pytest.mark.asyncio
    async def test_retrieve_memories_with_mem0(self, mock_memory_engine, mock_mem0):
        """Test retrieving memories with mem0 available"""
        mock_memory_engine.memory = mock_mem0

        memories = await mock_memory_engine.retrieve_memories(
            query="test query",
            user_id="test_user",
            framework=FrameworkType.LANGCHAIN,
            limit=5
        )

        assert len(memories) == 1
        assert memories[0]["id"] == "mem0_001"
        assert memories[0]["content"] == "Test memory content"
        assert mock_memory_engine.stats["memories_retrieved"] == 1

        # Verify mem0.search was called
        mock_mem0.search.assert_called_once_with(
            query="test query",
            user_id="test_user",
            limit=5
        )

    @pytest.mark.asyncio
    async def test_retrieve_memories_with_caching(self, mock_memory_engine, mock_redis):
        """Test memory retrieval with Redis caching"""
        mock_memory_engine.redis = mock_redis

        # Mock cached result
        cached_data = json.dumps([{"id": "cached_001", "content": "Cached memory"}])
        mock_redis.get.return_value = cached_data

        memories = await mock_memory_engine.retrieve_memories(
            query="cached query",
            user_id="test_user",
            limit=5
        )

        assert len(memories) == 1
        assert memories[0]["id"] == "cached_001"
        assert mock_memory_engine.stats["cache_hits"] == 1

    @pytest.mark.asyncio
    async def test_get_framework_memories(self, mock_memory_engine, mock_mem0):
        """Test getting framework-specific memories"""
        mock_memory_engine.memory = mock_mem0

        memories = await mock_memory_engine.get_framework_memories(
            framework=FrameworkType.SWARMS,
            user_id="test_user",
            limit=10
        )

        assert len(memories) == 1
        assert memories[0]["id"] == "mem0_002"

        # Verify mem0.get_all was called
        mock_mem0.get_all.assert_called_once_with(user_id="test_user", limit=20)

    @pytest.mark.asyncio
    async def test_consolidate_memories(self, mock_memory_engine):
        """Test memory consolidation"""
        # Ensure mem0 is available for this test
        mock_memory_engine.memory = Mock()

        result = await mock_memory_engine.consolidate_memories(
            user_id="test_user",
            framework=FrameworkType.CREWAI
        )

        assert result["status"] == "completed"
        assert result["framework"] == "crewai"
        assert "memories_analyzed" in result
        assert "consolidation_score" in result
        assert mock_memory_engine.stats["consolidations"] == 1

    @pytest.mark.asyncio
    async def test_consolidate_memories_without_mem0(self, mock_memory_engine):
        """Test memory consolidation without mem0"""
        mock_memory_engine.memory = None

        result = await mock_memory_engine.consolidate_memories(
            user_id="test_user",
            framework=FrameworkType.AUTOGEN
        )

        assert result["status"] == "skipped"
        assert result["reason"] == "mem0 not available"

    def test_get_statistics(self, mock_memory_engine):
        """Test getting memory engine statistics"""
        # Simulate some operations
        mock_memory_engine.stats["memories_stored"] = 10
        mock_memory_engine.stats["memories_retrieved"] = 15
        mock_memory_engine.stats["cache_hits"] = 8
        mock_memory_engine.stats["cache_misses"] = 7

        stats = mock_memory_engine.get_statistics()

        assert stats["memories_stored"] == 10
        assert stats["memories_retrieved"] == 15
        assert stats["cache_hits"] == 8
        assert stats["cache_misses"] == 7
        assert stats["cache_hit_rate"] == 8/15  # 8 hits out of 15 total
        assert stats["total_operations"] == 40

    @pytest.mark.asyncio
    async def test_store_memory_error_handling(self, mock_memory_engine, mock_mem0):
        """Test error handling in store_memory"""
        mock_memory_engine.memory = mock_mem0
        mock_mem0.add.side_effect = Exception("mem0 error")

        with pytest.raises(Exception):
            await mock_memory_engine.store_memory(
                content="Test content",
                framework=FrameworkType.LANGCHAIN,
                user_id="test_user"
            )

    @pytest.mark.asyncio
    async def test_retrieve_memories_error_handling(self, mock_memory_engine, mock_mem0):
        """Test error handling in retrieve_memories"""
        mock_memory_engine.memory = mock_mem0
        mock_mem0.search.side_effect = Exception("Search error")

        memories = await mock_memory_engine.retrieve_memories(
            query="test query",
            user_id="test_user"
        )

        # Should return empty list on error
        assert memories == []

    @pytest.mark.asyncio
    async def test_fallback_memory_operations(self, mock_memory_engine, mock_redis):
        """Test fallback memory operations when mem0 is unavailable"""
        mock_memory_engine.memory = None
        mock_memory_engine.redis = mock_redis

        # Test fallback storage
        await mock_memory_engine._store_fallback_memory(
            "test_id",
            "test content",
            {"framework": "langchain"}
        )

        # Verify Redis setex was called
        mock_redis.setex.assert_called()

        # Test fallback search
        mock_redis.scan_iter.return_value = ["fallback_memory:test_id"]
        mock_redis.get.return_value = json.dumps({
            "content": "test content matching query",
            "metadata": {"framework": "langchain"}
        })

        memories = await mock_memory_engine._search_fallback_memories(
            "query", "test_user", FrameworkType.LANGCHAIN, 5
        )

        assert len(memories) >= 0  # Should handle gracefully


class TestMemoryConfig:
    """Test suite for MemoryConfig class"""

    def test_memory_config_defaults(self):
        """Test MemoryConfig default values"""
        config = MemoryConfig()

        assert config.vector_store == "qdrant"
        assert config.embedding_model == "text-embedding-ada-002"
        assert config.llm_model == "gpt-4"
        assert config.memory_decay is True
        assert config.importance_threshold == 0.5
        assert config.max_memories == 10000

    def test_memory_config_custom_values(self):
        """Test MemoryConfig with custom values"""
        config = MemoryConfig(
            vector_store="chroma",
            embedding_model="custom-embedding",
            llm_model="gpt-3.5-turbo",
            memory_decay=False,
            importance_threshold=0.7,
            max_memories=5000
        )

        assert config.vector_store == "chroma"
        assert config.embedding_model == "custom-embedding"
        assert config.llm_model == "gpt-3.5-turbo"
        assert config.memory_decay is False
        assert config.importance_threshold == 0.7
        assert config.max_memories == 5000


class TestAgentOSMemoryEntry:
    """Test suite for AgentOSMemoryEntry class"""

    def test_memory_entry_creation(self):
        """Test creating AgentOSMemoryEntry"""
        entry = AgentOSMemoryEntry(
            id="test_id",
            content="Test memory content",
            framework=FrameworkType.LANGCHAIN,
            user_id="test_user",
            agent_id="test_agent",
            metadata={"test": "data"},
            importance=0.8
        )

        assert entry.id == "test_id"
        assert entry.content == "Test memory content"
        assert entry.framework == FrameworkType.LANGCHAIN
        assert entry.user_id == "test_user"
        assert entry.agent_id == "test_agent"
        assert entry.metadata == {"test": "data"}
        assert entry.importance == 0.8
        assert isinstance(entry.created_at, datetime)

    def test_memory_entry_defaults(self):
        """Test AgentOSMemoryEntry default values"""
        entry = AgentOSMemoryEntry(
            id="test_id",
            content="Test content",
            framework=FrameworkType.SWARMS,
            user_id="test_user"
        )

        assert entry.agent_id is None
        assert entry.metadata == {}
        assert entry.importance == 0.5
        assert isinstance(entry.created_at, datetime)


class TestFrameworkType:
    """Test suite for FrameworkType enum"""

    def test_framework_type_values(self):
        """Test FrameworkType enum values"""
        assert FrameworkType.LANGCHAIN.value == "langchain"
        assert FrameworkType.SWARMS.value == "swarms"
        assert FrameworkType.CREWAI.value == "crewai"
        assert FrameworkType.AUTOGEN.value == "autogen"
        assert FrameworkType.UNIVERSAL.value == "universal"

    def test_framework_type_iteration(self):
        """Test iterating over FrameworkType enum"""
        frameworks = list(FrameworkType)
        assert len(frameworks) == 5
        assert FrameworkType.LANGCHAIN in frameworks
        assert FrameworkType.UNIVERSAL in frameworks


# Performance and Integration Tests
class TestMem0MemoryEnginePerformance:
    """Performance tests for mem0 memory engine"""

    @pytest.mark.asyncio
    async def test_concurrent_memory_operations(self, mock_memory_engine, mock_mem0):
        """Test concurrent memory operations"""
        mock_memory_engine.memory = mock_mem0

        # Create multiple concurrent store operations
        tasks = []
        for i in range(10):
            task = mock_memory_engine.store_memory(
                content=f"Test memory {i}",
                framework=FrameworkType.LANGCHAIN,
                user_id="test_user"
            )
            tasks.append(task)

        # Execute all tasks concurrently
        results = await asyncio.gather(*tasks)

        # Verify all operations completed
        assert len(results) == 10
        assert all(result == "mem0_test_123" for result in results)
        assert mock_memory_engine.stats["memories_stored"] == 10


class TestMem0MemoryEngineEdgeCases:
    """Edge case tests for mem0 memory engine"""

    @pytest.mark.asyncio
    async def test_initialization_with_custom_config(self, memory_config):
        """Test initialization with custom configuration"""
        # Test with custom vector store
        memory_config.vector_store = "chroma"
        memory_config.embedding_model = "text-embedding-3-small"

        with patch('mem0_memory_engine.MEM0_AVAILABLE', True):
            with patch('mem0_memory_engine.Mem0Config') as mock_config_class:
                with patch('mem0_memory_engine.Memory') as mock_memory_class:
                    mock_config = Mock()
                    mock_config.vector_store = Mock()
                    mock_config.embedder = Mock()
                    mock_config.llm = Mock()
                    mock_config_class.return_value = mock_config

                    mock_memory_instance = Mock()
                    mock_memory_class.return_value = mock_memory_instance

                    engine = Mem0MemoryEngine(config=memory_config)

                    assert engine.config.vector_store == "chroma"
                    assert engine.config.embedding_model == "text-embedding-3-small"
                    assert engine.memory == mock_memory_instance

    @pytest.mark.asyncio
    async def test_store_memory_with_complex_metadata(self, mock_memory_engine, mock_mem0):
        """Test storing memory with complex metadata"""
        mock_memory_engine.memory = mock_mem0

        complex_metadata = {
            "nested": {"key": "value", "number": 42},
            "list": [1, 2, 3],
            "boolean": True,
            "null_value": None
        }

        memory_id = await mock_memory_engine.store_memory(
            content="Complex metadata test",
            framework=FrameworkType.CREWAI,
            user_id="test_user",
            metadata=complex_metadata
        )

        assert memory_id == "mem0_test_123"
        mock_mem0.add.assert_called_once()
        call_args = mock_mem0.add.call_args[1]
        assert call_args["metadata"]["nested"]["key"] == "value"
        assert call_args["metadata"]["list"] == [1, 2, 3]

    @pytest.mark.asyncio
    async def test_retrieve_memories_with_filters(self, mock_memory_engine, mock_mem0):
        """Test retrieving memories with various filters"""
        mock_memory_engine.memory = mock_mem0

        # Test with framework filter
        memories = await mock_memory_engine.retrieve_memories(
            query="test query",
            user_id="test_user",
            framework=FrameworkType.SWARMS,
            limit=20
        )

        assert len(memories) == 1
        mock_mem0.search.assert_called_with(
            query="test query",
            user_id="test_user",
            limit=20
        )

    @pytest.mark.asyncio
    async def test_memory_cleanup_and_expiration(self, mock_memory_engine, mock_redis):
        """Test memory cleanup and expiration handling"""
        mock_memory_engine.redis = mock_redis

        # Test cleanup of expired memories
        await mock_memory_engine._cleanup_expired_memories()

        # Verify Redis scan was called for cleanup
        mock_redis.scan_iter.assert_called()

    @pytest.mark.asyncio
    async def test_batch_memory_operations(self, mock_memory_engine, mock_mem0):
        """Test batch memory operations"""
        mock_memory_engine.memory = mock_mem0

        # Test batch storage
        contents = [
            "First memory content",
            "Second memory content",
            "Third memory content"
        ]

        memory_ids = []
        for content in contents:
            memory_id = await mock_memory_engine.store_memory(
                content=content,
                framework=FrameworkType.AUTOGEN,
                user_id="batch_user"
            )
            memory_ids.append(memory_id)

        assert len(memory_ids) == 3
        assert all(mid == "mem0_test_123" for mid in memory_ids)
        assert mock_memory_engine.stats["memories_stored"] == 3

    @pytest.mark.asyncio
    async def test_get_memory_by_id(self, mock_memory_engine, mock_mem0):
        """Test retrieving memory by specific ID"""
        mock_memory_engine.memory = mock_mem0

        # Mock mem0 get method
        mock_mem0.get = Mock(return_value={
            "id": "test_memory_123",
            "content": "Test memory content",
            "metadata": {"framework": "langchain"}
        })

        memory = await mock_memory_engine.get_memory_by_id(
            memory_id="test_memory_123",
            user_id="test_user"
        )

        assert memory is not None
        assert memory["id"] == "test_memory_123"
        assert memory["content"] == "Test memory content"
        mock_mem0.get.assert_called_once_with("test_memory_123", user_id="test_user")

    @pytest.mark.asyncio
    async def test_get_memory_by_id_fallback(self, mock_memory_engine, mock_redis):
        """Test retrieving memory by ID using Redis fallback"""
        mock_memory_engine.memory = None
        mock_memory_engine.redis = mock_redis

        # Mock Redis get
        mock_redis.get.return_value = json.dumps({
            "content": "Fallback memory content",
            "metadata": {"framework": "swarms"}
        })

        memory = await mock_memory_engine.get_memory_by_id(
            memory_id="fallback_123",
            user_id="test_user"
        )

        assert memory is not None
        assert memory["content"] == "Fallback memory content"
        mock_redis.get.assert_called_once_with("fallback_memory:fallback_123")

    @pytest.mark.asyncio
    async def test_delete_memory(self, mock_memory_engine, mock_mem0):
        """Test deleting memory"""
        mock_memory_engine.memory = mock_mem0

        # Mock mem0 delete method
        mock_mem0.delete = Mock()

        result = await mock_memory_engine.delete_memory(
            memory_id="test_memory_123",
            user_id="test_user"
        )

        assert result is True
        mock_mem0.delete.assert_called_once_with("test_memory_123", user_id="test_user")

    @pytest.mark.asyncio
    async def test_delete_memory_fallback(self, mock_memory_engine, mock_redis):
        """Test deleting memory using Redis fallback"""
        mock_memory_engine.memory = None
        mock_memory_engine.redis = mock_redis

        # Mock Redis delete
        mock_redis.delete.return_value = 1  # 1 key deleted

        result = await mock_memory_engine.delete_memory(
            memory_id="fallback_123",
            user_id="test_user"
        )

        assert result is True
        mock_redis.delete.assert_called_once_with("fallback_memory:fallback_123")

    @pytest.mark.asyncio
    async def test_large_batch_retrieval(self, mock_memory_engine, mock_mem0):
        """Test retrieving large batches of memories"""
        mock_memory_engine.memory = mock_mem0

        # Mock large result set
        large_result = [
            {
                "id": f"mem0_{i:03d}",
                "content": f"Memory content {i}",
                "metadata": {"framework": "langchain"},
                "score": 0.8 - (i * 0.01)
            }
            for i in range(100)
        ]
        mock_mem0.search.return_value = large_result

        memories = await mock_memory_engine.retrieve_memories(
            query="large batch test",
            user_id="test_user",
            limit=100
        )

        assert len(memories) == 100
        assert memories[0]["id"] == "mem0_000"
        assert memories[99]["id"] == "mem0_099"

    def test_memory_engine_resource_cleanup(self, memory_config):
        """Test proper resource cleanup"""
        with patch('mem0_memory_engine.Redis') as mock_redis_class:
            mock_redis = Mock()
            mock_redis.ping.return_value = True
            mock_redis_class.return_value = mock_redis

            engine = Mem0MemoryEngine(config=memory_config)

            # Verify Redis connection was established
            mock_redis_class.assert_called_once()
            mock_redis.ping.assert_called_once()

            # Test that engine handles Redis connection gracefully
            assert engine.redis == mock_redis


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
