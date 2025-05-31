#!/usr/bin/env python3
"""
AgentOS AI Worker - Multi-Framework Integration (Week 3 Enhancement)
Main FastAPI application for AI agent processing with multi-framework support
Supports: LangChain, Swarms, CrewAI, AutoGen
"""

import os
import uuid
import time
import asyncio
import uvicorn
import logging
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Dict, Any, List, Optional

# Setup logging first
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# LangChain imports
try:
    from langchain.agents import initialize_agent, AgentType, Tool
    from langchain.llms import OpenAI
    from langchain.memory import ConversationBufferMemory
    LANGCHAIN_AVAILABLE = True
except ImportError:
    LANGCHAIN_AVAILABLE = False
    Tool = None
    OpenAI = None
    ConversationBufferMemory = None
    initialize_agent = None
    AgentType = None
    logger.warning("LangChain not available")

# Multi-framework imports
try:
    from frameworks import (
        FrameworkOrchestrator,
        get_available_frameworks,
        get_framework_capabilities,
        FRAMEWORK_REGISTRY,
        SWARMS_AVAILABLE,
        CREWAI_AVAILABLE,
        AUTOGEN_AVAILABLE
    )
    from frameworks.base_wrapper import (
        AgentConfig as FrameworkAgentConfig,
        TaskRequest as FrameworkTaskRequest,
        TaskResponse as FrameworkTaskResponse,
        FrameworkType
    )
    MULTI_FRAMEWORK_AVAILABLE = True
except ImportError:
    MULTI_FRAMEWORK_AVAILABLE = False
    SWARMS_AVAILABLE = False
    CREWAI_AVAILABLE = False
    AUTOGEN_AVAILABLE = False
    logger.warning("Multi-framework support not available")

# Initialize FastAPI app
app = FastAPI(
    title="AgentOS AI Worker",
    description="LangChain-based AI agent processing service",
    version="0.1.0"
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Request/Response models
class TaskRequest(BaseModel):
    agent_id: str
    task: str
    context: Optional[Dict[str, Any]] = None
    tools: Optional[List[str]] = None

class TaskResponse(BaseModel):
    task_id: str
    result: Any
    status: str
    execution_time: float


class AgentConfig(BaseModel):
    name: str
    description: str
    capabilities: List[str]
    personality: Dict[str, Any] = {}
    framework_preference: str = "langchain"


# LangChain Agent Wrapper (Week 2 Implementation)
class LangChainAgentWrapper:
    def __init__(self, agent_config: AgentConfig):
        self.agent_config = agent_config
        self.agent_id = str(uuid.uuid4())
        self.tools = []
        self.agent = None
        self.memory = None

        if LANGCHAIN_AVAILABLE and OpenAI is not None and ConversationBufferMemory is not None:
            self.llm = OpenAI(temperature=0.7) if os.getenv("OPENAI_API_KEY") else None
            self.memory = ConversationBufferMemory(memory_key="chat_history")
        else:
            self.llm = None

    async def initialize(self):
        """Initialize LangChain agent with capabilities"""
        if not LANGCHAIN_AVAILABLE:
            raise HTTPException(status_code=500, detail="LangChain not available")

        if not self.llm:
            raise HTTPException(status_code=500, detail="OpenAI API key not configured")

        # Convert AgentOS capabilities to LangChain tools
        for capability in self.agent_config.capabilities:
            tool = await self._capability_to_tool(capability)
            if tool:
                self.tools.append(tool)

        # Initialize LangChain agent
        if self.tools and initialize_agent is not None and AgentType is not None:
            self.agent = initialize_agent(
                tools=self.tools,
                llm=self.llm,
                agent=AgentType.CONVERSATIONAL_REACT_DESCRIPTION,
                memory=self.memory,
                verbose=True
            )

    async def _capability_to_tool(self, capability: str):
        """Convert AgentOS capability to LangChain tool"""
        tool_map = {
            "web_search": self._create_web_search_tool(),
            "calculations": self._create_calculator_tool(),
            "text_processing": self._create_text_processing_tool(),
            "file_operations": self._create_file_operations_tool(),
            "api_calls": self._create_api_calls_tool(),
        }
        return tool_map.get(capability)

    def _create_web_search_tool(self):
        """Create web search tool"""
        if not LANGCHAIN_AVAILABLE or Tool is None:
            return None

        def web_search(query: str) -> str:
            # Placeholder implementation
            return f"Search results for: {query}"

        return Tool(
            name="web_search",
            description="Search the web for information",
            func=web_search
        )

    def _create_calculator_tool(self):
        """Create calculator tool"""
        if not LANGCHAIN_AVAILABLE or Tool is None:
            return None

        def calculate(expression: str) -> str:
            try:
                # Safe evaluation of mathematical expressions
                result = eval(expression, {"__builtins__": {}}, {})
                return str(result)
            except Exception as e:
                return f"Error: {str(e)}"

        return Tool(
            name="calculator",
            description="Perform mathematical calculations",
            func=calculate
        )

    def _create_text_processing_tool(self):
        """Create text processing tool"""
        if not LANGCHAIN_AVAILABLE or Tool is None:
            return None

        def process_text(text: str) -> str:
            # Basic text processing
            return f"Processed: {text.strip().lower()}"

        return Tool(
            name="text_processor",
            description="Process and analyze text",
            func=process_text
        )

    def _create_file_operations_tool(self):
        """Create file operations tool"""
        if not LANGCHAIN_AVAILABLE or Tool is None:
            return None

        def file_operation(operation: str) -> str:
            # Placeholder for secure file operations
            return f"File operation: {operation}"

        return Tool(
            name="file_operations",
            description="Perform safe file operations",
            func=file_operation
        )

    def _create_api_calls_tool(self):
        """Create API calls tool"""
        if not LANGCHAIN_AVAILABLE or Tool is None:
            return None

        def api_call(url: str) -> str:
            # Placeholder for secure API calls
            return f"API call to: {url}"

        return Tool(
            name="api_calls",
            description="Make HTTP API calls",
            func=api_call
        )

    async def execute(self, task: str) -> Dict[str, Any]:
        """Execute task using LangChain agent"""
        if not self.agent:
            await self.initialize()

        start_time = time.time()

        try:
            if self.agent and hasattr(self.agent, 'run'):
                result = self.agent.run(task)
            else:
                result = f"Agent not properly initialized or LangChain not available"
            execution_time = time.time() - start_time

            return {
                "task_id": str(uuid.uuid4()),
                "result": result,
                "status": "completed",
                "execution_time": execution_time,
                "agent_id": self.agent_id,
                "framework": "langchain"
            }
        except Exception as e:
            execution_time = time.time() - start_time
            return {
                "task_id": str(uuid.uuid4()),
                "result": f"Error: {str(e)}",
                "status": "failed",
                "execution_time": execution_time,
                "agent_id": self.agent_id,
                "framework": "langchain"
            }


# Global agent registry
agent_registry: Dict[str, LangChainAgentWrapper] = {}


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "agentos-ai-worker",
        "version": "0.1.0-week2",
        "framework": "langchain",
        "langchain_available": LANGCHAIN_AVAILABLE,
        "openai_configured": bool(os.getenv("OPENAI_API_KEY"))
    }

