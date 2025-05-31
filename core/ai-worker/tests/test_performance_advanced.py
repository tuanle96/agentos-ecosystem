"""
Advanced performance and concurrency tests
"""
import pytest
import asyncio
import time
from concurrent.futures import ThreadPoolExecutor
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
import main


class TestAdvancedPerformance:
    """Test advanced performance scenarios"""

    def setup_method(self):
        """Setup test client and clear registry"""
        self.client = TestClient(main.app)
        main.agent_registry.clear()

    def test_health_check_response_time(self):
        """Test health check response time is acceptable"""
        start_time = time.time()
        response = self.client.get("/health")
        end_time = time.time()
        
        response_time = end_time - start_time
        
        assert response.status_code == 200
        assert response_time < 0.1  # Should respond within 100ms

    def test_framework_status_response_time(self):
        """Test framework status response time"""
        start_time = time.time()
        response = self.client.get("/framework/status")
        end_time = time.time()
        
        response_time = end_time - start_time
        
        assert response.status_code == 200
        assert response_time < 0.1  # Should respond within 100ms

    def test_tools_endpoint_response_time(self):
        """Test tools endpoint response time"""
        start_time = time.time()
        response = self.client.get("/tools")
        end_time = time.time()
        
        response_time = end_time - start_time
        
        assert response.status_code == 200
        assert response_time < 0.1  # Should respond within 100ms

    def test_list_agents_response_time(self):
        """Test list agents response time with many agents"""
        # Add multiple agents to test performance
        for i in range(50):
            main.agent_registry[f"agent-{i}"] = MagicMock()
        
        start_time = time.time()
        response = self.client.get("/agents")
        end_time = time.time()
        
        response_time = end_time - start_time
        
        assert response.status_code == 200
        assert response_time < 0.2  # Should handle 50 agents within 200ms
        
        data = response.json()
        assert data["count"] == 50

    def test_concurrent_health_checks_advanced(self):
        """Test many concurrent health check requests"""
        def make_health_request():
            response = self.client.get("/health")
            return response.status_code == 200
        
        # Test with 100 concurrent requests
        with ThreadPoolExecutor(max_workers=20) as executor:
            futures = [executor.submit(make_health_request) for _ in range(100)]
            results = [future.result() for future in futures]
        
        # All requests should succeed
        assert all(results)
        assert len(results) == 100

    def test_concurrent_framework_status_requests(self):
        """Test concurrent framework status requests"""
        def make_status_request():
            response = self.client.get("/framework/status")
            return response.status_code == 200 and "framework" in response.json()
        
        # Test with 50 concurrent requests
        with ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(make_status_request) for _ in range(50)]
            results = [future.result() for future in futures]
        
        # All requests should succeed
        assert all(results)
        assert len(results) == 50

    def test_memory_usage_with_many_agents(self):
        """Test memory usage doesn't grow excessively with many agents"""
        import psutil
        import os
        
        # Get initial memory usage
        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB
        
        # Create many mock agents
        for i in range(100):
            mock_wrapper = MagicMock()
            mock_wrapper.agent_config = MagicMock()
            mock_wrapper.agent_config.name = f"Agent {i}"
            mock_wrapper.agent_config.description = f"Test agent {i}"
            mock_wrapper.agent_config.capabilities = ["calculations"]
            mock_wrapper.tools = [MagicMock()]
            mock_wrapper.agent = MagicMock()
            main.agent_registry[f"agent-{i}"] = mock_wrapper
        
        # Get memory usage after creating agents
        final_memory = process.memory_info().rss / 1024 / 1024  # MB
        memory_increase = final_memory - initial_memory
        
        # Memory increase should be reasonable (less than 50MB for 100 agents)
        assert memory_increase < 50
        
        # Test that we can still make requests efficiently
        response = self.client.get("/agents")
        assert response.status_code == 200
        assert response.json()["count"] == 100


