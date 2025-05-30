#!/usr/bin/env python3
"""
AgentOS AI Worker - AutoGen Framework Wrapper
Week 3 Day 5-6 Implementation: AutoGen Integration

This module provides integration with the AutoGen framework for conversational
AI and code generation capabilities.
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

# AutoGen imports with fallback
try:
    from autogen import AssistantAgent, UserProxyAgent, GroupChat, GroupChatManager
    from autogen.coding import LocalCommandLineCodeExecutor
    AUTOGEN_AVAILABLE = True
except ImportError:
    AUTOGEN_AVAILABLE = False
    # Create dummy classes for when AutoGen is not available
    class AssistantAgent:
        def __init__(self, *args, **kwargs):
            pass
    class UserProxyAgent:
        def __init__(self, *args, **kwargs):
            pass
    class GroupChat:
        def __init__(self, *args, **kwargs):
            pass
    class GroupChatManager:
        def __init__(self, *args, **kwargs):
            pass
    class LocalCommandLineCodeExecutor:
        def __init__(self, *args, **kwargs):
            pass

logger = logging.getLogger(__name__)

class AutoGenAgentWrapper(BaseFrameworkWrapper):
    """
    AutoGen framework wrapper for AgentOS.
    
    Provides conversational AI and code generation capabilities
    through the AutoGen framework integration.
    """
    
    def __init__(self, agent_config: AgentConfig):
        super().__init__(agent_config)
        self.assistant_agent = None
        self.user_proxy = None
        self.group_chat = None
        self.group_chat_manager = None
        self.code_executor = None
        self.llm_config = None
        self.conversation_history = []
        
    def _get_framework_type(self) -> FrameworkType:
        """Return AutoGen framework type"""
        return FrameworkType.AUTOGEN
    
    async def initialize(self) -> bool:
        """Initialize AutoGen agent with capabilities"""
        try:
            if not AUTOGEN_AVAILABLE:
                raise InitializationError(
                    "AutoGen framework not available. Install with: pip install pyautogen",
                    framework="autogen",
                    agent_id=self.agent_id
                )
            
            # For now, we'll create a mock implementation since AutoGen is not installed
            # This allows the framework to be tested without requiring all dependencies
            
            self.llm_config = {
                "model": self.agent_config.model,
                "temperature": self.agent_config.temperature,
                "timeout": self.agent_config.timeout,
            }
            
            # Mock assistant agent
            self.assistant_agent = {
                "name": self.agent_config.name,
                "system_message": self._create_system_message(),
                "llm_config": self.llm_config,
                "max_consecutive_auto_reply": self.agent_config.max_iterations,
            }
            
            # Mock user proxy agent
            self.user_proxy = {
                "name": "user_proxy",
                "human_input_mode": "NEVER",
                "max_consecutive_auto_reply": 0,
                "code_execution_config": self._create_code_execution_config(),
            }
            
            # Setup capabilities
            await self._setup_capabilities()
            
            # Mock group chat
            self.group_chat = {
                "agents": [self.assistant_agent, self.user_proxy],
                "messages": [],
                "max_round": self.agent_config.max_iterations * 2,
            }
            
            self.is_initialized = True
            logger.info(f"AutoGen agent {self.agent_id} initialized successfully")
            return True
            
        except Exception as e:
            logger.error(f"Failed to initialize AutoGen agent {self.agent_id}: {str(e)}")
            raise InitializationError(
                f"AutoGen initialization failed: {str(e)}",
                framework="autogen",
                agent_id=self.agent_id
            )
    
    def _create_system_message(self) -> str:
        """Create system message based on agent configuration"""
        capabilities_text = ", ".join(self.agent_config.capabilities)
        
        base_message = f"""You are {self.agent_config.name}, {self.agent_config.description}.

Your capabilities include: {capabilities_text}

You are helpful, accurate, and efficient. When asked to write code, provide clean, 
well-commented code. When asked to analyze data, provide thorough analysis.
When asked questions, provide comprehensive and accurate answers.

