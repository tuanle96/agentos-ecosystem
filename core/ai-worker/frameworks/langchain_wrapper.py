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
from typing import Dict, Any, List, Optional

from .base_wrapper import (
    BaseFrameworkWrapper, FrameworkType, AgentConfig, 
    TaskRequest, TaskResponse, InitializationError, ExecutionError
)

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
        """Create web search tool for LangChain"""
        def web_search(query: str) -> str:
            # Placeholder implementation - integrate with actual search API
            return f"LangChain web search results for: {query}"
        
        return Tool(
            name="web_search",
            description="Search the web for information using LangChain",
            func=web_search
        )
    
    def _create_calculator_tool(self) -> Tool:
        """Create calculator tool for LangChain"""
        def calculate(expression: str) -> str:
            try:
                # Safe evaluation of mathematical expressions
                result = eval(expression, {"__builtins__": {}}, {})
                return f"Calculation result: {result}"
            except Exception as e:
                return f"Calculation error: {str(e)}"
        
        return Tool(
            name="calculator",
            description="Perform mathematical calculations using LangChain",
            func=calculate
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
        """Create file operations tool for LangChain"""
        def file_operation(operation: str) -> str:
            # Placeholder for secure file operations
            return f"LangChain file operation: {operation}"
        
        return Tool(
            name="file_operations",
            description="Perform safe file operations using LangChain",
            func=file_operation
        )
    
    def _create_api_calls_tool(self) -> Tool:
        """Create API calls tool for LangChain"""
        def api_call(url: str) -> str:
            # Placeholder for secure API calls
            return f"LangChain API call to: {url}"
        
        return Tool(
            name="api_calls",
            description="Make HTTP API calls using LangChain",
            func=api_call
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
