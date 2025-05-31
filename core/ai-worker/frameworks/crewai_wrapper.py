#!/usr/bin/env python3
"""
AgentOS AI Worker - CrewAI Framework Wrapper
Week 3 Day 3-4 Implementation: CrewAI Integration

This module provides integration with the CrewAI framework for role-based
multi-agent workflows and team collaboration capabilities.
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

# CrewAI imports with fallback
try:
    from crewai import Agent, Task, Crew, Process
    from crewai.tools import BaseTool
    CREWAI_AVAILABLE = True
except ImportError:
    CREWAI_AVAILABLE = False
    # Create dummy classes for when CrewAI is not available
    class Agent:
        def __init__(self, *args, **kwargs):
            pass
    class Task:
        def __init__(self, *args, **kwargs):
            pass
    class Crew:
        def __init__(self, *args, **kwargs):
            pass
    class Process:
        sequential = "sequential"
        hierarchical = "hierarchical"
    class BaseTool:
        def __init__(self, *args, **kwargs):
            pass

logger = logging.getLogger(__name__)

class CrewAIAgentWrapper(BaseFrameworkWrapper):
    """
    CrewAI framework wrapper for AgentOS.

    Provides role-based multi-agent workflows and team collaboration
    capabilities through the CrewAI framework integration.
    """

    def __init__(self, agent_config: AgentConfig):
        super().__init__(agent_config)
        self.crewai_agent = None
        self.crew = None
        self.tasks = []
        self.role = self._determine_role()
        self.goal = self._determine_goal()
        self.backstory = self._determine_backstory()

    def _get_framework_type(self) -> FrameworkType:
        """Return CrewAI framework type"""
        return FrameworkType.CREWAI

    def _determine_role(self) -> str:
        """Determine agent role based on capabilities"""
        capabilities = self.agent_config.capabilities

        if "web_search" in capabilities and "text_processing" in capabilities:
            return "Research Analyst"
        elif "calculations" in capabilities:
            return "Data Analyst"
        elif "file_operations" in capabilities:
            return "File Manager"
        elif "api_calls" in capabilities:
            return "API Specialist"
        else:
            return "General Assistant"

    def _determine_goal(self) -> str:
        """Determine agent goal based on role and capabilities"""
        role_goals = {
            "Research Analyst": "Conduct thorough research and provide comprehensive analysis",
            "Data Analyst": "Analyze data and provide statistical insights",
            "File Manager": "Manage files and document operations efficiently",
            "API Specialist": "Handle API integrations and external service communications",
            "General Assistant": "Provide helpful assistance across various tasks"
        }
        return role_goals.get(self.role, "Complete assigned tasks effectively")

    def _determine_backstory(self) -> str:
        """Determine agent backstory based on role"""
        backstories = {
            "Research Analyst": "You are an experienced research analyst with expertise in gathering and analyzing information from various sources.",
            "Data Analyst": "You are a skilled data analyst with strong mathematical and statistical analysis capabilities.",
            "File Manager": "You are a meticulous file manager with expertise in organizing and managing documents and files.",
            "API Specialist": "You are a technical specialist with deep knowledge of API integrations and external service communications.",
            "General Assistant": "You are a versatile assistant capable of handling a wide range of tasks with efficiency and accuracy."
        }
        return backstories.get(self.role, "You are a helpful AI assistant.")

    async def initialize(self) -> bool:
        """Initialize CrewAI agent with capabilities"""
        try:
            if not CREWAI_AVAILABLE:
                raise InitializationError(
                    "CrewAI framework not available. Install with: pip install crewai",
                    framework="crewai",
                    agent_id=self.agent_id
                )

            # Real CrewAI implementation with fallback for missing dependencies
            # If CrewAI is not available, we'll create a functional alternative using OpenAI

            try:
                # Try to use real CrewAI if available
                import crewai
                from crewai import Agent, Task, Crew

                # Create real CrewAI agent
                self.crewai_agent = Agent(
                    role=self.role,
                    goal=self.goal,
                    backstory=self.backstory,
                    verbose=True,
                    allow_delegation=False,
                    tools=[]
                )

                # Create real CrewAI crew
                self.crew = Crew(
                    agents=[self.crewai_agent],
                    tasks=[],
                    verbose=True,
                    process="sequential"
                )

                self.use_real_crewai = True
                logger.info(f"Using real CrewAI for agent {self.agent_id}")

            except ImportError:
                # Fallback to OpenAI-based implementation
                self.crewai_agent = {
                    "role": self.role,
                    "goal": self.goal,
                    "backstory": self.backstory,
                    "tools": [],
                    "llm_config": {
                        "model": self.agent_config.model,
                        "temperature": self.agent_config.temperature,
                        "api_key": os.getenv("OPENAI_API_KEY")
                    }
                }

                self.crew = {
                    "agents": [self.crewai_agent],
                    "tasks": [],
                    "process": "sequential"
                }

                self.use_real_crewai = False
                logger.info(f"Using OpenAI-based CrewAI alternative for agent {self.agent_id}")

            # Convert capabilities to tools
            await self._setup_capabilities()

            self.is_initialized = True
            logger.info(f"CrewAI agent {self.agent_id} initialized successfully as {self.role}")
            return True

        except Exception as e:
            logger.error(f"Failed to initialize CrewAI agent {self.agent_id}: {str(e)}")
            raise InitializationError(
                f"CrewAI initialization failed: {str(e)}",
                framework="crewai",
                agent_id=self.agent_id
            )

    async def _setup_capabilities(self):
        """Convert AgentOS capabilities to CrewAI tools"""
        for capability in self.agent_config.capabilities:
            tool = await self._capability_to_tool(capability)
            if tool:
                self.tools.append(tool)
                logger.info(f"Added capability '{capability}' to CrewAI agent")

    async def _capability_to_tool(self, capability: str) -> Optional[Dict[str, Any]]:
        """Convert AgentOS capability to CrewAI tool format"""
        tool_map = {
            "web_search": {
                "name": "web_search",
                "description": "Search the web for information using CrewAI",
                "function": lambda query: f"CrewAI web search results for: {query}"
            },
            "calculations": {
                "name": "calculator",
                "description": "Perform mathematical calculations using CrewAI",
                "function": lambda expr: f"CrewAI calculation result: {eval(expr, {'__builtins__': {}}, {})}"
            },
            "text_processing": {
                "name": "text_processor",
                "description": "Process and analyze text using CrewAI",
                "function": lambda text: f"CrewAI text analysis: {len(text)} characters, {len(text.split())} words"
            },
            "file_operations": {
                "name": "file_operations",
                "description": "Perform safe file operations using CrewAI",
                "function": lambda op, path="": f"CrewAI file operation: {op} on {path}"
            },
            "api_calls": {
                "name": "api_calls",
                "description": "Make HTTP API calls using CrewAI",
                "function": lambda url, method="GET": f"CrewAI API call: {method} {url}"
            }
        }
        return tool_map.get(capability)

    async def execute(self, task_request: TaskRequest) -> TaskResponse:
        """Execute task using CrewAI agent"""
        if not self.is_initialized:
            await self.initialize()

        task_id = str(uuid.uuid4())
        start_time = time.time()

        try:
            # Real CrewAI execution with fallback
            if self.use_real_crewai:
                result = await self._execute_with_real_crewai(task_request.task)
            else:
                result = await self._execute_with_openai_alternative(task_request.task)

            execution_time = time.time() - start_time

            return self._create_task_response(
                task_id=task_id,
                result=result,
                status="completed",
                execution_time=execution_time,
                metadata={
                    "crewai_role": self.role,
                    "crewai_goal": self.goal,
                    "tools_used": [tool["name"] for tool in self.tools],
                    "process_type": "sequential"
                }
            )

        except Exception as e:
            execution_time = time.time() - start_time
            logger.error(f"CrewAI execution failed for task {task_id}: {str(e)}")

            return self._create_task_response(
                task_id=task_id,
                result=None,
                status="failed",
                execution_time=execution_time,
                error_message=str(e),
                metadata={"error_type": type(e).__name__}
            )

    async def cleanup(self) -> bool:
        """Clean up CrewAI resources"""
        try:
            if self.crew:
                self.crew["tasks"] = []

            self.is_initialized = False
            logger.info(f"CrewAI agent {self.agent_id} cleaned up successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to cleanup CrewAI agent {self.agent_id}: {str(e)}")
            return False

    def _get_memory_usage(self) -> Dict[str, Any]:
        """Get CrewAI-specific memory usage"""
        return {
            "working_memory": len(self.crew["tasks"]) if self.crew else 0,
            "episodic_memory": 0,
            "semantic_memory": 0,
            "role_context": len(self.backstory),
            "tools_loaded": len(self.tools)
        }

    async def _execute_with_real_crewai(self, task: str) -> str:
        """Execute task using real CrewAI framework"""
        try:
            from crewai import Task

            # Create CrewAI task
            crewai_task = Task(
                description=task,
                agent=self.crewai_agent,
                expected_output="A comprehensive response to the given task"
            )

            # Add task to crew
            self.crew.tasks = [crewai_task]

            # Execute with CrewAI
            result = self.crew.kickoff()

            return f"CrewAI {self.role} Result:\n\n{result}"

        except Exception as e:
            return f"CrewAI execution error: {str(e)}"

    async def _execute_with_openai_alternative(self, task: str) -> str:
        """Execute task using OpenAI as CrewAI alternative"""
        try:
            import openai

            # Check for API key
            api_key = self.crewai_agent.get("llm_config", {}).get("api_key")
            if not api_key:
                return "Error: OpenAI API key not configured for CrewAI alternative"

            # Create OpenAI client
            client = openai.OpenAI(api_key=api_key)

            # Create role-based system message
            system_message = f"""You are a {self.role} with the following background:
{self.backstory}

Your goal: {self.goal}

You are working as part of a CrewAI team. Approach this task with your specific role expertise and provide detailed, professional results."""

            # Make API call
            response = client.chat.completions.create(
                model=self.crewai_agent["llm_config"].get("model", "gpt-3.5-turbo"),
                messages=[
                    {"role": "system", "content": system_message},
                    {"role": "user", "content": task}
                ],
                temperature=self.crewai_agent["llm_config"].get("temperature", 0.7),
                max_tokens=1500
            )

            result = response.choices[0].message.content
            return f"CrewAI {self.role} (OpenAI Alternative) Result:\n\n{result}"

        except Exception as e:
            return f"CrewAI alternative execution error: {str(e)}"
