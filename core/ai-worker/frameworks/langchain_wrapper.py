#!/usr/bin/env python3
"""
AgentOS AI Worker - LangChain Framework Wrapper
Week 3 Implementation: Refactored LangChain Integration

This module provides integration with LangChain framework, refactored from
the existing Week 2 implementation into the new multi-framework architecture.
"""

import os
import uuid
import time
import asyncio
import logging
import ast
import operator
import math
import json
import requests
from pathlib import Path
from urllib.parse import urlparse
from typing import Dict, Any, List, Optional

from .base_wrapper import (
    BaseFrameworkWrapper, FrameworkType, AgentConfig,
    TaskRequest, TaskResponse, InitializationError, ExecutionError
)

# Real tool implementations
try:
    from duckduckgo_search import DDGS
    DUCKDUCKGO_AVAILABLE = True
except ImportError:
    DUCKDUCKGO_AVAILABLE = False

# LangChain imports with fallback
try:
    from langchain.agents import initialize_agent, AgentType
    from langchain.llms import OpenAI
    from langchain.memory import ConversationBufferMemory
    from langchain.tools import Tool
    from langchain.schema import AgentAction, AgentFinish
    LANGCHAIN_AVAILABLE = True
except ImportError:
    LANGCHAIN_AVAILABLE = False
    # Create dummy classes for when LangChain is not available
    class Tool:
        def __init__(self, name, description, func):
            self.name = name
            self.description = description
            self.func = func
    class AgentType:
        CONVERSATIONAL_REACT_DESCRIPTION = "conversational-react-description"

logger = logging.getLogger(__name__)

