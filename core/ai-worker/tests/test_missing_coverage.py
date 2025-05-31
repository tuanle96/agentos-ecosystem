"""
Tests to cover missing lines and improve coverage to 80%+
"""
import pytest
import os
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
from fastapi import HTTPException
import main


class TestMissingCoverage:
    """Test missing coverage areas to reach 80%+"""

    def setup_method(self):
        """Setup test client and clear registry"""
        self.client = TestClient(main.app)
        main.agent_registry.clear()

    def test_multi_framework_import_error_coverage(self):
        """Test multi-framework import error handling (lines 56-61)"""
        # Test that MULTI_FRAMEWORK_AVAILABLE is properly set
        assert hasattr(main, 'MULTI_FRAMEWORK_AVAILABLE')
        assert hasattr(main, 'SWARMS_AVAILABLE')
        assert hasattr(main, 'CREWAI_AVAILABLE')
        assert hasattr(main, 'AUTOGEN_AVAILABLE')
        
        # These should be False due to import errors
        assert main.MULTI_FRAMEWORK_AVAILABLE is False
        assert main.SWARMS_AVAILABLE is False
        assert main.CREWAI_AVAILABLE is False
        assert main.AUTOGEN_AVAILABLE is False

    def test_langchain_wrapper_without_openai_key(self):
        """Test LangChainWrapper initialization without OpenAI key (lines 111-112)"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        # Test without OpenAI API key
        with patch.dict(os.environ, {}, clear=True):
            with patch('main.LANGCHAIN_AVAILABLE', True):
                with patch('main.OpenAI', MagicMock()):
                    with patch('main.ConversationBufferMemory', MagicMock()):
                        wrapper = LangChainAgentWrapper(config)
                        
                        # Should have llm as None due to missing API key
                        assert wrapper.llm is None

    @pytest.mark.asyncio
    async def test_langchain_wrapper_initialization_error_handling(self):
        """Test LangChain wrapper initialization error handling (lines 118-132)"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        # Test with LangChain not available
        with patch('main.LANGCHAIN_AVAILABLE', False):
            wrapper = LangChainAgentWrapper(config)
            
            with pytest.raises(HTTPException) as exc_info:
                await wrapper.initialize()
            
            assert exc_info.value.status_code == 500
            assert "LangChain not available" in str(exc_info.value.detail)

    @pytest.mark.asyncio
    async def test_langchain_wrapper_no_openai_key_error(self):
        """Test LangChain wrapper without OpenAI key error (lines 121-122)"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        # Test with LangChain available but no OpenAI key
        with patch('main.LANGCHAIN_AVAILABLE', True):
            wrapper = LangChainAgentWrapper(config)
            wrapper.llm = None  # Simulate no OpenAI key
            
            with pytest.raises(HTTPException) as exc_info:
                await wrapper.initialize()
            
            assert exc_info.value.status_code == 500
            assert "OpenAI API key not configured" in str(exc_info.value.detail)

    @pytest.mark.asyncio
    async def test_tool_creation_without_langchain(self):
        """Test tool creation when LangChain is not available (lines 156-160, 171-179, etc.)"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["web_search", "calculations", "text_processing", "file_operations", "api_calls"]
        )
        
        # Test with LangChain not available
        with patch('main.LANGCHAIN_AVAILABLE', False):
            with patch('main.Tool', None):
                wrapper = LangChainAgentWrapper(config)
                
                # All tool creation methods should return None
                assert wrapper._create_web_search_tool() is None
                assert wrapper._create_calculator_tool() is None
                assert wrapper._create_text_processing_tool() is None
                assert wrapper._create_file_operations_tool() is None
                assert wrapper._create_api_calls_tool() is None

    @pytest.mark.asyncio
    async def test_agent_execution_without_proper_initialization(self):
        """Test agent execution without proper initialization (lines 232-254)"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        wrapper.agent = None  # Not properly initialized
        
        # Mock initialize to avoid actual initialization
        with patch.object(wrapper, 'initialize', new_callable=AsyncMock) as mock_init:
            # Make initialize not create a proper agent
            async def mock_initialize():
                wrapper.agent = None  # Still no agent after init
            
            mock_init.side_effect = mock_initialize
            
            result = await wrapper.execute("Test task")
            
            mock_init.assert_called_once()
            assert result["status"] == "completed"
            assert "Agent not properly initialized" in result["result"]

    @pytest.mark.asyncio
    async def test_agent_execution_with_mock_agent(self):
        """Test agent execution with mock agent that has run method"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        # Create a mock agent with run method
        mock_agent = MagicMock()
        mock_agent.run.return_value = "Mock execution result"
        wrapper.agent = mock_agent
        
        result = await wrapper.execute("Test task")
        
        assert result["status"] == "completed"
        assert result["result"] == "Mock execution result"
        assert "task_id" in result
        assert "execution_time" in result
        mock_agent.run.assert_called_once_with("Test task")

    @pytest.mark.asyncio
    async def test_agent_execution_exception_handling(self):
        """Test agent execution exception handling"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        # Create a mock agent that raises an exception
        mock_agent = MagicMock()
        mock_agent.run.side_effect = Exception("Mock execution error")
        wrapper.agent = mock_agent
        
        result = await wrapper.execute("Test task")
        
        assert result["status"] == "failed"
        assert "Error: Mock execution error" in result["result"]
        assert "task_id" in result
        assert "execution_time" in result

    def test_create_agent_endpoint_with_initialization_error(self):
        """Test create agent endpoint with initialization error (lines 311-312)"""
        agent_config = {
            "name": "Failing Agent",
            "description": "Agent that fails to initialize",
            "capabilities": ["calculations"]
        }
        
        with patch('main.LangChainAgentWrapper') as mock_wrapper_class:
            # Mock wrapper that fails during initialization
            mock_wrapper = MagicMock()
            mock_wrapper.initialize = AsyncMock(side_effect=Exception("Initialization failed"))
            mock_wrapper_class.return_value = mock_wrapper
            
            response = self.client.post("/agents/create", json=agent_config)
            
            assert response.status_code == 500
            assert "Initialization failed" in response.json()["detail"]

    def test_execute_task_endpoint_with_execution_error(self):
        """Test execute task endpoint with execution error (lines 329-330)"""
        # Create a mock agent that fails during execution
        mock_wrapper = MagicMock()
        mock_wrapper.execute = AsyncMock(side_effect=Exception("Execution failed"))
        
        main.agent_registry["failing-agent"] = mock_wrapper
        
        task_request = {
            "agent_id": "failing-agent",
            "task": "Failing task"
        }
        
        response = self.client.post("/agents/failing-agent/execute", json=task_request)
        
        assert response.status_code == 500
        assert "Execution failed" in response.json()["detail"]

    @pytest.mark.asyncio
    async def test_capability_to_tool_with_all_capabilities(self):
        """Test capability to tool conversion with all supported capabilities"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["web_search", "calculations", "text_processing", "file_operations", "api_calls"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        # Test with LangChain available
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch('main.Tool', MagicMock()) as mock_tool:
                mock_tool.return_value = MagicMock()
                
                # Test all supported capabilities
                for capability in ["web_search", "calculations", "text_processing", "file_operations", "api_calls"]:
                    tool = await wrapper._capability_to_tool(capability)
                    assert tool is not None
                
                # Test unsupported capability
                tool = await wrapper._capability_to_tool("unsupported_capability")
                assert tool is None

    def test_tool_function_implementations(self):
        """Test the actual tool function implementations"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["web_search", "calculations", "text_processing", "file_operations", "api_calls"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        # Test with LangChain available
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch('main.Tool') as mock_tool:
                # Test web search tool function
                wrapper._create_web_search_tool()
                call_args = mock_tool.call_args
                web_search_func = call_args[1]['func']
                result = web_search_func("test query")
                assert result == "Search results for: test query"
                
                # Test calculator tool function
                mock_tool.reset_mock()
                wrapper._create_calculator_tool()
                call_args = mock_tool.call_args
                calc_func = call_args[1]['func']
                
                # Test valid calculation
                result = calc_func("2+2")
                assert result == "4"
                
                # Test invalid calculation
                result = calc_func("invalid")
                assert "Error:" in result
                
                # Test text processing tool function
                mock_tool.reset_mock()
                wrapper._create_text_processing_tool()
                call_args = mock_tool.call_args
                text_func = call_args[1]['func']
                result = text_func("  HELLO WORLD  ")
                assert result == "Processed: hello world"
                
                # Test file operations tool function
                mock_tool.reset_mock()
                wrapper._create_file_operations_tool()
                call_args = mock_tool.call_args
                file_func = call_args[1]['func']
                result = file_func("read file.txt")
                assert result == "File operation: read file.txt"
                
                # Test API calls tool function
                mock_tool.reset_mock()
                wrapper._create_api_calls_tool()
                call_args = mock_tool.call_args
                api_func = call_args[1]['func']
                result = api_func("https://api.example.com")
                assert result == "API call to: https://api.example.com"

    @pytest.mark.asyncio
    async def test_agent_initialization_with_tools(self):
        """Test agent initialization with tools"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations", "web_search"]
        )
        
        with patch('main.LANGCHAIN_AVAILABLE', True):
            with patch('main.OpenAI') as mock_openai:
                with patch('main.ConversationBufferMemory') as mock_memory:
                    with patch('main.Tool') as mock_tool:
                        with patch('main.initialize_agent') as mock_init_agent:
                            with patch('main.AgentType') as mock_agent_type:
                                # Setup mocks
                                mock_openai.return_value = MagicMock()
                                mock_memory.return_value = MagicMock()
                                mock_tool.return_value = MagicMock()
                                mock_init_agent.return_value = MagicMock()
                                mock_agent_type.CONVERSATIONAL_REACT_DESCRIPTION = "mock_agent_type"
                                
                                with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
                                    wrapper = LangChainAgentWrapper(config)
                                    await wrapper.initialize()
                                    
                                    # Should have created tools
                                    assert len(wrapper.tools) == 2
                                    
                                    # Should have initialized agent
                                    mock_init_agent.assert_called_once()

    def test_edge_case_coverage(self):
        """Test edge cases for better coverage"""
        # Test empty agent registry
        response = self.client.get("/agents")
        assert response.status_code == 200
        assert response.json()["count"] == 0
        
        # Test framework status with no agents
        response = self.client.get("/framework/status")
        assert response.status_code == 200
        assert response.json()["active_agents"] == 0
        
        # Test tools endpoint
        response = self.client.get("/tools")
        assert response.status_code == 200
        assert response.json()["count"] == 5
