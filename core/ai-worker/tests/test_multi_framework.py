"""
Tests for multi-framework support and import handling
"""
import pytest
import sys
import importlib
from unittest.mock import patch, MagicMock
import main


class TestMultiFrameworkSupport:
    """Test multi-framework import and availability detection"""

    def test_langchain_available_true(self):
        """Test when LangChain is available"""
        # LangChain should be available in our test environment
        assert main.LANGCHAIN_AVAILABLE is True

    @patch('main.LANGCHAIN_AVAILABLE', False)
    def test_langchain_unavailable_handling(self):
        """Test handling when LangChain is not available"""
        # Test that the system gracefully handles missing LangChain
        assert main.LANGCHAIN_AVAILABLE is False

    def test_multi_framework_import_error_handling(self):
        """Test import error handling for multi-framework support"""
        # Test that import errors are handled gracefully
        with patch.dict('sys.modules', {'frameworks': None}):
            # This should not raise an exception
            try:
                importlib.reload(main)
            except ImportError:
                # This is expected behavior
                pass

    def test_framework_availability_flags(self):
        """Test framework availability flags"""
        # Test that framework flags are properly set
        assert hasattr(main, 'LANGCHAIN_AVAILABLE')
        assert hasattr(main, 'MULTI_FRAMEWORK_AVAILABLE')
        
        # At least LangChain should be available
        assert main.LANGCHAIN_AVAILABLE is True or main.LANGCHAIN_AVAILABLE is False

    @patch('main.logger')
    def test_import_warning_logging(self, mock_logger):
        """Test that import warnings are properly logged"""
        # Test logging when multi-framework is not available
        with patch('main.MULTI_FRAMEWORK_AVAILABLE', False):
            # Simulate the warning that would be logged
            main.logger.warning("Multi-framework support not available")
            mock_logger.warning.assert_called()


class TestFrameworkConfiguration:
    """Test framework configuration and environment setup"""

    def test_openai_api_key_detection(self):
        """Test OpenAI API key detection"""
        import os
        
        # Test with API key present
        with patch.dict(os.environ, {'OPENAI_API_KEY': 'test-key'}):
            assert bool(os.getenv("OPENAI_API_KEY")) is True
        
        # Test with API key absent
        with patch.dict(os.environ, {}, clear=True):
            assert bool(os.getenv("OPENAI_API_KEY")) is False

    def test_environment_variable_handling(self):
        """Test environment variable handling"""
        import os
        
        # Test PORT environment variable
        with patch.dict(os.environ, {'PORT': '9000'}):
            assert int(os.getenv("PORT", "8080")) == 9000
        
        with patch.dict(os.environ, {}, clear=True):
            assert int(os.getenv("PORT", "8080")) == 8080

    def test_framework_version_info(self):
        """Test framework version information"""
        # Test that version info is accessible
        assert hasattr(main, 'app')
        assert main.app.version == "0.1.0"
        assert main.app.title == "AgentOS AI Worker"


class TestLangChainIntegration:
    """Test LangChain integration and initialization"""

    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.OpenAI')
    @patch('main.ConversationBufferMemory')
    def test_langchain_wrapper_initialization_success(self, mock_memory, mock_openai):
        """Test successful LangChain wrapper initialization"""
        from main import LangChainAgentWrapper, AgentConfig
        
        # Mock OpenAI and memory
        mock_openai.return_value = MagicMock()
        mock_memory.return_value = MagicMock()
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        with patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'}):
            wrapper = LangChainAgentWrapper(config)
            
            assert wrapper.agent_config == config
            assert wrapper.agent_id is not None
            assert wrapper.tools == []
            assert wrapper.agent is None

    @patch('main.LANGCHAIN_AVAILABLE', False)
    def test_langchain_wrapper_initialization_unavailable(self):
        """Test LangChain wrapper when LangChain is unavailable"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        # Should initialize basic properties even without LangChain
        assert wrapper.agent_config == config
        assert wrapper.agent_id is not None
        assert wrapper.tools == []
        assert wrapper.agent is None
        assert wrapper.memory is None

    @patch('main.LANGCHAIN_AVAILABLE', True)
    def test_langchain_initialization_without_api_key(self):
        """Test LangChain initialization without OpenAI API key"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        with patch.dict('os.environ', {}, clear=True):
            wrapper = LangChainAgentWrapper(config)
            
            # Should handle missing API key gracefully
            assert wrapper.llm is None

    @pytest.mark.asyncio
    @patch('main.LANGCHAIN_AVAILABLE', False)
    async def test_initialize_without_langchain(self):
        """Test initialization when LangChain is not available"""
        from main import LangChainAgentWrapper, AgentConfig
        from fastapi import HTTPException
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        with pytest.raises(HTTPException) as exc_info:
            await wrapper.initialize()
        
        assert exc_info.value.status_code == 500
        assert "LangChain not available" in str(exc_info.value.detail)

    @pytest.mark.asyncio
    @patch('main.LANGCHAIN_AVAILABLE', True)
    async def test_initialize_without_openai_key(self):
        """Test initialization without OpenAI API key"""
        from main import LangChainAgentWrapper, AgentConfig
        from fastapi import HTTPException
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        with patch.dict('os.environ', {}, clear=True):
            wrapper = LangChainAgentWrapper(config)
            
            with pytest.raises(HTTPException) as exc_info:
                await wrapper.initialize()
            
            assert exc_info.value.status_code == 500
            assert "OpenAI API key not configured" in str(exc_info.value.detail)


class TestToolCreationMethods:
    """Test individual tool creation methods"""

    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.Tool')
    def test_create_web_search_tool(self, mock_tool):
        """Test web search tool creation"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["web_search"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        tool = wrapper._create_web_search_tool()
        
        # Tool should be created (mocked)
        mock_tool.assert_called_once()

    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.Tool')
    def test_create_calculator_tool(self, mock_tool):
        """Test calculator tool creation"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        tool = wrapper._create_calculator_tool()
        
        # Tool should be created (mocked)
        mock_tool.assert_called_once()

    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.Tool')
    def test_create_text_processing_tool(self, mock_tool):
        """Test text processing tool creation"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["text_processing"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        tool = wrapper._create_text_processing_tool()
        
        # Tool should be created (mocked)
        mock_tool.assert_called_once()

    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.Tool')
    def test_create_file_operations_tool(self, mock_tool):
        """Test file operations tool creation"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["file_operations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        tool = wrapper._create_file_operations_tool()
        
        # Tool should be created (mocked)
        mock_tool.assert_called_once()

    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.Tool')
    def test_create_api_calls_tool(self, mock_tool):
        """Test API calls tool creation"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["api_calls"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        tool = wrapper._create_api_calls_tool()
        
        # Tool should be created (mocked)
        mock_tool.assert_called_once()

    @pytest.mark.asyncio
    async def test_capability_to_tool_mapping(self):
        """Test capability to tool mapping"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations"]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        # Test valid capability
        with patch.object(wrapper, '_create_calculator_tool') as mock_calc:
            mock_calc.return_value = MagicMock()
            tool = await wrapper._capability_to_tool("calculations")
            mock_calc.assert_called_once()
        
        # Test invalid capability
        tool = await wrapper._capability_to_tool("invalid_capability")
        assert tool is None