class TestConcurrentAgentOperations:
    """Test concurrent agent operations"""

    def setup_method(self):
        """Setup test client and clear registry"""
        self.client = TestClient(main.app)
        main.agent_registry.clear()

    def test_concurrent_agent_creation(self):
        """Test concurrent agent creation"""
        agent_configs = [
            {
                "name": f"Concurrent Agent {i}",
                "description": f"Agent created concurrently {i}",
                "capabilities": ["calculations"],
                "framework_preference": "langchain"
            }
            for i in range(10)
        ]
        
        def create_agent(config):
            with patch('main.LANGCHAIN_AVAILABLE', True):
                with patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'}):
                    with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
                        mock_wrapper = MagicMock()
                        mock_wrapper.agent_id = f"agent-{config['name']}"
                        mock_wrapper.tools = [MagicMock()]
                        mock_wrapper.initialize = AsyncMock()
                        mock_wrapper_class.return_value = mock_wrapper
                        
                        response = self.client.post("/agents/create", json=config)
                        return response.status_code == 200
        
        # Create agents concurrently
        with ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(create_agent, config) for config in agent_configs]
            results = [future.result() for future in futures]
        
        # All creations should succeed
        assert all(results)

    def test_concurrent_agent_retrieval(self):
        """Test concurrent agent retrieval"""
        # Create some test agents
        for i in range(10):
            mock_config = MagicMock()
            mock_config.name = f"Test Agent {i}"
            mock_config.description = f"Description {i}"
            mock_config.capabilities = ["calculations"]
            
            mock_wrapper = MagicMock()
            mock_wrapper.agent_config = mock_config
            mock_wrapper.tools = [MagicMock()]
            mock_wrapper.agent = MagicMock()
            
            main.agent_registry[f"agent-{i}"] = mock_wrapper
        
        def get_agent(agent_id):
            response = self.client.get(f"/agents/{agent_id}")
            return response.status_code == 200
        
        # Retrieve agents concurrently
        agent_ids = [f"agent-{i}" for i in range(10)]
        with ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(get_agent, agent_id) for agent_id in agent_ids]
            results = [future.result() for future in futures]
        
        # All retrievals should succeed
        assert all(results)

    def test_concurrent_agent_deletion(self):
        """Test concurrent agent deletion"""
        # Create test agents
        for i in range(10):
            main.agent_registry[f"delete-agent-{i}"] = MagicMock()
        
        def delete_agent(agent_id):
            response = self.client.delete(f"/agents/{agent_id}")
            return response.status_code == 200
        
        # Delete agents concurrently
        agent_ids = [f"delete-agent-{i}" for i in range(10)]
        with ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(delete_agent, agent_id) for agent_id in agent_ids]
            results = [future.result() for future in futures]
        
        # All deletions should succeed
        assert all(results)
        
        # Registry should be empty
        assert len(main.agent_registry) == 0

    @pytest.mark.asyncio
    async def test_concurrent_task_execution(self):
        """Test concurrent task execution"""
        # Create a mock agent
        mock_wrapper = MagicMock()
        mock_wrapper.execute = AsyncMock(return_value={
            "task_id": "task-123",
            "result": "Task completed",
            "status": "completed",
            "execution_time": 0.01,
            "agent_id": "test-agent",
            "framework": "langchain"
        })
        
        main.agent_registry["test-agent"] = mock_wrapper
        
        async def execute_task(task_num):
            task_request = {
                "agent_id": "test-agent",
                "task": f"Task {task_num}",
                "context": {"task_number": task_num}
            }
            
            response = self.client.post("/agents/test-agent/execute", json=task_request)
            return response.status_code == 200
        
        # Execute tasks concurrently
        tasks = [execute_task(i) for i in range(20)]
        results = await asyncio.gather(*tasks)
        
        # All executions should succeed
        assert all(results)
        assert len(results) == 20


class TestStressTests:
    """Stress tests for the AI worker"""

    def setup_method(self):
        """Setup test client"""
        self.client = TestClient(main.app)
        main.agent_registry.clear()

    def test_rapid_health_checks(self):
        """Test rapid successive health checks"""
        start_time = time.time()
        
        # Make 1000 rapid health checks
        for _ in range(1000):
            response = self.client.get("/health")
            assert response.status_code == 200
        
        end_time = time.time()
        total_time = end_time - start_time
        
        # Should complete 1000 requests in reasonable time (less than 10 seconds)
        assert total_time < 10.0
        
        # Average response time should be reasonable
        avg_response_time = total_time / 1000
        assert avg_response_time < 0.01  # Less than 10ms average

    def test_large_agent_registry_operations(self):
        """Test operations with large agent registry"""
        # Create 500 mock agents
        for i in range(500):
            mock_wrapper = MagicMock()
            mock_wrapper.agent_config = MagicMock()
            mock_wrapper.agent_config.name = f"Agent {i}"
            mock_wrapper.agent_config.description = f"Description {i}"
            mock_wrapper.agent_config.capabilities = ["calculations"]
            mock_wrapper.tools = [MagicMock()]
            mock_wrapper.agent = MagicMock()
            main.agent_registry[f"agent-{i}"] = mock_wrapper
        
        # Test listing agents with large registry
        start_time = time.time()
        response = self.client.get("/agents")
        end_time = time.time()
        
        assert response.status_code == 200
        assert response.json()["count"] == 500
        assert (end_time - start_time) < 1.0  # Should complete within 1 second
        
        # Test getting specific agents
        for i in range(0, 500, 50):  # Test every 50th agent
            response = self.client.get(f"/agents/agent-{i}")
            assert response.status_code == 200

    def test_error_handling_under_load(self):
        """Test error handling under high load"""
        def make_failing_request():
            # Try to get non-existent agent
            response = self.client.get("/agents/nonexistent-agent")
            return response.status_code == 404
        
        # Make many concurrent failing requests
        with ThreadPoolExecutor(max_workers=20) as executor:
            futures = [executor.submit(make_failing_request) for _ in range(200)]
            results = [future.result() for future in futures]
        
        # All should properly return 404
        assert all(results)
        assert len(results) == 200

    def test_mixed_operations_under_load(self):
        """Test mixed operations under load"""
        # Add some agents
        for i in range(20):
            main.agent_registry[f"load-agent-{i}"] = MagicMock()
        
        def mixed_operations():
            operations = [
                lambda: self.client.get("/health"),
                lambda: self.client.get("/agents"),
                lambda: self.client.get("/tools"),
                lambda: self.client.get("/framework/status"),
                lambda: self.client.get(f"/agents/load-agent-{time.time() % 20:.0f}"),
            ]
            
            # Perform random operations
            import random
            operation = random.choice(operations)
            response = operation()
            return response.status_code in [200, 404]  # 404 is acceptable for some operations
        
        # Perform mixed operations concurrently
        with ThreadPoolExecutor(max_workers=15) as executor:
            futures = [executor.submit(mixed_operations) for _ in range(300)]
            results = [future.result() for future in futures]
        
        # Most operations should succeed
        success_rate = sum(results) / len(results)
        assert success_rate > 0.95  # At least 95% success rate
