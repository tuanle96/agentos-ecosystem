"""
Tests for agent execution and advanced functionality
"""
import pytest
import asyncio
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
from fastapi import HTTPException
import main


class TestAgentExecution:
    """Test agent execution functionality"""

    @pytest.mark.asyncio
    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.initialize_agent')
    async def test_agent_execution_success(self, mock_init_agent):
        """Test successful agent execution"""
        from main import LangChainAgentWrapper, AgentConfig
        
        # Mock the agent
        mock_agent = MagicMock()
        mock_agent.run.return_value = "Task completed successfully"
        mock_init_agent.return_value = mock_agent
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        with patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'}):
            wrapper = LangChainAgentWrapper(config)
            wrapper.agent = mock_agent  # Set agent directly
            
            result = await wrapper.execute("Calculate 2+2")
            
            assert result["status"] == "completed"
            assert result["result"] == "Task completed successfully"
            assert "task_id" in result
            assert "execution_time" in result
            assert result["agent_id"] == wrapper.agent_id
            assert result["framework"] == "langchain"

    @pytest.mark.asyncio
    @patch('main.LANGCHAIN_AVAILABLE', True)
    async def test_agent_execution_failure(self):
        """Test agent execution failure handling"""
        from main import LangChainAgentWrapper, AgentConfig
        
        # Mock the agent to raise an exception
        mock_agent = MagicMock()
        mock_agent.run.side_effect = Exception("Execution failed")
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        wrapper.agent = mock_agent
        
        result = await wrapper.execute("Calculate 2+2")
        
        assert result["status"] == "failed"
        assert "Error: Execution failed" in result["result"]
        assert "task_id" in result
        assert "execution_time" in result
        assert result["agent_id"] == wrapper.agent_id
        assert result["framework"] == "langchain"

    @pytest.mark.asyncio
    @patch('main.LANGCHAIN_AVAILABLE', True)
    async def test_agent_execution_without_initialization(self):
        """Test agent execution when agent is not initialized"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        wrapper.agent = None  # Not initialized
        
        # Mock the initialize method to avoid actual initialization
        with patch.object(wrapper, 'initialize', new_callable=AsyncMock) as mock_init:
            mock_agent = MagicMock()
            mock_agent.run.return_value = "Initialized and executed"
            
            async def mock_initialize():
                wrapper.agent = mock_agent
            
            mock_init.side_effect = mock_initialize
            
            result = await wrapper.execute("Test task")
            
            mock_init.assert_called_once()
            assert result["status"] == "completed"
            assert result["result"] == "Initialized and executed"

    @pytest.mark.asyncio
    async def test_agent_execution_timing(self):
        """Test that execution timing is properly measured"""
        from main import LangChainAgentWrapper, AgentConfig
        import time
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        # Mock agent with delay
        mock_agent = MagicMock()
        def slow_run(task):
            time.sleep(0.1)  # 100ms delay
            return "Slow result"
        
        mock_agent.run.side_effect = slow_run
        wrapper.agent = mock_agent
        
        result = await wrapper.execute("Slow task")
        
        assert result["execution_time"] >= 0.1
        assert result["status"] == "completed"


class TestAPIEndpoints:
    """Test API endpoints with comprehensive scenarios"""

    def setup_method(self):
        """Setup test client"""
        self.client = TestClient(main.app)
        # Clear agent registry before each test
        main.agent_registry.clear()

    def test_create_agent_endpoint_success(self):
        """Test successful agent creation via API"""
        agent_config = {
            "name": "Test API Agent",
            "description": "Agent created via API",
            "capabilities": ["calculations", "text_processing"],
            "framework_preference": "langchain"
        }
        
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'}):
                with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
                    # Mock the wrapper instance
                    mock_wrapper = MagicMock()
                    mock_wrapper.agent_id = "test-agent-123"
                    mock_wrapper.tools = [MagicMock(), MagicMock()]
                    mock_wrapper.initialize = AsyncMock()
                    mock_wrapper_class.return_value = mock_wrapper
                    
                    response = self.client.post("/agents/create", json=agent_config)
                    
                    assert response.status_code == 200
                    data = response.json()
                    assert data["agent_id"] == "test-agent-123"
                    assert data["name"] == "Test API Agent"
                    assert data["capabilities"] == ["calculations", "text_processing"]
                    assert data["framework"] == "langchain"
                    assert data["status"] == "created"
                    assert data["tools_count"] == 2

    def test_create_agent_endpoint_failure(self):
        """Test agent creation failure via API"""
        agent_config = {
            "name": "Failing Agent",
            "description": "Agent that fails to create",
            "capabilities": ["calculations"]
        }
        
        with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
            mock_wrapper_class.side_effect = Exception("Creation failed")
            
            response = self.client.post("/agents/create", json=agent_config)
            
            assert response.status_code == 500
            assert "Creation failed" in response.json()["detail"]

    def test_execute_task_endpoint_success(self):
        """Test successful task execution via API"""
        # First create an agent
        mock_wrapper = MagicMock()
        mock_wrapper.execute = AsyncMock(return_value={
            "task_id": "task-123",
            "result": "Calculation result: 4",
            "status": "completed",
            "execution_time": 0.05,
            "agent_id": "agent-123",
            "framework": "langchain"
        })
        
        main.agent_registry["agent-123"] = mock_wrapper
        
        task_request = {
            "agent_id": "agent-123",
            "task": "Calculate 2+2",
            "context": {"priority": "high"}
        }
        
        response = self.client.post("/agents/agent-123/execute", json=task_request)
        
        assert response.status_code == 200
        data = response.json()
        assert data["task_id"] == "task-123"
        assert data["result"] == "Calculation result: 4"
        assert data["status"] == "completed"
        assert data["execution_time"] == 0.05

    def test_execute_task_agent_not_found(self):
        """Test task execution with non-existent agent"""
        task_request = {
            "agent_id": "nonexistent-agent",
            "task": "Calculate 2+2"
        }
        
        response = self.client.post("/agents/nonexistent-agent/execute", json=task_request)
        
        assert response.status_code == 404
        assert "Agent not found" in response.json()["detail"]

    def test_execute_task_endpoint_failure(self):
        """Test task execution failure via API"""
        # Create a mock agent that fails
        mock_wrapper = MagicMock()
        mock_wrapper.execute = AsyncMock(side_effect=Exception("Execution error"))
        
        main.agent_registry["failing-agent"] = mock_wrapper
        
        task_request = {
            "agent_id": "failing-agent",
            "task": "Failing task"
        }
        
        response = self.client.post("/agents/failing-agent/execute", json=task_request)
        
        assert response.status_code == 500
        assert "Execution error" in response.json()["detail"]

    def test_get_agent_endpoint_success(self):
        """Test successful agent retrieval via API"""
        # Create a mock agent
        mock_config = MagicMock()
        mock_config.name = "Test Agent"
        mock_config.description = "Test Description"
        mock_config.capabilities = ["calculations", "text_processing"]
        
        mock_wrapper = MagicMock()
        mock_wrapper.agent_config = mock_config
        mock_wrapper.tools = [MagicMock(), MagicMock()]
        mock_wrapper.agent = MagicMock()  # Agent is initialized
        
        main.agent_registry["test-agent"] = mock_wrapper
        
        response = self.client.get("/agents/test-agent")
        
        assert response.status_code == 200
        data = response.json()
        assert data["agent_id"] == "test-agent"
        assert data["name"] == "Test Agent"
        assert data["description"] == "Test Description"
        assert data["capabilities"] == ["calculations", "text_processing"]
        assert data["framework"] == "langchain"
        assert data["tools_count"] == 2
        assert data["status"] == "active"

    def test_get_agent_not_found(self):
        """Test agent retrieval with non-existent agent"""
        response = self.client.get("/agents/nonexistent-agent")
        
        assert response.status_code == 404
        assert "Agent not found" in response.json()["detail"]

    def test_delete_agent_endpoint_success(self):
        """Test successful agent deletion via API"""
        # Create a mock agent
        mock_wrapper = MagicMock()
        main.agent_registry["delete-me"] = mock_wrapper
        
        response = self.client.delete("/agents/delete-me")
        
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "Agent deleted successfully"
        assert data["agent_id"] == "delete-me"
        assert "delete-me" not in main.agent_registry

    def test_delete_agent_not_found(self):
        """Test agent deletion with non-existent agent"""
        response = self.client.delete("/agents/nonexistent-agent")
        
        assert response.status_code == 404
        assert "Agent not found" in response.json()["detail"]

    def test_list_agents_with_agents(self):
        """Test listing agents when agents exist"""
        # Add some mock agents
        main.agent_registry["agent-1"] = MagicMock()
        main.agent_registry["agent-2"] = MagicMock()
        main.agent_registry["agent-3"] = MagicMock()
        
        response = self.client.get("/agents")
        
        assert response.status_code == 200
        data = response.json()
        assert data["count"] == 3
        assert set(data["agents"]) == {"agent-1", "agent-2", "agent-3"}
        assert data["available_frameworks"] == ["langchain"]
        assert "langchain_available" in data

    def test_framework_status_endpoint(self):
        """Test framework status endpoint"""
        # Add some agents to test active count
        main.agent_registry["agent-1"] = MagicMock()
        main.agent_registry["agent-2"] = MagicMock()
        
        with patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'}):
            response = self.client.get("/framework/status")
            
            assert response.status_code == 200
            data = response.json()
            assert data["framework"] == "langchain"
            assert "available" in data
            assert data["openai_configured"] is True
            assert data["active_agents"] == 2
            assert "web_search" in data["supported_capabilities"]
            assert "calculations" in data["supported_capabilities"]
            assert data["version"] == "0.1.0-week2"

    def test_tools_endpoint(self):
        """Test tools listing endpoint"""
        response = self.client.get("/tools")
        
        assert response.status_code == 200
        data = response.json()
        assert data["count"] == 5
        assert data["framework"] == "langchain"
        
        tool_names = [tool["name"] for tool in data["tools"]]
        assert "web_search" in tool_names
        assert "calculator" in tool_names
        assert "text_processor" in tool_names
        assert "file_operations" in tool_names
        assert "api_calls" in tool_names
        
        # Check tool structure
        first_tool = data["tools"][0]
        assert "name" in first_tool
        assert "description" in first_tool
        assert "category" in first_tool
