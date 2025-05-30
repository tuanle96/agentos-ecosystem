#!/usr/bin/env python3
"""
AgentOS AI Worker - Swarms Integration Tests
Week 3 Day 1 Implementation: Test Swarms Framework Integration

This module tests the Swarms framework integration for AgentOS AI Worker.
"""

import pytest
import asyncio
import os
from unittest.mock import Mock, patch, AsyncMock
from typing import Dict, Any

# Import framework components
try:
    from frameworks.swarms_wrapper import SwarmAgentWrapper
    from frameworks.base_wrapper import AgentConfig, TaskRequest, FrameworkType
    SWARMS_WRAPPER_AVAILABLE = True
except ImportError:
    SWARMS_WRAPPER_AVAILABLE = False

@pytest.mark.skipif(not SWARMS_WRAPPER_AVAILABLE, reason="Swarms wrapper not available")
class TestSwarmsIntegration:
    """Test suite for Swarms framework integration"""
    
    @pytest.fixture
    def sample_agent_config(self):
        """Sample agent configuration for testing"""
        return AgentConfig(
            name="test_swarm_agent",
            description="Test agent for Swarms integration",
            capabilities=["web_search", "calculations", "text_processing"],
            personality={"role": "assistant", "style": "helpful"},
            framework_preference="swarms",
            max_iterations=5,
            timeout=60,
            temperature=0.7,
            model="gpt-3.5-turbo"
        )
    
    @pytest.fixture
    def sample_task_request(self):
        """Sample task request for testing"""
        return TaskRequest(
            task="Calculate the sum of 2 + 2 and explain the result",
            context={"user_id": "test_user", "session_id": "test_session"},
            tools=["calculations", "text_processing"],
            max_iterations=3,
            timeout=30
        )
    
    def test_swarm_wrapper_initialization(self, sample_agent_config):
        """Test Swarms wrapper initialization"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        # Check basic properties
        assert wrapper.agent_config == sample_agent_config
        assert wrapper.framework_type == FrameworkType.SWARMS
        assert wrapper.agent_id is not None
        assert not wrapper.is_initialized
        assert wrapper.tools == []
        assert wrapper.swarm_agent is None
    
    @pytest.mark.asyncio
    async def test_swarm_wrapper_initialization_without_openai_key(self, sample_agent_config):
        """Test Swarms wrapper initialization fails without OpenAI API key"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(Exception):  # Should raise InitializationError
                await wrapper.initialize()
    
    @pytest.mark.asyncio
    @patch.dict(os.environ, {"OPENAI_API_KEY": "test_key"})
    @patch('frameworks.swarms_wrapper.SWARMS_AVAILABLE', True)
    @patch('frameworks.swarms_wrapper.OpenAIChat')
    @patch('frameworks.swarms_wrapper.Agent')
    @patch('frameworks.swarms_wrapper.Flow')
    async def test_swarm_wrapper_successful_initialization(
        self, mock_flow, mock_agent, mock_openai_chat, sample_agent_config
    ):
        """Test successful Swarms wrapper initialization"""
        # Setup mocks
        mock_llm = Mock()
        mock_openai_chat.return_value = mock_llm
        
        mock_swarm_agent = Mock()
        mock_agent.return_value = mock_swarm_agent
        
        mock_flow_instance = Mock()
        mock_flow.return_value = mock_flow_instance
        
        wrapper = SwarmAgentWrapper(sample_agent_config)
        result = await wrapper.initialize()
        
        # Verify initialization
        assert result is True
        assert wrapper.is_initialized is True
        assert wrapper.llm == mock_llm
        assert wrapper.swarm_agent == mock_swarm_agent
        assert wrapper.flow == mock_flow_instance
        
        # Verify OpenAI LLM was created with correct parameters
        mock_openai_chat.assert_called_once_with(
            model_name="gpt-3.5-turbo",
            temperature=0.7,
            max_tokens=2000,
            openai_api_key="test_key"
        )
        
        # Verify Swarms agent was created
        mock_agent.assert_called_once()
        agent_call_kwargs = mock_agent.call_args[1]
        assert agent_call_kwargs["agent_name"] == "test_swarm_agent"
        assert agent_call_kwargs["agent_description"] == "Test agent for Swarms integration"
        assert agent_call_kwargs["llm"] == mock_llm
        assert agent_call_kwargs["max_loops"] == 5
    
    def test_capability_to_tool_conversion(self, sample_agent_config):
        """Test conversion of AgentOS capabilities to Swarms tools"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        # Test web search tool
        web_search_tool = asyncio.run(wrapper._capability_to_tool("web_search"))
        assert web_search_tool is not None
        assert web_search_tool["name"] == "web_search"
        assert "function" in web_search_tool
        
        # Test calculator tool
        calculator_tool = asyncio.run(wrapper._capability_to_tool("calculations"))
        assert calculator_tool is not None
        assert calculator_tool["name"] == "calculator"
        assert "function" in calculator_tool
        
        # Test text processing tool
        text_tool = asyncio.run(wrapper._capability_to_tool("text_processing"))
        assert text_tool is not None
        assert text_tool["name"] == "text_processor"
        assert "function" in text_tool
        
        # Test unknown capability
        unknown_tool = asyncio.run(wrapper._capability_to_tool("unknown_capability"))
        assert unknown_tool is None
    
    def test_tool_functions(self, sample_agent_config):
        """Test individual tool functions"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        # Test web search tool function
        web_search_tool = asyncio.run(wrapper._capability_to_tool("web_search"))
        result = web_search_tool["function"]("test query")
        assert "Swarms web search results for: test query" in result
        
        # Test calculator tool function
        calculator_tool = asyncio.run(wrapper._capability_to_tool("calculations"))
        result = calculator_tool["function"]("2 + 2")
        assert "Calculation result: 4" in result
        
        # Test calculator error handling
        error_result = calculator_tool["function"]("invalid expression")
        assert "Calculation error:" in error_result
        
        # Test text processing tool function
        text_tool = asyncio.run(wrapper._capability_to_tool("text_processing"))
        result = text_tool["function"]("Test Text", "analyze")
        assert "Swarms text analysis:" in result
        assert "10 characters" in result
        assert "2 words" in result
    
    @pytest.mark.asyncio
    @patch.dict(os.environ, {"OPENAI_API_KEY": "test_key"})
    @patch('frameworks.swarms_wrapper.SWARMS_AVAILABLE', True)
    async def test_swarm_wrapper_execute_without_initialization(self, sample_agent_config, sample_task_request):
        """Test task execution without initialization triggers auto-initialization"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        with patch.object(wrapper, 'initialize', new_callable=AsyncMock) as mock_init:
            with patch.object(wrapper, '_run_swarm_task', new_callable=AsyncMock) as mock_run:
                mock_init.return_value = True
                mock_run.return_value = "Test result"
                
                response = await wrapper.execute(sample_task_request)
                
                # Verify initialization was called
                mock_init.assert_called_once()
                
                # Verify response structure
                assert response.status == "completed"
                assert response.result == "Test result"
                assert response.framework_used == "swarms"
                assert response.agent_id == wrapper.agent_id
    
    @pytest.mark.asyncio
    @patch.dict(os.environ, {"OPENAI_API_KEY": "test_key"})
    @patch('frameworks.swarms_wrapper.SWARMS_AVAILABLE', True)
    async def test_swarm_wrapper_execute_with_timeout(self, sample_agent_config):
        """Test task execution with timeout"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        wrapper.is_initialized = True
        wrapper.swarm_agent = Mock()
        
        # Create task request with short timeout
        task_request = TaskRequest(
            task="Long running task",
            timeout=0.1  # Very short timeout
        )
        
        with patch.object(wrapper, '_run_swarm_task', new_callable=AsyncMock) as mock_run:
            # Make the task take longer than timeout
            async def slow_task(*args):
                await asyncio.sleep(0.2)
                return "Should not reach here"
            
            mock_run.side_effect = slow_task
            
            response = await wrapper.execute(task_request)
            
            # Should fail due to timeout
            assert response.status == "failed"
            assert "timeout" in response.error_message.lower() or "timed out" in response.error_message.lower()
    
    @pytest.mark.asyncio
    async def test_swarm_wrapper_cleanup(self, sample_agent_config):
        """Test Swarms wrapper cleanup"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        wrapper.is_initialized = True
        
        # Mock swarm agent with save_state method
        mock_agent = Mock()
        mock_agent.save_state = Mock()
        wrapper.swarm_agent = mock_agent
        
        result = await wrapper.cleanup()
        
        assert result is True
        assert wrapper.is_initialized is False
        mock_agent.save_state.assert_called_once()
    
    def test_swarm_wrapper_get_agent_info(self, sample_agent_config):
        """Test getting agent information"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        info = wrapper.get_agent_info()
        
        assert info["agent_id"] == wrapper.agent_id
        assert info["framework"] == "swarms"
        assert info["name"] == "test_swarm_agent"
        assert info["description"] == "Test agent for Swarms integration"
        assert info["capabilities"] == ["web_search", "calculations", "text_processing"]
        assert info["is_initialized"] is False
        assert info["tools_count"] == 0
        assert "created_at" in info
        assert "uptime" in info
    
    def test_swarm_wrapper_get_performance_metrics(self, sample_agent_config):
        """Test getting performance metrics"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        metrics = wrapper.get_performance_metrics()
        
        assert metrics["agent_id"] == wrapper.agent_id
        assert metrics["framework"] == "swarms"
        assert metrics["is_healthy"] is False  # Not initialized
        assert "uptime" in metrics
        assert "memory_usage" in metrics
        assert "tools_available" in metrics
    
    def test_swarm_wrapper_memory_usage(self, sample_agent_config):
        """Test memory usage calculation"""
        wrapper = SwarmAgentWrapper(sample_agent_config)
        
        memory_usage = wrapper._get_memory_usage()
        
        assert "working_memory" in memory_usage
        assert "episodic_memory" in memory_usage
        assert "semantic_memory" in memory_usage
        assert "agent_state_size" in memory_usage
        
        # Should be 0 when no agent is initialized
        assert memory_usage["working_memory"] == 0
        assert memory_usage["agent_state_size"] == 0

@pytest.mark.skipif(SWARMS_WRAPPER_AVAILABLE, reason="Testing import failure")
def test_swarms_wrapper_import_failure():
    """Test behavior when Swarms wrapper cannot be imported"""
    # This test runs when the import fails
    # Just verify that the test framework handles the import failure gracefully
    assert not SWARMS_WRAPPER_AVAILABLE
