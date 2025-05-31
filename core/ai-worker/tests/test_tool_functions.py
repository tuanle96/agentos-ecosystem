"""
Tests for individual tool functions and their logic
"""
import pytest
from unittest.mock import patch, MagicMock
import main


class TestToolFunctions:
    """Test individual tool function implementations"""

    def setup_method(self):
        """Setup test agent wrapper"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations", "web_search", "text_processing", "file_operations", "api_calls"]
        )
        
        self.wrapper = LangChainAgentWrapper(config)

    def test_web_search_tool_function(self):
        """Test web search tool function logic"""
        # Get the web search tool function
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_web_search_tool()
            
            # Get the function that was passed to Tool
            call_args = mock_tool.call_args
            assert call_args[1]['name'] == 'web_search'
            assert call_args[1]['description'] == 'Search the web for information'
            
            # Test the actual function
            web_search_func = call_args[1]['func']
            result = web_search_func("Python programming")
            assert result == "Search results for: Python programming"
            
            # Test with different queries
            result = web_search_func("machine learning")
            assert result == "Search results for: machine learning"
            
            result = web_search_func("")
            assert result == "Search results for: "

    def test_calculator_tool_function(self):
        """Test calculator tool function logic"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_calculator_tool()
            
            # Get the function that was passed to Tool
            call_args = mock_tool.call_args
            assert call_args[1]['name'] == 'calculator'
            assert call_args[1]['description'] == 'Perform mathematical calculations'
            
            # Test the actual function
            calc_func = call_args[1]['func']
            
            # Test valid expressions
            assert calc_func("2+2") == "4"
            assert calc_func("10*5") == "50"
            assert calc_func("100/4") == "25.0"
            assert calc_func("2**3") == "8"
            assert calc_func("(5+3)*2") == "16"
            
            # Test complex expressions
            assert calc_func("3.14159 * 2") == "6.28318"
            assert calc_func("abs(-5)") == "5"

    def test_calculator_tool_error_handling(self):
        """Test calculator tool error handling"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_calculator_tool()
            
            calc_func = mock_tool.call_args[1]['func']
            
            # Test invalid expressions
            result = calc_func("2/0")
            assert "Error:" in result
            
            result = calc_func("invalid_expression")
            assert "Error:" in result
            
            result = calc_func("import os")
            assert "Error:" in result
            
            result = calc_func("__import__('os')")
            assert "Error:" in result
            
            # Test empty expression
            result = calc_func("")
            assert "Error:" in result

    def test_calculator_tool_security(self):
        """Test calculator tool security restrictions"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_calculator_tool()
            
            calc_func = mock_tool.call_args[1]['func']
            
            # Test that dangerous operations are blocked
            dangerous_expressions = [
                "open('/etc/passwd')",
                "exec('import os')",
                "eval('__import__(\"os\")')",
                "__builtins__",
                "globals()",
                "locals()",
                "dir()",
                "vars()"
            ]
            
            for expr in dangerous_expressions:
                result = calc_func(expr)
                assert "Error:" in result

    def test_text_processing_tool_function(self):
        """Test text processing tool function logic"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_text_processing_tool()
            
            # Get the function that was passed to Tool
            call_args = mock_tool.call_args
            assert call_args[1]['name'] == 'text_processor'
            assert call_args[1]['description'] == 'Process and analyze text'
            
            # Test the actual function
            text_func = call_args[1]['func']
            
            # Test text processing
            result = text_func("  HELLO WORLD  ")
            assert result == "Processed: hello world"
            
            result = text_func("Python Programming")
            assert result == "Processed: python programming"
            
            result = text_func("   Mixed   Case   Text   ")
            assert result == "Processed: mixed   case   text"
            
            # Test empty text
            result = text_func("")
            assert result == "Processed: "
            
            # Test whitespace only
            result = text_func("   ")
            assert result == "Processed: "

    def test_file_operations_tool_function(self):
        """Test file operations tool function logic"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_file_operations_tool()
            
            # Get the function that was passed to Tool
            call_args = mock_tool.call_args
            assert call_args[1]['name'] == 'file_operations'
            assert call_args[1]['description'] == 'Perform safe file operations'
            
            # Test the actual function
            file_func = call_args[1]['func']
            
            # Test different operations
            result = file_func("read file.txt")
            assert result == "File operation: read file.txt"
            
            result = file_func("write data.json")
            assert result == "File operation: write data.json"
            
            result = file_func("list directory")
            assert result == "File operation: list directory"
            
            result = file_func("delete temp.log")
            assert result == "File operation: delete temp.log"
            
            # Test empty operation
            result = file_func("")
            assert result == "File operation: "

    def test_api_calls_tool_function(self):
        """Test API calls tool function logic"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_api_calls_tool()
            
            # Get the function that was passed to Tool
            call_args = mock_tool.call_args
            assert call_args[1]['name'] == 'api_calls'
            assert call_args[1]['description'] == 'Make HTTP API calls'
            
            # Test the actual function
            api_func = call_args[1]['func']
            
            # Test different URLs
            result = api_func("https://api.example.com/data")
            assert result == "API call to: https://api.example.com/data"
            
            result = api_func("http://localhost:8080/health")
            assert result == "API call to: http://localhost:8080/health"
            
            result = api_func("https://jsonplaceholder.typicode.com/posts/1")
            assert result == "API call to: https://jsonplaceholder.typicode.com/posts/1"
            
            # Test empty URL
            result = api_func("")
            assert result == "API call to: "


class TestToolIntegration:
    """Test tool integration with agent wrapper"""

    def setup_method(self):
        """Setup test agent wrapper"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Integration Test Agent",
            description="Test tool integration",
            capabilities=["calculations", "web_search", "text_processing"]
        )
        
        self.wrapper = LangChainAgentWrapper(config)

    @pytest.mark.asyncio
    async def test_capability_to_tool_all_types(self):
        """Test capability to tool conversion for all tool types"""
        with patch('main.Tool') as mock_tool:
            mock_tool.return_value = MagicMock()
            
            # Test all supported capabilities
            capabilities = ["web_search", "calculations", "text_processing", "file_operations", "api_calls"]
            
            for capability in capabilities:
                tool = await self.wrapper._capability_to_tool(capability)
                assert tool is not None
                mock_tool.assert_called()
                mock_tool.reset_mock()

    @pytest.mark.asyncio
    async def test_capability_to_tool_unsupported(self):
        """Test capability to tool conversion for unsupported capabilities"""
        unsupported_capabilities = [
            "unsupported_capability",
            "invalid_tool",
            "nonexistent_feature",
            "",
            None
        ]
        
        for capability in unsupported_capabilities:
            if capability is not None:
                tool = await self.wrapper._capability_to_tool(capability)
                assert tool is None

    @pytest.mark.asyncio
    @patch('main.LANGCHAIN_AVAILABLE', True)
    @patch('main.initialize_agent')
    async def test_tool_initialization_in_agent(self, mock_init_agent):
        """Test that tools are properly initialized in agent"""
        with patch('main.Tool') as mock_tool:
            mock_tool.return_value = MagicMock()
            mock_agent = MagicMock()
            mock_init_agent.return_value = mock_agent
            
            with patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'}):
                await self.wrapper.initialize()
                
                # Should have created tools for each capability
                assert len(self.wrapper.tools) == 3  # calculations, web_search, text_processing
                
                # Agent should be initialized with tools
                mock_init_agent.assert_called_once()
                call_args = mock_init_agent.call_args
                assert call_args[1]['tools'] == self.wrapper.tools

    @pytest.mark.asyncio
    @patch('main.LANGCHAIN_AVAILABLE', True)
    async def test_tool_initialization_without_tools(self):
        """Test agent initialization when no tools are created"""
        from main import LangChainAgentWrapper, AgentConfig
        
        # Create agent with no capabilities
        config = AgentConfig(
            name="No Tools Agent",
            description="Agent with no tools",
            capabilities=[]
        )
        
        wrapper = LangChainAgentWrapper(config)
        
        with patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'}):
            await wrapper.initialize()
            
            # Should have no tools
            assert len(wrapper.tools) == 0
            # Agent should not be initialized without tools
            assert wrapper.agent is None

    def test_tool_creation_method_coverage(self):
        """Test that all tool creation methods are accessible"""
        # Test that all tool creation methods exist and are callable
        assert hasattr(self.wrapper, '_create_web_search_tool')
        assert hasattr(self.wrapper, '_create_calculator_tool')
        assert hasattr(self.wrapper, '_create_text_processing_tool')
        assert hasattr(self.wrapper, '_create_file_operations_tool')
        assert hasattr(self.wrapper, '_create_api_calls_tool')
        
        assert callable(self.wrapper._create_web_search_tool)
        assert callable(self.wrapper._create_calculator_tool)
        assert callable(self.wrapper._create_text_processing_tool)
        assert callable(self.wrapper._create_file_operations_tool)
        assert callable(self.wrapper._create_api_calls_tool)