class LangChainAgentWrapper(BaseFrameworkWrapper):
    """
    LangChain framework wrapper for AgentOS.

    Refactored from Week 2 implementation to fit the new multi-framework
    architecture while maintaining all existing functionality.
    """

    def __init__(self, agent_config: AgentConfig):
        super().__init__(agent_config)
        self.langchain_agent = None
        self.llm = None
        self.memory = None

    def _get_framework_type(self) -> FrameworkType:
        """Return LangChain framework type"""
        return FrameworkType.LANGCHAIN

    async def initialize(self) -> bool:
        """Initialize LangChain agent with capabilities"""
        try:
            if not LANGCHAIN_AVAILABLE:
                raise InitializationError(
                    "LangChain framework not available. Install with: pip install langchain",
                    framework="langchain",
                    agent_id=self.agent_id
                )

            # Check for OpenAI API key
            if not os.getenv("OPENAI_API_KEY"):
                raise InitializationError(
                    "OpenAI API key not configured for LangChain",
                    framework="langchain",
                    agent_id=self.agent_id
                )

            # Initialize LLM
            self.llm = OpenAI(
                temperature=self.agent_config.temperature,
                model_name=self.agent_config.model,
                openai_api_key=os.getenv("OPENAI_API_KEY")
            )

            # Initialize memory
            self.memory = ConversationBufferMemory(memory_key="chat_history")

            # Convert capabilities to tools
            await self._setup_capabilities()

            # Initialize LangChain agent
            if self.tools:
                self.langchain_agent = initialize_agent(
                    tools=self.tools,
                    llm=self.llm,
                    agent=AgentType.CONVERSATIONAL_REACT_DESCRIPTION,
                    memory=self.memory,
                    verbose=True,
                    max_iterations=self.agent_config.max_iterations,
                    return_intermediate_steps=True
                )

            self.is_initialized = True
            logger.info(f"LangChain agent {self.agent_id} initialized successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to initialize LangChain agent {self.agent_id}: {str(e)}")
            raise InitializationError(
                f"LangChain initialization failed: {str(e)}",
                framework="langchain",
                agent_id=self.agent_id
            )

    async def _setup_capabilities(self):
        """Convert AgentOS capabilities to LangChain tools"""
        for capability in self.agent_config.capabilities:
            tool = await self._capability_to_tool(capability)
            if tool:
                self.tools.append(tool)
                logger.info(f"Added capability '{capability}' to LangChain agent")

    async def _capability_to_tool(self, capability: str) -> Optional[Tool]:
        """Convert AgentOS capability to LangChain tool"""
        tool_map = {
            "web_search": self._create_web_search_tool(),
            "calculations": self._create_calculator_tool(),
            "text_processing": self._create_text_processing_tool(),
            "file_operations": self._create_file_operations_tool(),
            "api_calls": self._create_api_calls_tool(),
        }
        return tool_map.get(capability)

    def _create_web_search_tool(self) -> Tool:
        """Create real web search tool using DuckDuckGo"""
        def web_search_real(query: str) -> str:
            """Real web search using DuckDuckGo API"""
            try:
                if not DUCKDUCKGO_AVAILABLE:
                    return f"DuckDuckGo search not available. Query was: {query}"

                with DDGS() as ddgs:
                    results = list(ddgs.text(query, max_results=5))

                if not results:
                    return f"No search results found for: {query}"

                formatted_results = []
                for i, result in enumerate(results, 1):
                    formatted_results.append(
                        f"{i}. {result['title']}\n"
                        f"   URL: {result['href']}\n"
                        f"   Summary: {result['body'][:200]}...\n"
                    )

                return f"Web search results for '{query}':\n\n" + "\n".join(formatted_results)

            except Exception as e:
                return f"Search error for '{query}': {str(e)}"

        return Tool(
            name="web_search",
            description="Search the web for current information using DuckDuckGo",
            func=web_search_real
        )

    def _create_calculator_tool(self) -> Tool:
        """Create enhanced calculator tool with safe evaluation"""
        def calculate_real(expression: str) -> str:
            """Real calculator with safe evaluation and math functions"""
            try:
                # Safe mathematical operations
                allowed_operators = {
                    ast.Add: operator.add,
                    ast.Sub: operator.sub,
                    ast.Mult: operator.mul,
                    ast.Div: operator.truediv,
                    ast.Pow: operator.pow,
                    ast.USub: operator.neg,
                    ast.UAdd: operator.pos,
                }

                # Safe mathematical functions
                allowed_functions = {
                    'sin': math.sin,
                    'cos': math.cos,
                    'tan': math.tan,
                    'sqrt': math.sqrt,
                    'log': math.log,
                    'log10': math.log10,
                    'exp': math.exp,
                    'abs': abs,
                    'round': round,
                    'floor': math.floor,
                    'ceil': math.ceil,
                    'pi': math.pi,
                    'e': math.e,
                }

                # Parse and evaluate safely
                result = self._safe_eval(expression, allowed_operators, allowed_functions)
                return f"Calculation result: {result}"

            except Exception as e:
                return f"Calculation error: {str(e)}"

        return Tool(
            name="calculator",
            description="Perform mathematical calculations with support for basic operations and math functions (sin, cos, sqrt, log, etc.)",
            func=calculate_real
        )

    def _create_text_processing_tool(self) -> Tool:
        """Create text processing tool for LangChain"""
        def process_text(text: str) -> str:
            # Basic text processing
            return f"LangChain processed: {text.strip().lower()}"

        return Tool(
            name="text_processor",
            description="Process and analyze text using LangChain",
            func=process_text
        )

    def _create_file_operations_tool(self) -> Tool:
        """Create real file operations tool with security constraints"""
        def file_operation_real(operation_json: str) -> str:
            """Real file operations with security sandboxing"""
            try:
                # Parse operation JSON
                operation_data = json.loads(operation_json)
                operation = operation_data.get('operation', '')
                path = operation_data.get('path', '')
                content = operation_data.get('content', '')

                # Security: Only allow operations in safe directory
                safe_dir = Path("/tmp/agentos_files")
                safe_dir.mkdir(exist_ok=True)

                # Prevent directory traversal
                if path:
                    file_path = safe_dir / Path(path).name
                else:
                    return "Error: No file path specified"

                if operation == "read":
                    if file_path.exists() and file_path.is_file():
                        content = file_path.read_text(encoding='utf-8')
                        return f"File content of '{file_path.name}':\n{content}"
                    else:
                        return f"File not found: {file_path.name}"

                elif operation == "write":
                    if not content:
                        return "Error: No content specified for write operation"
                    file_path.write_text(content, encoding='utf-8')
                    return f"File written successfully: {file_path.name}"

                elif operation == "list":
                    files = [f.name for f in safe_dir.iterdir() if f.is_file()]
                    return f"Files in directory: {', '.join(files) if files else 'No files found'}"

                elif operation == "delete":
                    if file_path.exists():
                        file_path.unlink()
                        return f"File deleted successfully: {file_path.name}"
                    else:
                        return f"File not found: {file_path.name}"

                else:
                    return f"Unsupported operation: {operation}. Supported: read, write, list, delete"

            except json.JSONDecodeError:
                return "Error: Invalid JSON format. Use: {\"operation\": \"read/write/list/delete\", \"path\": \"filename\", \"content\": \"text\"}"
            except Exception as e:
                return f"File operation error: {str(e)}"

        return Tool(
            name="file_operations",
            description="Perform secure file operations (read, write, list, delete) in sandboxed directory. Use JSON format: {\"operation\": \"read\", \"path\": \"filename\", \"content\": \"text\"}",
            func=file_operation_real
        )

    def _create_api_calls_tool(self) -> Tool:
        """Create real API calls tool with security constraints"""
        def api_call_real(request_json: str) -> str:
            """Real API calls with security constraints"""
            try:
                # Parse request JSON
                request_data = json.loads(request_json)
                url = request_data.get('url', '')
                method = request_data.get('method', 'GET').upper()
                headers = request_data.get('headers', {})
                data = request_data.get('data', None)

                if not url:
                    return "Error: No URL specified"

                # Security: Only allow calls to approved domains
                approved_domains = [
                    "api.github.com",
                    "jsonplaceholder.typicode.com",
                    "httpbin.org",
                    "api.openweathermap.org",
                    "api.exchangerate-api.com",
                    "restcountries.com"
                ]

                domain = urlparse(url).netloc
                if domain not in approved_domains:
                    return f"Domain not approved: {domain}. Approved domains: {', '.join(approved_domains)}"

                # Make the API call
                response = requests.request(
                    method=method,
                    url=url,
                    headers=headers,
                    json=data if data else None,
                    timeout=10
                )

                # Format response
                result = f"API Response ({response.status_code}):\n"
                result += f"URL: {url}\n"
                result += f"Method: {method}\n"

                # Limit response size
                response_text = response.text[:1000]
                if len(response.text) > 1000:
                    response_text += "... (truncated)"

                result += f"Response: {response_text}"
                return result

            except json.JSONDecodeError:
                return "Error: Invalid JSON format. Use: {\"url\": \"https://api.example.com\", \"method\": \"GET\", \"headers\": {}, \"data\": {}}"
            except requests.RequestException as e:
                return f"API call error: {str(e)}"
            except Exception as e:
                return f"Unexpected error: {str(e)}"

        return Tool(
            name="api_calls",
            description="Make HTTP API calls to approved domains. Use JSON format: {\"url\": \"https://api.example.com\", \"method\": \"GET\", \"headers\": {}, \"data\": {}}",
            func=api_call_real
        )

    async def execute(self, task_request: TaskRequest) -> TaskResponse:
        """Execute task using LangChain agent"""
        if not self.is_initialized:
            await self.initialize()

        task_id = str(uuid.uuid4())
        start_time = time.time()

        try:
            if not self.langchain_agent:
                raise ExecutionError(
                    "LangChain agent not properly initialized",
                    framework="langchain",
                    agent_id=self.agent_id
                )

            # Execute with timeout
            timeout = task_request.timeout or self.agent_config.timeout
            result = await self._execute_with_timeout(
                self._run_langchain_task(task_request.task),
                timeout
            )

            execution_time = time.time() - start_time

            return self._create_task_response(
                task_id=task_id,
                result=result,
                status="completed",
                execution_time=execution_time,
                metadata={
                    "langchain_agent_type": "conversational-react-description",
                    "tools_used": [tool.name for tool in self.tools],
                    "memory_length": len(self.memory.chat_memory.messages) if self.memory else 0
                }
            )

        except Exception as e:
            execution_time = time.time() - start_time
            logger.error(f"LangChain execution failed for task {task_id}: {str(e)}")

            return self._create_task_response(
                task_id=task_id,
                result=None,
                status="failed",
                execution_time=execution_time,
                error_message=str(e),
                metadata={"error_type": type(e).__name__}
            )

    async def _run_langchain_task(self, task: str) -> str:
        """Run LangChain task asynchronously"""
        try:
            # Run in thread pool to avoid blocking
            loop = asyncio.get_event_loop()
            result = await loop.run_in_executor(
                None,
                self.langchain_agent.run,
                task
            )
            return result
        except Exception as e:
            raise ExecutionError(
                f"LangChain task execution failed: {str(e)}",
                framework="langchain",
                agent_id=self.agent_id
            )

    async def cleanup(self) -> bool:
        """Clean up LangChain resources"""
        try:
            if self.memory:
                self.memory.clear()

            self.is_initialized = False
            logger.info(f"LangChain agent {self.agent_id} cleaned up successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to cleanup LangChain agent {self.agent_id}: {str(e)}")
            return False

    def _get_memory_usage(self) -> Dict[str, Any]:
        """Get LangChain-specific memory usage"""
        memory_messages = 0
        if self.memory and hasattr(self.memory, 'chat_memory'):
            memory_messages = len(self.memory.chat_memory.messages)

        return {
            "working_memory": memory_messages,
            "episodic_memory": 0,  # LangChain doesn't have explicit episodic memory
            "semantic_memory": 0,  # LangChain doesn't have explicit semantic memory
            "conversation_length": memory_messages
        }

    def _safe_eval(self, expression: str, allowed_operators: dict, allowed_functions: dict):
        """Safely evaluate mathematical expressions"""
        try:
            # Parse the expression
            tree = ast.parse(expression, mode='eval')

            # Evaluate the AST
            return self._eval_node(tree.body, allowed_operators, allowed_functions)

        except Exception as e:
            raise ValueError(f"Invalid expression: {str(e)}")

    def _eval_node(self, node, allowed_operators: dict, allowed_functions: dict):
        """Recursively evaluate AST nodes"""
        if isinstance(node, ast.Constant):  # Python 3.8+
            return node.value
        elif isinstance(node, ast.Num):  # Python < 3.8
            return node.n
        elif isinstance(node, ast.Name):
            if node.id in allowed_functions:
                return allowed_functions[node.id]
            else:
                raise ValueError(f"Name '{node.id}' not allowed")
        elif isinstance(node, ast.BinOp):
            left = self._eval_node(node.left, allowed_operators, allowed_functions)
            right = self._eval_node(node.right, allowed_operators, allowed_functions)
            op_type = type(node.op)
            if op_type in allowed_operators:
                return allowed_operators[op_type](left, right)
            else:
                raise ValueError(f"Operator {op_type.__name__} not allowed")
        elif isinstance(node, ast.UnaryOp):
            operand = self._eval_node(node.operand, allowed_operators, allowed_functions)
            op_type = type(node.op)
            if op_type in allowed_operators:
                return allowed_operators[op_type](operand)
            else:
                raise ValueError(f"Unary operator {op_type.__name__} not allowed")
        elif isinstance(node, ast.Call):
            func = self._eval_node(node.func, allowed_operators, allowed_functions)
            args = [self._eval_node(arg, allowed_operators, allowed_functions) for arg in node.args]
            if callable(func):
                return func(*args)
            else:
                raise ValueError(f"Function call not allowed")
        else:
            raise ValueError(f"Node type {type(node).__name__} not allowed")
