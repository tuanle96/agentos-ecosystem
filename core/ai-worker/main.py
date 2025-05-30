#!/usr/bin/env python3
"""
AgentOS AI Worker - LangChain Integration
Main FastAPI application for AI agent processing
"""

import os
import uvicorn
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Dict, Any, List, Optional

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

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "agentos-ai-worker",
        "version": "0.1.0",
        "framework": "langchain"
    }

@app.get("/agents")
async def list_agents():
    """List available agents"""
    return {"agents": ["langchain-agent", "openai-agent", "anthropic-agent"]}

@app.post("/agents/{agent_id}/execute")
async def execute_task(agent_id: str, request: TaskRequest):
    """Execute a task using specified agent"""
    try:
        # Placeholder implementation
        result = {
            "task_id": f"task_{agent_id}_{hash(request.task)}",
            "result": f"Executed task: {request.task}",
            "status": "completed",
            "execution_time": 1.5
        }
        return TaskResponse(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/tools")
async def list_tools():
    """List available tools"""
    return {"tools": ["web_search", "calculator", "file_reader", "code_executor"]}

if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=True,
        log_level="info"
    )