class TestToolErrorHandling:
    """Test error handling in tool functions"""

    def setup_method(self):
        """Setup test agent wrapper"""
        from main import LangChainAgentWrapper, AgentConfig
        
        config = AgentConfig(
            name="Error Test Agent",
            description="Test error handling",
            capabilities=["calculations"]
        )
        
        self.wrapper = LangChainAgentWrapper(config)

    def test_calculator_division_by_zero(self):
        """Test calculator handles division by zero"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_calculator_tool()
            calc_func = mock_tool.call_args[1]['func']
            
            result = calc_func("1/0")
            assert "Error:" in result
            assert "division by zero" in result.lower() or "zerodivisionerror" in result.lower()

    def test_calculator_syntax_error(self):
        """Test calculator handles syntax errors"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_calculator_tool()
            calc_func = mock_tool.call_args[1]['func']
            
            syntax_errors = [
                "2 +",
                "* 5",
                "((2+3)",
                "2 + + 3",
                "2..5"
            ]
            
            for expr in syntax_errors:
                result = calc_func(expr)
                assert "Error:" in result

    def test_calculator_name_error(self):
        """Test calculator handles undefined variables"""
        with patch('main.Tool') as mock_tool:
            self.wrapper._create_calculator_tool()
            calc_func = mock_tool.call_args[1]['func']
            
            result = calc_func("undefined_variable + 5")
            assert "Error:" in result
            
            result = calc_func("x * y")
            assert "Error:" in result
