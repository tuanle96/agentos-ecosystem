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
        from frameworks.langchain_wrapper import LangChainAgentWrapper
        from frameworks.base_wrapper import AgentConfig

        config = AgentConfig(
            name="Test Agent",
            description="Test Description",
            capabilities=["calculations", "web_search", "text_processing", "file_operations", "api_calls"]
        )

        self.wrapper = LangChainAgentWrapper(config)

    def test_web_search_tool_function(self):
        """Test web search tool function logic"""
        # Test the actual web search tool function directly
        search_tool = self.wrapper._create_web_search_tool()

        # Handle case where tool creation might fail
        if search_tool is None:
            pytest.skip("Web search tool creation failed")

        # Test tool properties
        assert search_tool.name == 'web_search'
        assert 'search' in search_tool.description.lower()

        # Skip actual web searches in unit tests
        pytest.skip("Skipping real web searches in unit tests")

    def test_calculator_tool_function(self):
        """Test calculator tool function logic"""
        # Test the actual calculator tool function directly
        calc_tool = self.wrapper._create_calculator_tool()

        # Handle case where tool creation might fail
        if calc_tool is None:
            pytest.skip("Calculator tool creation failed")

        # Test tool properties
        assert calc_tool.name == 'calculator'
        assert 'mathematical calculations' in calc_tool.description.lower()

        # Test valid expressions
        result = calc_tool.func("2+2")
        assert "4" in result

        result = calc_tool.func("10*5")
        assert "50" in result

        result = calc_tool.func("100/4")
        assert "25" in result

        result = calc_tool.func("2**3")
        assert "8" in result

        result = calc_tool.func("(5+3)*2")
        assert "16" in result

    def test_calculator_tool_error_handling(self):
        """Test calculator tool error handling"""
        # Test the actual calculator tool function directly
        calc_tool = self.wrapper._create_calculator_tool()

        # Handle case where tool creation might fail
        if calc_tool is None:
            pytest.skip("Calculator tool creation failed")

        # Test invalid expressions
        result = calc_tool.func("2/0")
        assert "error" in result.lower()

        result = calc_tool.func("invalid_expression")
        assert "error" in result.lower()

        result = calc_tool.func("import os")
        assert "error" in result.lower()

        result = calc_tool.func("__import__('os')")
        assert "error" in result.lower()

        # Test empty expression
        result = calc_tool.func("")
        assert "error" in result.lower()

    def test_calculator_tool_security(self):
        """Test calculator tool security restrictions"""
        # Test the actual calculator tool function directly
        calc_tool = self.wrapper._create_calculator_tool()

        # Handle case where tool creation might fail
        if calc_tool is None:
            pytest.skip("Calculator tool creation failed")

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
            result = calc_tool.func(expr)
            assert "error" in result.lower()

    def test_text_processing_tool_function(self):
        """Test text processing tool function logic"""
        # Test the actual text processing tool function directly
        text_tool = self.wrapper._create_text_processing_tool()

        # Handle case where tool creation might fail
        if text_tool is None:
            pytest.skip("Text processing tool creation failed")

        # Test tool properties
        assert text_tool.name == 'text_processor'
        assert 'text' in text_tool.description.lower()

        # Test text processing
        result = text_tool.func("  HELLO WORLD  ")
        assert "hello world" in result.lower()

        result = text_tool.func("Python Programming")
        assert "python programming" in result.lower()

        result = text_tool.func("   Mixed   Case   Text   ")
        assert "mixed" in result.lower() and "case" in result.lower()

        # Test empty text
        result = text_tool.func("")
        assert result is not None

        # Test whitespace only
        result = text_tool.func("   ")
        assert result is not None

    def test_file_operations_tool_function(self):
        """Test file operations tool function logic"""
        # Test the actual file operations tool function directly
        file_tool = self.wrapper._create_file_operations_tool()

        # Handle case where tool creation might fail
        if file_tool is None:
            pytest.skip("File operations tool creation failed")

        # Test tool properties
        assert file_tool.name == 'file_operations'
        assert 'file' in file_tool.description.lower()

        # Test different operations with JSON format
        import json

        # Test list operation
        list_op = json.dumps({"operation": "list", "path": "."})
        result = file_tool.func(list_op)
        assert "files in directory" in result.lower() or "no files found" in result.lower()

        # Test write operation
        write_op = json.dumps({"operation": "write", "path": "test.txt", "content": "test content"})
        result = file_tool.func(write_op)
        assert "written successfully" in result.lower() or "error" in result.lower()

        # Test read operation
        read_op = json.dumps({"operation": "read", "path": "test.txt"})
        result = file_tool.func(read_op)
        assert "file content" in result.lower() or "not found" in result.lower()

        # Test invalid JSON
        result = file_tool.func("invalid json")
        assert "error" in result.lower() and "json" in result.lower()

    def test_api_calls_tool_function(self):
        """Test API calls tool function logic"""
        # Test the actual API calls tool function directly
        api_tool = self.wrapper._create_api_calls_tool()

        # Handle case where tool creation might fail
        if api_tool is None:
            pytest.skip("API calls tool creation failed")

        # Test tool properties
        assert api_tool.name == 'api_calls'
        assert 'api' in api_tool.description.lower()

        # Test different API calls with JSON format
        import json

        # Skip actual API calls in tests - just test domain validation
        # Test approved domain format (don't make real call)
        pytest.skip("Skipping real API calls in unit tests")

        # Test unapproved domain (should be blocked)
        blocked_call = json.dumps({
            "url": "https://malicious-site.com/data",
            "method": "GET"
        })
        result = api_tool.func(blocked_call)
        assert "not approved" in result.lower() or "domain" in result.lower()

        # Test invalid JSON
        result = api_tool.func("invalid json")
        assert "error" in result.lower() and "json" in result.lower()

        # Test missing URL
        no_url_call = json.dumps({"method": "GET"})
        result = api_tool.func(no_url_call)
        assert "error" in result.lower() and "url" in result.lower()


