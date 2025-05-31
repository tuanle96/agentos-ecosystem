"""
Pytest configuration and fixtures for Week 4 Memory System Tests
"""

import pytest
import asyncio
import os
import sys
from unittest.mock import Mock, AsyncMock, patch

# Add the parent directory to the path so we can import our modules
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from memory.mem0_memory_engine import (
    Mem0MemoryEngine, 
    MemoryConfig, 
    FrameworkType
)


@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
def memory_config():
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
def mock_redis():
    """Create mock Redis client"""
    redis_mock = Mock()
    redis_mock.ping.return_value = True
    redis_mock.setex.return_value = True
    redis_mock.get.return_value = None
    redis_mock.scan_iter.return_value = []
    redis_mock.delete.return_value = 1
    redis_mock.exists.return_value = False
    return redis_mock


@pytest.fixture
def mock_mem0():
    """Create mock mem0 Memory instance"""
    mem0_mock = Mock()
    mem0_mock.add.return_value = {"id": "mem0_test_123"}
    mem0_mock.search.return_value = [
        {
            "id": "mem0_001",
            "content": "Test memory content",
            "metadata": {"framework": "langchain", "timestamp": "2024-12-27T10:00:00"},
            "score": 0.85
        },
        {
            "id": "mem0_002", 
            "content": "Another test memory",
            "metadata": {"framework": "swarms", "timestamp": "2024-12-27T09:00:00"},
            "score": 0.75
        }
    ]
    mem0_mock.get_all.return_value = [
        {
            "id": "mem0_003",
            "content": "Framework specific memory",
            "metadata": {"framework": "crewai", "timestamp": "2024-12-27T08:00:00"}
        }
    ]
    mem0_mock.delete.return_value = True
    mem0_mock.update.return_value = {"id": "mem0_test_123", "status": "updated"}
    return mem0_mock


@pytest.fixture
def mock_memory_engine(memory_config, mock_redis, mock_mem0):
    """Create mock Mem0MemoryEngine instance"""
    with patch('memory.mem0_memory_engine.Redis', return_value=mock_redis):
        with patch('memory.mem0_memory_engine.MEM0_AVAILABLE', True):
            with patch('memory.mem0_memory_engine.Memory', return_value=mock_mem0):
                engine = Mem0MemoryEngine(config=memory_config)
                engine.memory = mock_mem0
                engine.redis = mock_redis
                return engine


@pytest.fixture
def sample_messages():
    """Sample messages for testing different frameworks"""
    return {
        "langchain": [
            {"role": "user", "content": "What is machine learning?"},
            {"role": "assistant", "content": "Machine learning is a subset of AI that enables computers to learn and improve from experience without being explicitly programmed."},
            {"role": "user", "content": "Can you give me some examples?"},
            {"role": "assistant", "content": "Sure! Examples include image recognition, natural language processing, recommendation systems, and autonomous vehicles."}
        ],
        "swarms": [
            {"agent_id": "agent_1", "content": "Initiating swarm coordination protocol"},
            {"agent_id": "agent_2", "content": "Acknowledging coordination request"},
            {"agent_id": "agent_3", "content": "Joining swarm network"},
            {"agent_id": "agent_1", "content": "Task distribution complete"}
        ],
        "crewai": [
            {"role": "researcher", "content": "Conducting market research on AI trends"},
            {"role": "analyst", "content": "Analyzing collected data for patterns"},
            {"role": "writer", "content": "Drafting comprehensive report"},
            {"role": "reviewer", "content": "Reviewing and finalizing content"}
        ],
        "autogen": [
            {"role": "user", "content": "Write a Python function to calculate fibonacci numbers"},
            {"role": "assistant", "content": "Here's a Python function for fibonacci calculation:\n\ndef fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)"},
            {"role": "code_executor", "content": "Executing fibonacci function..."},
            {"role": "assistant", "content": "The function executed successfully. For n=10, result is 55."}
        ]
    }