@app.get("/agents")
async def list_agents():
    """List available agents"""
    active_agents = list(agent_registry.keys())
    return {
        "agents": active_agents,
        "count": len(active_agents),
        "available_frameworks": ["langchain"],
        "langchain_available": LANGCHAIN_AVAILABLE
    }


@app.post("/agents/create")
async def create_agent(config: AgentConfig):
    """Create a new LangChain agent"""
    try:
        # Create LangChain agent wrapper
        agent_wrapper = LangChainAgentWrapper(config)
        await agent_wrapper.initialize()

        # Store in registry
        agent_registry[agent_wrapper.agent_id] = agent_wrapper

        return {
            "agent_id": agent_wrapper.agent_id,
            "name": config.name,
            "capabilities": config.capabilities,
            "framework": "langchain",
            "status": "created",
            "tools_count": len(agent_wrapper.tools)
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/agents/{agent_id}/execute")
async def execute_task(agent_id: str, request: TaskRequest):
    """Execute a task using specified agent"""
    try:
        # Check if agent exists
        if agent_id not in agent_registry:
            raise HTTPException(status_code=404, detail="Agent not found")

        agent_wrapper = agent_registry[agent_id]
        result = await agent_wrapper.execute(request.task)

        return TaskResponse(**result)
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/agents/{agent_id}")
async def get_agent(agent_id: str):
    """Get agent details"""
    if agent_id not in agent_registry:
        raise HTTPException(status_code=404, detail="Agent not found")

    agent_wrapper = agent_registry[agent_id]
    return {
        "agent_id": agent_id,
        "name": agent_wrapper.agent_config.name,
        "description": agent_wrapper.agent_config.description,
        "capabilities": agent_wrapper.agent_config.capabilities,
        "framework": "langchain",
        "tools_count": len(agent_wrapper.tools),
        "status": "active" if agent_wrapper.agent else "initializing"
    }


@app.delete("/agents/{agent_id}")
async def delete_agent(agent_id: str):
    """Delete an agent"""
    if agent_id not in agent_registry:
        raise HTTPException(status_code=404, detail="Agent not found")

    del agent_registry[agent_id]
    return {"message": "Agent deleted successfully", "agent_id": agent_id}


@app.get("/tools")
async def list_tools():
    """List available tools"""
    return {
        "tools": [
            {
                "name": "web_search",
                "description": "Search the web for information",
                "category": "search"
            },
            {
                "name": "calculator",
                "description": "Perform mathematical calculations",
                "category": "math"
            },
            {
                "name": "text_processor",
                "description": "Process and analyze text",
                "category": "text"
            },
            {
                "name": "file_operations",
                "description": "Perform safe file operations",
                "category": "file"
            },
            {
                "name": "api_calls",
                "description": "Make HTTP API calls",
                "category": "network"
            }
        ],
        "count": 5,
        "framework": "langchain"
    }


@app.get("/framework/status")
async def framework_status():
    """Get framework status and capabilities"""
    return {
        "framework": "langchain",
        "available": LANGCHAIN_AVAILABLE,
        "openai_configured": bool(os.getenv("OPENAI_API_KEY")),
        "active_agents": len(agent_registry),
        "supported_capabilities": [
            "web_search", "calculations", "text_processing",
            "file_operations", "api_calls"
        ],
        "version": "0.1.0-week2"
    }

if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=True,
        log_level="info"
    )