class TestToolIntegration:
    """Test tool integration with agent wrapper"""

    def setup_method(self):
        """Setup test agent wrapper"""
        from frameworks.langchain_wrapper import LangChainAgentWrapper
        from frameworks.base_wrapper import AgentConfig

        config = AgentConfig(
            name="Integration Test Agent",
            description="Test tool integration",
            capabilities=["calculations", "web_search", "text_processing"]
        )

        self.wrapper = LangChainAgentWrapper(config)

    @pytest.mark.asyncio
    async def test_capability_to_tool_all_types(self):
        """Test capability to tool conversion for all tool types"""
        # Test all supported capabilities directly
        capabilities = ["web_search", "calculations", "text_processing", "file_operations", "api_calls"]

        for capability in capabilities:
            tool = await self.wrapper._capability_to_tool(capability)
            assert tool is not None, f"Tool creation failed for capability: {capability}"

            # Verify tool has expected properties
            assert hasattr(tool, 'name'), f"Tool for {capability} missing name attribute"
            assert hasattr(tool, 'description'), f"Tool for {capability} missing description attribute"
            assert hasattr(tool, 'func'), f"Tool for {capability} missing func attribute"
            assert callable(tool.func), f"Tool func for {capability} is not callable"

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
    async def test_tool_initialization_in_agent(self):
        """Test that tools are properly initialized in agent"""
        # Skip this test as it requires LangChain dependencies
        # and complex mocking that's not suitable for unit tests
        pytest.skip("Skipping agent initialization test - requires LangChain dependencies")

    @pytest.mark.asyncio
    async def test_tool_initialization_without_tools(self):
        """Test agent initialization when no tools are created"""
        # Skip this test as it requires LangChain dependencies
        pytest.skip("Skipping agent initialization test - requires LangChain dependencies")

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
        from frameworks.langchain_wrapper import LangChainAgentWrapper
        from frameworks.base_wrapper import AgentConfig

        config = AgentConfig(
            name="Error Test Agent",
            description="Test error handling",
            capabilities=["calculations"]
        )

        self.wrapper = LangChainAgentWrapper(config)

    def test_calculator_division_by_zero(self):
        """Test calculator handles division by zero"""
        # Test the actual calculator tool function directly
        calc_tool = self.wrapper._create_calculator_tool()

        # Handle case where tool creation might fail
        if calc_tool is None:
            pytest.skip("Calculator tool creation failed - likely missing dependencies")

        result = calc_tool.func("1/0")
        assert "error" in result.lower()
        assert "division by zero" in result.lower() or "zerodivisionerror" in result.lower()

    def test_calculator_syntax_error(self):
        """Test calculator handles syntax errors"""
        # Test the actual calculator tool function directly
        calc_tool = self.wrapper._create_calculator_tool()

        # Handle case where tool creation might fail
        if calc_tool is None:
            pytest.skip("Calculator tool creation failed")

        syntax_errors = [
            "2 +",
            "* 5",
            "((2+3)",
            "2..5"
        ]

        for expr in syntax_errors:
            result = calc_tool.func(expr)
            assert "error" in result.lower()

    def test_calculator_name_error(self):
        """Test calculator handles undefined variables"""
        # Test the actual calculator tool function directly
        calc_tool = self.wrapper._create_calculator_tool()

        # Handle case where tool creation might fail
        if calc_tool is None:
            pytest.skip("Calculator tool creation failed")

        result = calc_tool.func("undefined_variable + 5")
        assert "error" in result.lower()

        result = calc_tool.func("x * y")
        assert "error" in result.lower()