@pytest.fixture
def sample_memory_entries():
    """Sample memory entries for testing"""
    return [
        {
            "content": "LangChain provides a framework for developing applications with LLMs",
            "framework": FrameworkType.LANGCHAIN,
            "concepts": ["langchain", "LLM", "framework", "applications"],
            "importance": 0.9,
            "metadata": {"source": "documentation", "category": "framework"}
        },
        {
            "content": "Swarms enable distributed AI agent coordination and collaboration",
            "framework": FrameworkType.SWARMS,
            "concepts": ["swarms", "distributed", "coordination", "collaboration"],
            "importance": 0.8,
            "metadata": {"source": "research", "category": "coordination"}
        },
        {
            "content": "CrewAI facilitates role-based multi-agent workflows and task management",
            "framework": FrameworkType.CREWAI,
            "concepts": ["crewai", "roles", "workflows", "task management"],
            "importance": 0.85,
            "metadata": {"source": "tutorial", "category": "workflow"}
        },
        {
            "content": "AutoGen enables conversational AI with code generation capabilities",
            "framework": FrameworkType.AUTOGEN,
            "concepts": ["autogen", "conversational", "code generation"],
            "importance": 0.75,
            "metadata": {"source": "example", "category": "conversation"}
        }
    ]


@pytest.fixture
def test_user_id():
    """Test user ID"""
    return "test_user_12345"


@pytest.fixture
def test_agent_id():
    """Test agent ID"""
    return "test_agent_67890"


@pytest.fixture
def framework_types():
    """List of all framework types for testing"""
    return [
        FrameworkType.LANGCHAIN,
        FrameworkType.SWARMS,
        FrameworkType.CREWAI,
        FrameworkType.AUTOGEN,
        FrameworkType.UNIVERSAL
    ]


@pytest.fixture
def mock_environment_variables():
    """Mock environment variables for testing"""
    env_vars = {
        "REDIS_HOST": "localhost",
        "REDIS_PORT": "6379",
        "REDIS_DB": "0",
        "MEM0_API_KEY": "test_api_key",
        "OPENAI_API_KEY": "test_openai_key",
        "QDRANT_URL": "http://localhost:6333",
        "QDRANT_API_KEY": "test_qdrant_key"
    }
    
    with patch.dict(os.environ, env_vars):
        yield env_vars


@pytest.fixture
def performance_test_data():
    """Generate test data for performance testing"""
    return {
        "small_dataset": [
            {
                "content": f"Performance test memory {i}: Testing memory operations at scale",
                "framework": FrameworkType.LANGCHAIN,
                "concepts": ["performance", "testing", "memory", f"test_{i}"],
                "importance": 0.5 + (i % 5) * 0.1
            }
            for i in range(10)
        ],
        "medium_dataset": [
            {
                "content": f"Medium scale test memory {i}: Evaluating system performance under moderate load",
                "framework": FrameworkType.SWARMS if i % 2 == 0 else FrameworkType.CREWAI,
                "concepts": ["medium", "scale", "performance", f"test_{i}"],
                "importance": 0.4 + (i % 6) * 0.1
            }
            for i in range(50)
        ],
        "large_dataset": [
            {
                "content": f"Large scale test memory {i}: Stress testing memory system with high volume operations",
                "framework": FrameworkType(list(FrameworkType)[i % 4]),
                "concepts": ["large", "scale", "stress", "testing", f"test_{i}"],
                "importance": 0.3 + (i % 7) * 0.1
            }
            for i in range(100)
        ]
    }


@pytest.fixture
def error_scenarios():
    """Error scenarios for testing error handling"""
    return {
        "invalid_content": [
            {"content": "", "error": "Empty content"},
            {"content": None, "error": "None content"},
            {"content": " " * 10000, "error": "Content too long"}
        ],
        "invalid_framework": [
            {"framework": "invalid_framework", "error": "Invalid framework"},
            {"framework": None, "error": "None framework"},
            {"framework": "", "error": "Empty framework"}
        ],
        "invalid_metadata": [
            {"metadata": "not_a_dict", "error": "Invalid metadata type"},
            {"metadata": {"key": None}, "error": "None values in metadata"}
        ],
        "invalid_importance": [
            {"importance": -0.5, "error": "Negative importance"},
            {"importance": 1.5, "error": "Importance > 1"},
            {"importance": "high", "error": "Non-numeric importance"}
        ]
    }


