"""
Advanced integration tests for comprehensive coverage
"""
import pytest
import os
import time
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
from fastapi import HTTPException
import main


class TestAdvancedIntegration:
    """Advanced integration tests for full workflow coverage"""

    def setup_method(self):
        """Setup test client and clear registry"""
        self.client = TestClient(main.app)
        main.agent_registry.clear()

    def test_complete_agent_lifecycle(self):
        """Test complete agent lifecycle: create -> execute -> get -> delete"""
        agent_config = {
            "name": "Lifecycle Test Agent",
            "description": "Agent for testing complete lifecycle",
            "capabilities": ["calculations", "text_processing"],
            "framework_preference": "langchain"
        }
        
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
                with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
                    # Setup mock wrapper
                    mock_wrapper = MagicMock()
                    mock_wrapper.agent_id = "lifecycle-agent-123"
                    mock_wrapper.tools = [MagicMock(), MagicMock()]
                    mock_wrapper.initialize = AsyncMock()
                    mock_wrapper.execute = AsyncMock(return_value={
                        "task_id": "task-123",
                        "result": "Lifecycle test completed",
                        "status": "completed",
                        "execution_time": 0.05,
                        "agent_id": "lifecycle-agent-123",
                        "framework": "langchain"
                    })
                    mock_wrapper.agent_config = MagicMock()
                    mock_wrapper.agent_config.name = "Lifecycle Test Agent"
                    mock_wrapper.agent_config.description = "Agent for testing complete lifecycle"
                    mock_wrapper.agent_config.capabilities = ["calculations", "text_processing"]
                    mock_wrapper.agent = MagicMock()  # Agent is initialized
                    mock_wrapper_class.return_value = mock_wrapper
                    
                    # 1. Create agent
                    create_response = self.client.post("/agents/create", json=agent_config)
                    assert create_response.status_code == 200
                    agent_data = create_response.json()
                    agent_id = agent_data["agent_id"]
                    
                    # 2. Execute task
                    task_request = {
                        "agent_id": agent_id,
                        "task": "Test lifecycle execution",
                        "context": {"test": "lifecycle"}
                    }
                    execute_response = self.client.post(f"/agents/{agent_id}/execute", json=task_request)
                    assert execute_response.status_code == 200
                    task_data = execute_response.json()
                    assert task_data["status"] == "completed"
                    
                    # 3. Get agent details
                    get_response = self.client.get(f"/agents/{agent_id}")
                    assert get_response.status_code == 200
                    get_data = get_response.json()
                    assert get_data["name"] == "Lifecycle Test Agent"
                    
                    # 4. Delete agent
                    delete_response = self.client.delete(f"/agents/{agent_id}")
                    assert delete_response.status_code == 200
                    
                    # 5. Verify agent is deleted
                    get_after_delete = self.client.get(f"/agents/{agent_id}")
                    assert get_after_delete.status_code == 404

    def test_multiple_agents_management(self):
        """Test managing multiple agents simultaneously"""
        agent_configs = [
            {
                "name": f"Multi Agent {i}",
                "description": f"Agent {i} for multi-agent testing",
                "capabilities": ["calculations"],
                "framework_preference": "langchain"
            }
            for i in range(5)
        ]
        
        created_agents = []
        
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
                with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
                    def create_mock_wrapper(config):
                        mock_wrapper = MagicMock()
                        mock_wrapper.agent_id = f"multi-agent-{len(created_agents)}"
                        mock_wrapper.tools = [MagicMock()]
                        mock_wrapper.initialize = AsyncMock()
                        mock_wrapper.agent_config = MagicMock()
                        mock_wrapper.agent_config.name = config["name"]
                        mock_wrapper.agent_config.description = config["description"]
                        mock_wrapper.agent_config.capabilities = config["capabilities"]
                        mock_wrapper.agent = MagicMock()
                        return mock_wrapper
                    
                    mock_wrapper_class.side_effect = lambda config: create_mock_wrapper(config)
                    
                    # Create multiple agents
                    for config in agent_configs:
                        response = self.client.post("/agents/create", json=config)
                        assert response.status_code == 200
                        created_agents.append(response.json()["agent_id"])
                    
                    # List all agents
                    list_response = self.client.get("/agents")
                    assert list_response.status_code == 200
                    list_data = list_response.json()
                    assert list_data["count"] == 5
                    assert len(list_data["agents"]) == 5
                    
                    # Get each agent individually
                    for agent_id in created_agents:
                        get_response = self.client.get(f"/agents/{agent_id}")
                        assert get_response.status_code == 200
                    
                    # Delete all agents
                    for agent_id in created_agents:
                        delete_response = self.client.delete(f"/agents/{agent_id}")
                        assert delete_response.status_code == 200
                    
                    # Verify all deleted
                    final_list = self.client.get("/agents")
                    assert final_list.json()["count"] == 0

    def test_error_recovery_scenarios(self):
        """Test error recovery in various scenarios"""
        # Test creating agent with invalid configuration
        invalid_configs = [
            {},  # Empty config
            {"name": ""},  # Empty name
            {"name": "Test", "description": ""},  # Empty description
            {"name": "Test", "description": "Test", "capabilities": []},  # No capabilities
        ]
        
        for invalid_config in invalid_configs:
            response = self.client.post("/agents/create", json=invalid_config)
            # Should return validation error (422) or server error (500)
            assert response.status_code in [422, 500]

    def test_concurrent_operations_on_same_agent(self):
        """Test concurrent operations on the same agent"""
        agent_config = {
            "name": "Concurrent Test Agent",
            "description": "Agent for concurrent testing",
            "capabilities": ["calculations"],
            "framework_preference": "langchain"
        }
        
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
                with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
                    mock_wrapper = MagicMock()
                    mock_wrapper.agent_id = "concurrent-agent"
                    mock_wrapper.tools = [MagicMock()]
                    mock_wrapper.initialize = AsyncMock()
                    mock_wrapper.execute = AsyncMock(return_value={
                        "task_id": "concurrent-task",
                        "result": "Concurrent execution",
                        "status": "completed",
                        "execution_time": 0.01,
                        "agent_id": "concurrent-agent",
                        "framework": "langchain"
                    })
                    mock_wrapper.agent_config = MagicMock()
                    mock_wrapper.agent_config.name = "Concurrent Test Agent"
                    mock_wrapper.agent_config.description = "Agent for concurrent testing"
                    mock_wrapper.agent_config.capabilities = ["calculations"]
                    mock_wrapper.agent = MagicMock()
                    mock_wrapper_class.return_value = mock_wrapper
                    
                    # Create agent
                    create_response = self.client.post("/agents/create", json=agent_config)
                    assert create_response.status_code == 200
                    agent_id = create_response.json()["agent_id"]
                    
                    # Test concurrent get operations
                    from concurrent.futures import ThreadPoolExecutor
                    
                    def get_agent():
                        response = self.client.get(f"/agents/{agent_id}")
                        return response.status_code == 200
                    
                    with ThreadPoolExecutor(max_workers=10) as executor:
                        futures = [executor.submit(get_agent) for _ in range(20)]
                        results = [future.result() for future in futures]
                    
                    # All concurrent gets should succeed
                    assert all(results)

    def test_framework_status_comprehensive(self):
        """Test comprehensive framework status reporting"""
        # Test with no agents
        response = self.client.get("/framework/status")
        assert response.status_code == 200
        data = response.json()
        assert data["framework"] == "langchain"
        assert data["active_agents"] == 0
        assert "supported_capabilities" in data
        assert len(data["supported_capabilities"]) == 5
        
        # Test with OpenAI key configured
        with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
            response = self.client.get("/framework/status")
            assert response.status_code == 200
            data = response.json()
            assert data["openai_configured"] is True
        
        # Test without OpenAI key
        with patch.dict(os.environ, {}, clear=True):
            response = self.client.get("/framework/status")
            assert response.status_code == 200
            data = response.json()
            assert data["openai_configured"] is False

    def test_tools_endpoint_comprehensive(self):
        """Test comprehensive tools endpoint"""
        response = self.client.get("/tools")
        assert response.status_code == 200
        data = response.json()
        
        assert data["count"] == 5
        assert data["framework"] == "langchain"
        assert len(data["tools"]) == 5
        
        # Verify all expected tools are present
        tool_names = [tool["name"] for tool in data["tools"]]
        expected_tools = ["web_search", "calculator", "text_processor", "file_operations", "api_calls"]
        for expected_tool in expected_tools:
            assert expected_tool in tool_names
        
        # Verify tool structure
        for tool in data["tools"]:
            assert "name" in tool
            assert "description" in tool
            assert "category" in tool
            assert tool["name"] in expected_tools

    def test_health_check_comprehensive(self):
        """Test comprehensive health check"""
        # Test with LangChain available
        with patch('main.LANGCHAIN_AVAILABLE', True):
            response = self.client.get("/health")
            assert response.status_code == 200
            data = response.json()
            assert data["status"] == "healthy"
            assert data["service"] == "agentos-ai-worker"
            assert data["framework"] == "langchain"
            assert data["langchain_available"] is True
        
        # Test with LangChain not available
        with patch('main.LANGCHAIN_AVAILABLE', False):
            response = self.client.get("/health")
            assert response.status_code == 200
            data = response.json()
            assert data["langchain_available"] is False

    def test_performance_under_load(self):
        """Test performance characteristics under load"""
        # Create multiple agents and perform operations
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
                with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
                    mock_wrapper = MagicMock()
                    mock_wrapper.agent_id = "perf-agent"
                    mock_wrapper.tools = [MagicMock()]
                    mock_wrapper.initialize = AsyncMock()
                    mock_wrapper.agent_config = MagicMock()
                    mock_wrapper.agent_config.name = "Performance Agent"
                    mock_wrapper.agent_config.description = "Agent for performance testing"
                    mock_wrapper.agent_config.capabilities = ["calculations"]
                    mock_wrapper.agent = MagicMock()
                    mock_wrapper_class.return_value = mock_wrapper
                    
                    # Measure response times for various operations
                    operations = [
                        lambda: self.client.get("/health"),
                        lambda: self.client.get("/framework/status"),
                        lambda: self.client.get("/tools"),
                        lambda: self.client.get("/agents"),
                    ]
                    
                    for operation in operations:
                        start_time = time.time()
                        response = operation()
                        end_time = time.time()
                        
                        assert response.status_code == 200
                        response_time = end_time - start_time
                        assert response_time < 0.1  # Should respond within 100ms

    def test_memory_and_resource_management(self):
        """Test memory and resource management"""
        import psutil
        import os
        
        # Get initial memory usage
        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB
        
        # Create and delete many agents to test memory management
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
                with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
                    mock_wrapper = MagicMock()
                    mock_wrapper.agent_id = "memory-agent"
                    mock_wrapper.tools = [MagicMock()]
                    mock_wrapper.initialize = AsyncMock()
                    mock_wrapper.agent_config = MagicMock()
                    mock_wrapper.agent_config.name = "Memory Agent"
                    mock_wrapper.agent_config.description = "Agent for memory testing"
                    mock_wrapper.agent_config.capabilities = ["calculations"]
                    mock_wrapper.agent = MagicMock()
                    mock_wrapper_class.return_value = mock_wrapper
                    
                    # Create and delete agents in cycles
                    for cycle in range(10):
                        # Create agents
                        agent_ids = []
                        for i in range(10):
                            config = {
                                "name": f"Memory Agent {cycle}-{i}",
                                "description": f"Memory test agent {cycle}-{i}",
                                "capabilities": ["calculations"]
                            }
                            response = self.client.post("/agents/create", json=config)
                            if response.status_code == 200:
                                agent_ids.append(response.json()["agent_id"])
                        
                        # Delete agents
                        for agent_id in agent_ids:
                            self.client.delete(f"/agents/{agent_id}")
                    
                    # Check final memory usage
                    final_memory = process.memory_info().rss / 1024 / 1024  # MB
                    memory_increase = final_memory - initial_memory
                    
                    # Memory increase should be reasonable (less than 20MB)
                    assert memory_increase < 20