Always strive to complete tasks effectively and provide value to the user."""
        
        # Add personality traits if specified
        if self.agent_config.personality:
            personality_text = ", ".join([f"{k}: {v}" for k, v in self.agent_config.personality.items()])
            base_message += f"\n\nPersonality traits: {personality_text}"
        
        return base_message
    
    def _create_code_execution_config(self) -> Dict[str, Any]:
        """Create code execution configuration"""
        if "code_generation" in self.agent_config.capabilities:
            return {
                "executor": "local",
                "timeout": 60,
                "work_dir": "./autogen_workspace"
            }
        else:
            return {"executor": False}
    
    async def _setup_capabilities(self):
        """Setup AutoGen-specific capabilities"""
        for capability in self.agent_config.capabilities:
            tool_function = await self._capability_to_tool_function(capability)
            if tool_function:
                self._register_tool_function(capability, tool_function)
                logger.info(f"Added capability '{capability}' to AutoGen agent")
    
    async def _capability_to_tool_function(self, capability: str) -> Optional[callable]:
        """Convert AgentOS capability to AutoGen tool function"""
        tool_map = {
            "web_search": lambda query: f"AutoGen web search results for: {query}",
            "calculations": lambda expr: f"AutoGen calculation result: {eval(expr, {'__builtins__': {}}, {})}",
            "text_processing": lambda text: f"AutoGen text analysis: {len(text)} characters, {len(text.split())} words",
            "file_operations": lambda op, path="": f"AutoGen file operation: {op} on {path}",
            "api_calls": lambda url, method="GET": f"AutoGen API call: {method} {url}",
            "code_generation": lambda task: f"AutoGen code generation for: {task}"
        }
        return tool_map.get(capability)
    
    def _register_tool_function(self, capability: str, tool_function: callable):
        """Register tool function with AutoGen agent"""
        if not hasattr(self, '_tool_functions'):
            self._tool_functions = {}
        self._tool_functions[capability] = tool_function
    
    async def execute(self, task_request: TaskRequest) -> TaskResponse:
        """Execute task using AutoGen agent"""
        if not self.is_initialized:
            await self.initialize()
        
        task_id = str(uuid.uuid4())
        start_time = time.time()
        
        try:
            # Determine if this is a conversational or code generation task
            is_code_task = self._is_code_generation_task(task_request.task)
            
            # Mock AutoGen execution for testing
            if is_code_task:
                result = f"AutoGen code generation completed for: {task_request.task}\n\n# Generated code would appear here"
            else:
                result = f"AutoGen conversation completed: {task_request.task}"
            
            # Add to conversation history
            self.conversation_history.append({
                "task": task_request.task,
                "result": result,
                "timestamp": time.time(),
                "task_type": "code_generation" if is_code_task else "conversation"
            })
            
            execution_time = time.time() - start_time
            
            return self._create_task_response(
                task_id=task_id,
                result=result,
                status="completed",
                execution_time=execution_time,
                metadata={
                    "autogen_mode": "code_generation" if is_code_task else "conversation",
                    "assistant_name": self.assistant_agent["name"],
                    "conversation_rounds": len(self.conversation_history),
                    "tools_available": list(getattr(self, '_tool_functions', {}).keys())
                }
            )
            
        except Exception as e:
            execution_time = time.time() - start_time
            logger.error(f"AutoGen execution failed for task {task_id}: {str(e)}")
            
            return self._create_task_response(
                task_id=task_id,
                result=None,
                status="failed",
                execution_time=execution_time,
                error_message=str(e),
                metadata={"error_type": type(e).__name__}
            )
    
    def _is_code_generation_task(self, task: str) -> bool:
        """Determine if task requires code generation"""
        code_keywords = [
            "code", "program", "script", "function", "class", "implement",
            "write code", "create function", "build script", "develop",
            "programming", "algorithm", "debug", "fix code"
        ]
        task_lower = task.lower()
        return any(keyword in task_lower for keyword in code_keywords)
    
    async def cleanup(self) -> bool:
        """Clean up AutoGen resources"""
        try:
            # Clear conversation history
            if self.group_chat:
                self.group_chat["messages"] = []
            
            self.conversation_history = []
            self.is_initialized = False
            logger.info(f"AutoGen agent {self.agent_id} cleaned up successfully")
            return True
            
        except Exception as e:
            logger.error(f"Failed to cleanup AutoGen agent {self.agent_id}: {str(e)}")
            return False
    
    def _get_memory_usage(self) -> Dict[str, Any]:
        """Get AutoGen-specific memory usage"""
        conversation_length = len(self.conversation_history)
        
        return {
            "working_memory": conversation_length,
            "episodic_memory": 0,
            "semantic_memory": 0,
            "conversation_history": conversation_length,
            "tools_registered": len(getattr(self, '_tool_functions', {}))
        }
