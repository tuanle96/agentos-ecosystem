#!/usr/bin/env python3
"""
AgentOS AI Worker - Base Framework Wrapper
Week 3 Implementation: Universal Framework Interface

This module provides the base abstract class for all framework wrappers,
ensuring consistent interface across LangChain, Swarms, CrewAI, and AutoGen.
"""

import uuid
import time
import asyncio
from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional, Union
from pydantic import BaseModel, field_validator
from enum import Enum

class FrameworkType(str, Enum):
    """Supported AI frameworks"""
    LANGCHAIN = "langchain"
    SWARMS = "swarms"
    CREWAI = "crewai"
    AUTOGEN = "autogen"

class AgentConfig(BaseModel):
    """Universal agent configuration"""
    name: str
    description: str
    capabilities: List[str]
    personality: Dict[str, Any] = {}
    framework_preference: str = "auto"
    max_iterations: int = 10
    timeout: int = 300  # 5 minutes default
    temperature: float = 0.7
    model: str = "gpt-3.5-turbo"

    @field_validator('name')
    @classmethod
    def name_must_not_be_empty(cls, v):
        if not v or not v.strip():
            raise ValueError('Agent name cannot be empty')
        return v

    @field_validator('description')
    @classmethod
    def description_must_not_be_empty(cls, v):
        if not v or not v.strip():
            raise ValueError('Agent description cannot be empty')
        return v

class TaskRequest(BaseModel):
    """Universal task request format"""
    task: str
    context: Optional[Dict[str, Any]] = None
    tools: Optional[List[str]] = None
    max_iterations: Optional[int] = None
    timeout: Optional[int] = None

class TaskResponse(BaseModel):
    """Universal task response format"""
    task_id: str
    result: Any
    status: str  # 'completed', 'failed', 'timeout'
    execution_time: float
    framework_used: str
    agent_id: str
    metadata: Dict[str, Any] = {}
    error_message: Optional[str] = None

class BaseFrameworkWrapper(ABC):
    """
    Abstract base class for all framework wrappers.

    This ensures consistent interface across all AI frameworks while allowing
    framework-specific implementations and optimizations.
    """

    def __init__(self, agent_config: AgentConfig):
        self.agent_config = agent_config
        self.agent_id = str(uuid.uuid4())
        self.framework_type = self._get_framework_type()
        self.is_initialized = False
        self.tools = []
        self.agent = None
        self.memory = None
        self.created_at = time.time()

    @abstractmethod
    def _get_framework_type(self) -> FrameworkType:
        """Return the framework type for this wrapper"""
        pass

    @abstractmethod
    async def initialize(self) -> bool:
        """
        Initialize the framework-specific agent.

        Returns:
            bool: True if initialization successful, False otherwise
        """
        pass

    @abstractmethod
    async def execute(self, task_request: TaskRequest) -> TaskResponse:
        """
        Execute a task using the framework-specific agent.

        Args:
            task_request: Universal task request

        Returns:
            TaskResponse: Universal task response
        """
        pass

    @abstractmethod
    async def cleanup(self) -> bool:
        """
        Clean up framework-specific resources.

        Returns:
            bool: True if cleanup successful, False otherwise
        """
        pass

    # Common utility methods
    async def _capability_to_tool(self, capability: str) -> Optional[Any]:
        """Convert AgentOS capability to framework-specific tool"""
        # Base implementation - override in specific wrappers
        return None

    def _create_task_response(self, task_id: str, result: Any, status: str,
                            execution_time: float, error_message: str = None,
                            metadata: Dict[str, Any] = None) -> TaskResponse:
        """Create standardized task response"""
        return TaskResponse(
            task_id=task_id,
            result=result,
            status=status,
            execution_time=execution_time,
            framework_used=self.framework_type.value,
            agent_id=self.agent_id,
            metadata=metadata or {},
            error_message=error_message
        )

    async def _execute_with_timeout(self, coro, timeout: int) -> Any:
        """Execute coroutine with timeout"""
        try:
            return await asyncio.wait_for(coro, timeout=timeout)
        except asyncio.TimeoutError:
            raise TimeoutError(f"Task execution timed out after {timeout} seconds")

    def get_agent_info(self) -> Dict[str, Any]:
        """Get agent information"""
        return {
            "agent_id": self.agent_id,
            "framework": self.framework_type.value,
            "name": self.agent_config.name,
            "description": self.agent_config.description,
            "capabilities": self.agent_config.capabilities,
            "is_initialized": self.is_initialized,
            "tools_count": len(self.tools),
            "created_at": self.created_at,
            "uptime": time.time() - self.created_at
        }

    def get_performance_metrics(self) -> Dict[str, Any]:
        """Get performance metrics for this agent"""
        return {
            "agent_id": self.agent_id,
            "framework": self.framework_type.value,
            "uptime": time.time() - self.created_at,
            "is_healthy": self.is_initialized,
            "memory_usage": self._get_memory_usage(),
            "tools_available": len(self.tools)
        }

    def _get_memory_usage(self) -> Dict[str, Any]:
        """Get memory usage statistics - override in specific wrappers"""
        return {
            "working_memory": 0,
            "episodic_memory": 0,
            "semantic_memory": 0
        }

class FrameworkError(Exception):
    """Base exception for framework-related errors"""
    def __init__(self, message: str, framework: str, agent_id: str = None):
        self.message = message
        self.framework = framework
        self.agent_id = agent_id
        super().__init__(f"[{framework}] {message}")

class InitializationError(FrameworkError):
    """Exception raised when framework initialization fails"""
    pass

class ExecutionError(FrameworkError):
    """Exception raised when task execution fails"""
    pass

class TimeoutError(FrameworkError):
    """Exception raised when task execution times out"""
    pass
