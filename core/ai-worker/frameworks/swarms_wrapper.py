#!/usr/bin/env python3
"""
AgentOS AI Worker - Swarms Framework Wrapper
Week 3 Day 1-2 Implementation: Swarms Integration

This module provides integration with the Swarms framework for distributed
agent coordination and swarm intelligence capabilities.
"""

import os
import uuid
import time
import asyncio
import logging
from typing import Dict, Any, List, Optional, Union

from .base_wrapper import (
    BaseFrameworkWrapper, FrameworkType, AgentConfig,
    TaskRequest, TaskResponse, InitializationError, ExecutionError
)

# Swarms imports with fallback
try:
    from swarms import Agent, OpenAIChat, Task
    from swarms.structs import Flow
    SWARMS_AVAILABLE = True
except ImportError:
    SWARMS_AVAILABLE = False
    # Create dummy classes for when Swarms is not available
    class Agent:
        def __init__(self, *args, **kwargs):
            pass
    class OpenAIChat:
        def __init__(self, *args, **kwargs):
            pass
    class Task:
        def __init__(self, *args, **kwargs):
            pass
    class Flow:
        def __init__(self, *args, **kwargs):
            pass

logger = logging.getLogger(__name__)

class SwarmAgentWrapper(BaseFrameworkWrapper):
    """
    Swarms framework wrapper for AgentOS.

    Provides distributed agent coordination and swarm intelligence capabilities
    through the Swarms framework integration.
    """

    def __init__(self, agent_config: AgentConfig):
        super().__init__(agent_config)
        self.swarm_agent = None
        self.llm = None
        self.flow = None
        self.tasks = []

    def _get_framework_type(self) -> FrameworkType:
        """Return Swarms framework type"""
        return FrameworkType.SWARMS

    async def initialize(self) -> bool:
        """Initialize Swarms agent with capabilities"""
        try:
            if not SWARMS_AVAILABLE:
                raise InitializationError(
                    "Swarms framework not available. Install with: pip install swarms",
                    framework="swarms",
                    agent_id=self.agent_id
                )

            # Check for OpenAI API key
            if not os.getenv("OPENAI_API_KEY"):
                raise InitializationError(
                    "OpenAI API key not configured for Swarms",
                    framework="swarms",
                    agent_id=self.agent_id
                )

            # Initialize OpenAI LLM for Swarms
            self.llm = OpenAIChat(
                model_name=self.agent_config.model,
                temperature=self.agent_config.temperature,
                max_tokens=2000,
                openai_api_key=os.getenv("OPENAI_API_KEY")
            )

            # Create Swarms agent
            self.swarm_agent = Agent(
                agent_name=self.agent_config.name,
                agent_description=self.agent_config.description,
                llm=self.llm,
                max_loops=self.agent_config.max_iterations,
                autosave=True,
                verbose=True,
                dynamic_temperature_enabled=True,
                saved_state_path=f"agent_states/{self.agent_id}.json",
                user_name="agentos_user",
                retry_attempts=3,
                context_length=8000,
                return_step_meta=True
            )

            # Convert capabilities to tools
            await self._setup_capabilities()

            # Initialize flow for multi-agent coordination
            self.flow = Flow(
                agents=[self.swarm_agent],
                flow_type="sequential"  # Can be "sequential", "parallel", or "round_robin"
            )

            self.is_initialized = True
            logger.info(f"Swarms agent {self.agent_id} initialized successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to initialize Swarms agent {self.agent_id}: {str(e)}")
            raise InitializationError(
                f"Swarms initialization failed: {str(e)}",
                framework="swarms",
                agent_id=self.agent_id
            )

    async def _setup_capabilities(self):
        """Convert AgentOS capabilities to Swarms tools"""
        for capability in self.agent_config.capabilities:
            tool = await self._capability_to_tool(capability)
            if tool:
                self.tools.append(tool)
                # Add tool to agent (Swarms handles this differently)
                logger.info(f"Added capability '{capability}' to Swarms agent")

    async def _capability_to_tool(self, capability: str) -> Optional[Dict[str, Any]]:
        """Convert AgentOS capability to Swarms tool format"""
        tool_map = {
            "web_search": self._create_web_search_tool(),
            "calculations": self._create_calculator_tool(),
            "text_processing": self._create_text_processing_tool(),
            "file_operations": self._create_file_operations_tool(),
            "api_calls": self._create_api_calls_tool(),
        }
        return tool_map.get(capability)

    def _create_web_search_tool(self) -> Dict[str, Any]:
        """Create web search tool for Swarms"""
        def web_search(query: str) -> str:
            # Real DuckDuckGo search implementation for Swarms
            try:
                from duckduckgo_search import DDGS

                with DDGS() as ddgs:
                    results = list(ddgs.text(query, max_results=3))

                if results:
                    formatted_results = []
                    for result in results:
                        formatted_results.append(f"Title: {result.get('title', 'N/A')}\nURL: {result.get('href', 'N/A')}\nDescription: {result.get('body', 'N/A')}")

                    return f"Swarms web search results for '{query}':\n\n" + "\n\n".join(formatted_results)
                else:
                    return f"Swarms web search found no results for: {query}"

            except ImportError:
                return f"Swarms web search unavailable (DuckDuckGo package not installed) for: {query}"
            except Exception as e:
                return f"Swarms web search error for '{query}': {str(e)}"

        return {
            "name": "web_search",
            "description": "Search the web for information using Swarms",
            "function": web_search,
            "parameters": {
                "query": {"type": "string", "description": "Search query"}
            }
        }

    def _create_calculator_tool(self) -> Dict[str, Any]:
        """Create calculator tool for Swarms"""
        def calculate(expression: str) -> str:
            try:
                # Safe evaluation of mathematical expressions
                result = eval(expression, {"__builtins__": {}}, {})
                return f"Calculation result: {result}"
            except Exception as e:
                return f"Calculation error: {str(e)}"

        return {
            "name": "calculator",
            "description": "Perform mathematical calculations in Swarms",
            "function": calculate,
            "parameters": {
                "expression": {"type": "string", "description": "Mathematical expression"}
            }
        }

    def _create_text_processing_tool(self) -> Dict[str, Any]:
        """Create text processing tool for Swarms"""
        def process_text(text: str, operation: str = "analyze") -> str:
            # Basic text processing
            if operation == "analyze":
                return f"Swarms text analysis: {len(text)} characters, {len(text.split())} words"
            elif operation == "summarize":
                return f"Swarms summary: {text[:100]}..."
            else:
                return f"Swarms processed: {text.strip().lower()}"

        return {
            "name": "text_processor",
            "description": "Process and analyze text using Swarms",
            "function": process_text,
            "parameters": {
                "text": {"type": "string", "description": "Text to process"},
                "operation": {"type": "string", "description": "Processing operation"}
            }
        }

    def _create_file_operations_tool(self) -> Dict[str, Any]:
        """Create file operations tool for Swarms"""
        def file_operation(operation: str, file_path: str = "", content: str = "") -> str:
            # Real secure file operations for Swarms
            try:
                import os
                import tempfile

                # Security: Use secure temp directory
                secure_dir = os.path.join(tempfile.gettempdir(), "agentos_swarms_files")
                os.makedirs(secure_dir, exist_ok=True)

                # Security: Validate file path
                if ".." in file_path or file_path.startswith("/"):
                    return f"Swarms file operation error: Invalid file path for security"

                full_path = os.path.join(secure_dir, file_path)

                if operation == "read":
                    if os.path.exists(full_path):
                        with open(full_path, 'r', encoding='utf-8') as f:
                            file_content = f.read()
                        return f"Swarms file read successful: {len(file_content)} characters from {file_path}"
                    else:
                        return f"Swarms file operation: File {file_path} does not exist"

                elif operation == "write":
                    with open(full_path, 'w', encoding='utf-8') as f:
                        f.write(content)
                    return f"Swarms file write successful: {len(content)} characters to {file_path}"

                elif operation == "list":
                    if os.path.isdir(full_path):
                        files = os.listdir(full_path)
                        return f"Swarms directory listing for {file_path}: {', '.join(files)}"
                    else:
                        return f"Swarms file operation: {file_path} is not a directory"

                else:
                    return f"Swarms file operation: Unsupported operation {operation}"

            except Exception as e:
                return f"Swarms file operation error: {str(e)}"

        return {
            "name": "file_operations",
            "description": "Perform safe file operations using Swarms",
            "function": file_operation,
            "parameters": {
                "operation": {"type": "string", "description": "File operation type"},
                "file_path": {"type": "string", "description": "File path"},
                "content": {"type": "string", "description": "File content"}
            }
        }

    def _create_api_calls_tool(self) -> Dict[str, Any]:
        """Create API calls tool for Swarms"""
        def api_call(url: str, method: str = "GET", data: Optional[Dict] = None) -> str:
            # Real secure API calls for Swarms
            try:
                import requests
                import time

                # Security: Validate URL
                allowed_domains = [
                    "api.github.com",
                    "httpbin.org",
                    "jsonplaceholder.typicode.com",
                    "api.openai.com"
                ]

                if not any(domain in url for domain in allowed_domains):
                    return f"Swarms API call error: Domain not in whitelist"

                # Make real HTTP request
                start_time = time.time()

                if method.upper() == "GET":
                    response = requests.get(url, timeout=10)
                elif method.upper() == "POST":
                    response = requests.post(url, json=data, timeout=10)
                else:
                    return f"Swarms API call error: Unsupported method {method}"

                execution_time = time.time() - start_time

                return f"Swarms API call successful: {method} {url} -> Status: {response.status_code}, Time: {execution_time:.2f}s, Response length: {len(response.text)} chars"

            except ImportError:
                return f"Swarms API call unavailable: requests package not installed"
            except Exception as e:
                return f"Swarms API call error: {str(e)}"

        return {
            "name": "api_calls",
            "description": "Make HTTP API calls using Swarms",
            "function": api_call,
            "parameters": {
                "url": {"type": "string", "description": "API URL"},
                "method": {"type": "string", "description": "HTTP method"},
                "data": {"type": "object", "description": "Request data"}
            }
        }

    async def execute(self, task_request: TaskRequest) -> TaskResponse:
        """Execute task using Swarms agent"""
        if not self.is_initialized:
            await self.initialize()

        task_id = str(uuid.uuid4())
        start_time = time.time()

        try:
            # Create Swarms task
            swarm_task = Task(
                task=task_request.task,
                agent=self.swarm_agent,
                context=task_request.context or {}
            )

            # Execute with timeout
            timeout = task_request.timeout or self.agent_config.timeout
            result = await self._execute_with_timeout(
                self._run_swarm_task(swarm_task),
                timeout
            )

            execution_time = time.time() - start_time

            return self._create_task_response(
                task_id=task_id,
                result=result,
                status="completed",
                execution_time=execution_time,
                metadata={
                    "swarm_agent_id": self.swarm_agent.agent_name,
                    "tools_used": [tool["name"] for tool in self.tools],
                    "iterations": getattr(result, 'iterations', 1)
                }
            )

        except Exception as e:
            execution_time = time.time() - start_time
            logger.error(f"Swarms execution failed for task {task_id}: {str(e)}")

            return self._create_task_response(
                task_id=task_id,
                result=None,
                status="failed",
                execution_time=execution_time,
                error_message=str(e),
                metadata={"error_type": type(e).__name__}
            )

    async def _run_swarm_task(self, task: Task) -> str:
        """Run Swarms task asynchronously"""
        # Swarms run method - adapt based on actual Swarms API
        try:
            result = self.swarm_agent.run(task.task)
            return result
        except Exception as e:
            raise ExecutionError(
                f"Swarms task execution failed: {str(e)}",
                framework="swarms",
                agent_id=self.agent_id
            )

    async def cleanup(self) -> bool:
        """Clean up Swarms resources"""
        try:
            if self.swarm_agent:
                # Save agent state
                if hasattr(self.swarm_agent, 'save_state'):
                    self.swarm_agent.save_state()

            self.is_initialized = False
            logger.info(f"Swarms agent {self.agent_id} cleaned up successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to cleanup Swarms agent {self.agent_id}: {str(e)}")
            return False

    def _get_memory_usage(self) -> Dict[str, Any]:
        """Get Swarms-specific memory usage"""
        return {
            "working_memory": len(getattr(self.swarm_agent, 'memory', [])) if self.swarm_agent else 0,
            "episodic_memory": 0,  # Swarms doesn't have explicit episodic memory
            "semantic_memory": 0,  # Swarms doesn't have explicit semantic memory
            "agent_state_size": self._get_agent_state_size()
        }

    def _get_agent_state_size(self) -> int:
        """Get size of agent state in bytes"""
        try:
            if self.swarm_agent and hasattr(self.swarm_agent, 'saved_state_path'):
                import os
                if os.path.exists(self.swarm_agent.saved_state_path):
                    return os.path.getsize(self.swarm_agent.saved_state_path)
            return 0
        except:
            return 0