# Async test helpers
@pytest.fixture
def async_test_helper():
    """Helper functions for async testing"""
    class AsyncTestHelper:
        @staticmethod
        async def run_concurrent_operations(operations, max_concurrent=10):
            """Run operations concurrently with limited concurrency"""
            semaphore = asyncio.Semaphore(max_concurrent)
            
            async def run_with_semaphore(operation):
                async with semaphore:
                    return await operation()
            
            tasks = [run_with_semaphore(op) for op in operations]
            return await asyncio.gather(*tasks, return_exceptions=True)
        
        @staticmethod
        async def measure_operation_time(operation):
            """Measure the time taken by an async operation"""
            import time
            start_time = time.time()
            result = await operation()
            end_time = time.time()
            return result, end_time - start_time
        
        @staticmethod
        def create_mock_async_operation(return_value, delay=0):
            """Create a mock async operation"""
            async def mock_operation():
                if delay > 0:
                    await asyncio.sleep(delay)
                return return_value
            return mock_operation
    
    return AsyncTestHelper()


# Test data validation helpers
@pytest.fixture
def validation_helpers():
    """Helper functions for test data validation"""
    class ValidationHelpers:
        @staticmethod
        def validate_memory_id(memory_id, prefix="mem0_"):
            """Validate memory ID format"""
            return (
                isinstance(memory_id, str) and
                memory_id.startswith(prefix) and
                len(memory_id) > len(prefix)
            )
        
        @staticmethod
        def validate_memory_entry(entry, required_fields=None):
            """Validate memory entry structure"""
            if required_fields is None:
                required_fields = ["content", "framework", "metadata"]
            
            return all(field in entry for field in required_fields)
        
        @staticmethod
        def validate_search_results(results, query, min_score=0.0):
            """Validate search results"""
            if not isinstance(results, list):
                return False
            
            for result in results:
                if not isinstance(result, dict):
                    return False
                if "score" in result and result["score"] < min_score:
                    return False
                if "content" not in result:
                    return False
            
            return True
        
        @staticmethod
        def validate_framework_type(framework):
            """Validate framework type"""
            return isinstance(framework, FrameworkType)
    
    return ValidationHelpers()


# Cleanup helpers
@pytest.fixture(autouse=True)
def cleanup_test_environment():
    """Automatically cleanup test environment after each test"""
    yield
    # Cleanup code here if needed
    # For now, we rely on mocks so no actual cleanup needed


# Performance monitoring
@pytest.fixture
def performance_monitor():
    """Monitor performance during tests"""
    class PerformanceMonitor:
        def __init__(self):
            self.metrics = {}
        
        def start_timer(self, operation_name):
            import time
            self.metrics[operation_name] = {"start": time.time()}
        
        def end_timer(self, operation_name):
            import time
            if operation_name in self.metrics:
                self.metrics[operation_name]["end"] = time.time()
                self.metrics[operation_name]["duration"] = (
                    self.metrics[operation_name]["end"] - 
                    self.metrics[operation_name]["start"]
                )
        
        def get_duration(self, operation_name):
            return self.metrics.get(operation_name, {}).get("duration", 0)
        
        def get_all_metrics(self):
            return self.metrics
    
    return PerformanceMonitor()


# Test markers for different test categories
def pytest_configure(config):
    """Configure pytest markers"""
    config.addinivalue_line(
        "markers", "unit: mark test as a unit test"
    )
    config.addinivalue_line(
        "markers", "integration: mark test as an integration test"
    )
    config.addinivalue_line(
        "markers", "performance: mark test as a performance test"
    )
    config.addinivalue_line(
        "markers", "slow: mark test as slow running"
    )
    config.addinivalue_line(
        "markers", "async_test: mark test as async test"
    )
