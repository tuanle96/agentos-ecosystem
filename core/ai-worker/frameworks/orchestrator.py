#!/usr/bin/env python3
"""
AgentOS AI Worker - Framework Orchestrator
Week 3 Implementation: Intelligent Framework Selection

This module provides intelligent framework selection and orchestration
across LangChain, Swarms, CrewAI, and AutoGen based on task requirements.
"""

import logging
import asyncio
from typing import Dict, Any, List, Optional, Union
from enum import Enum

from .base_wrapper import (
    BaseFrameworkWrapper, FrameworkType, AgentConfig, 
    TaskRequest, TaskResponse
)

logger = logging.getLogger(__name__)

class TaskType(str, Enum):
    """Task type classification for framework selection"""
    GENERAL_PURPOSE = "general_purpose"
    MULTI_AGENT = "multi_agent"
    CONVERSATIONAL = "conversational"
    CODE_GENERATION = "code_generation"
    DISTRIBUTED = "distributed"
    WORKFLOW = "workflow"
    TOOL_HEAVY = "tool_heavy"
    MEMORY_INTENSIVE = "memory_intensive"

class FrameworkOrchestrator:
    """
    Intelligent framework orchestrator for AgentOS.
    
    Selects optimal AI framework based on task requirements, capabilities,
    and performance characteristics.
    """
    
    def __init__(self):
        self.framework_registry = {}
        self.performance_metrics = {}
        self.task_history = []
        self.framework_preferences = self._initialize_preferences()
        
    def _initialize_preferences(self) -> Dict[str, Dict[str, float]]:
        """Initialize framework preferences for different task types"""
        return {
            TaskType.GENERAL_PURPOSE: {
                FrameworkType.LANGCHAIN: 0.8,
                FrameworkType.SWARMS: 0.6,
                FrameworkType.CREWAI: 0.7,
                FrameworkType.AUTOGEN: 0.6
            },
            TaskType.MULTI_AGENT: {
                FrameworkType.LANGCHAIN: 0.6,
                FrameworkType.SWARMS: 0.9,
                FrameworkType.CREWAI: 0.9,
                FrameworkType.AUTOGEN: 0.7
            },
            TaskType.CONVERSATIONAL: {
                FrameworkType.LANGCHAIN: 0.7,
                FrameworkType.SWARMS: 0.5,
                FrameworkType.CREWAI: 0.6,
                FrameworkType.AUTOGEN: 0.9
            },
            TaskType.CODE_GENERATION: {
                FrameworkType.LANGCHAIN: 0.6,
                FrameworkType.SWARMS: 0.5,
                FrameworkType.CREWAI: 0.5,
                FrameworkType.AUTOGEN: 0.9
            },
            TaskType.DISTRIBUTED: {
                FrameworkType.LANGCHAIN: 0.5,
                FrameworkType.SWARMS: 0.9,
                FrameworkType.CREWAI: 0.7,
                FrameworkType.AUTOGEN: 0.6
            },
            TaskType.WORKFLOW: {
                FrameworkType.LANGCHAIN: 0.7,
                FrameworkType.SWARMS: 0.7,
                FrameworkType.CREWAI: 0.9,
                FrameworkType.AUTOGEN: 0.6
            },
            TaskType.TOOL_HEAVY: {
                FrameworkType.LANGCHAIN: 0.9,
                FrameworkType.SWARMS: 0.7,
                FrameworkType.CREWAI: 0.7,
                FrameworkType.AUTOGEN: 0.6
            },
            TaskType.MEMORY_INTENSIVE: {
                FrameworkType.LANGCHAIN: 0.8,
                FrameworkType.SWARMS: 0.6,
                FrameworkType.CREWAI: 0.6,
                FrameworkType.AUTOGEN: 0.7
            }
        }
    
    def register_framework(self, framework_type: FrameworkType, 
                          wrapper_class: type) -> bool:
        """Register a framework wrapper"""
        try:
            self.framework_registry[framework_type] = wrapper_class
            self.performance_metrics[framework_type] = {
                "total_executions": 0,
                "successful_executions": 0,
                "average_execution_time": 0.0,
                "error_rate": 0.0,
                "last_updated": 0
            }
            logger.info(f"Registered framework: {framework_type}")
            return True
        except Exception as e:
            logger.error(f"Failed to register framework {framework_type}: {str(e)}")
            return False
    
    def get_available_frameworks(self) -> List[FrameworkType]:
        """Get list of available frameworks"""
        return list(self.framework_registry.keys())
    
    async def select_optimal_framework(self, task_request: TaskRequest, 
                                     agent_config: AgentConfig) -> FrameworkType:
        """
        Select optimal framework based on task requirements and performance.
        
        Args:
            task_request: The task to be executed
            agent_config: Agent configuration
            
        Returns:
            FrameworkType: Optimal framework for the task
        """
        try:
            # Check for explicit framework preference
            if (agent_config.framework_preference != "auto" and 
                agent_config.framework_preference in [f.value for f in self.framework_registry.keys()]):
                preferred_framework = FrameworkType(agent_config.framework_preference)
                if preferred_framework in self.framework_registry:
                    logger.info(f"Using preferred framework: {preferred_framework}")
                    return preferred_framework
            
            # Analyze task to determine type
            task_type = await self._analyze_task_type(task_request)
            
            # Calculate framework scores
            framework_scores = await self._calculate_framework_scores(
                task_type, task_request, agent_config
            )
            
            # Select framework with highest score
            optimal_framework = max(framework_scores.items(), key=lambda x: x[1])[0]
            
            logger.info(f"Selected optimal framework: {optimal_framework} for task type: {task_type}")
            logger.debug(f"Framework scores: {framework_scores}")
            
            return optimal_framework
            
        except Exception as e:
            logger.error(f"Framework selection failed: {str(e)}")
            # Fallback to LangChain if available
            if FrameworkType.LANGCHAIN in self.framework_registry:
                return FrameworkType.LANGCHAIN
            # Otherwise return first available framework
            return list(self.framework_registry.keys())[0]
    
    async def _analyze_task_type(self, task_request: TaskRequest) -> TaskType:
        """Analyze task to determine its type"""
        task_text = task_request.task.lower()
        context = task_request.context or {}
        tools = task_request.tools or []
        
        # Multi-agent indicators
        multi_agent_keywords = ["team", "collaborate", "multiple agents", "coordinate", "parallel"]
        if any(keyword in task_text for keyword in multi_agent_keywords):
            return TaskType.MULTI_AGENT
        
        # Conversational indicators
        conversational_keywords = ["chat", "conversation", "discuss", "dialogue", "talk"]
        if any(keyword in task_text for keyword in conversational_keywords):
            return TaskType.CONVERSATIONAL
        
        # Code generation indicators
        code_keywords = ["code", "program", "script", "function", "class", "implement"]
        if any(keyword in task_text for keyword in code_keywords):
            return TaskType.CODE_GENERATION
        
        # Distributed processing indicators
        distributed_keywords = ["distribute", "parallel", "concurrent", "scale", "swarm"]
        if any(keyword in task_text for keyword in distributed_keywords):
            return TaskType.DISTRIBUTED
        
        # Workflow indicators
        workflow_keywords = ["workflow", "sequence", "steps", "process", "pipeline"]
        if any(keyword in task_text for keyword in workflow_keywords):
            return TaskType.WORKFLOW
        
        # Tool-heavy indicators
        if len(tools) > 3:
            return TaskType.TOOL_HEAVY
        
        # Memory-intensive indicators
        memory_keywords = ["remember", "recall", "history", "context", "memory"]
        if any(keyword in task_text for keyword in memory_keywords):
            return TaskType.MEMORY_INTENSIVE
        
        # Default to general purpose
        return TaskType.GENERAL_PURPOSE
    
    async def _calculate_framework_scores(self, task_type: TaskType, 
                                        task_request: TaskRequest,
                                        agent_config: AgentConfig) -> Dict[FrameworkType, float]:
        """Calculate scores for each framework based on task requirements"""
        scores = {}
        
        for framework_type in self.framework_registry.keys():
            # Base preference score
            base_score = self.framework_preferences.get(task_type, {}).get(framework_type, 0.5)
            
            # Performance adjustment
            performance_score = self._get_performance_score(framework_type)
            
            # Capability match score
            capability_score = self._get_capability_score(framework_type, agent_config.capabilities)
            
            # Tool compatibility score
            tool_score = self._get_tool_compatibility_score(framework_type, task_request.tools or [])
            
            # Calculate weighted final score
            final_score = (
                base_score * 0.4 +
                performance_score * 0.3 +
                capability_score * 0.2 +
                tool_score * 0.1
            )
            
            scores[framework_type] = final_score
        
        return scores
    
    def _get_performance_score(self, framework_type: FrameworkType) -> float:
        """Get performance score for framework"""
        metrics = self.performance_metrics.get(framework_type, {})
        
        if metrics.get("total_executions", 0) == 0:
            return 0.7  # Default score for new frameworks
        
        success_rate = metrics.get("successful_executions", 0) / metrics.get("total_executions", 1)
        avg_time = metrics.get("average_execution_time", 1.0)
        
        # Score based on success rate and execution time
        time_score = max(0, 1 - (avg_time / 10.0))  # Penalize slow execution
        performance_score = (success_rate * 0.7) + (time_score * 0.3)
        
        return min(1.0, performance_score)
    
    def _get_capability_score(self, framework_type: FrameworkType, 
                            capabilities: List[str]) -> float:
        """Get capability compatibility score"""
        # Framework capability strengths
        framework_capabilities = {
            FrameworkType.LANGCHAIN: ["web_search", "api_calls", "text_processing", "calculations"],
            FrameworkType.SWARMS: ["distributed", "parallel", "coordination"],
            FrameworkType.CREWAI: ["role_based", "workflow", "collaboration"],
            FrameworkType.AUTOGEN: ["conversation", "code_generation", "iterative"]
        }
        
        supported_capabilities = framework_capabilities.get(framework_type, [])
        
        if not capabilities:
            return 0.7  # Default score
        
        # Calculate overlap
        overlap = len(set(capabilities) & set(supported_capabilities))
        score = overlap / len(capabilities) if capabilities else 0.7
        
        return min(1.0, score)
    
    def _get_tool_compatibility_score(self, framework_type: FrameworkType, 
                                    tools: List[str]) -> float:
        """Get tool compatibility score"""
        if not tools:
            return 0.8  # Default score when no specific tools required
        
        # All frameworks currently support basic tools
        return 0.8
    
    def update_performance_metrics(self, framework_type: FrameworkType, 
                                 execution_time: float, success: bool):
        """Update performance metrics for a framework"""
        if framework_type not in self.performance_metrics:
            return
        
        metrics = self.performance_metrics[framework_type]
        
        # Update counters
        metrics["total_executions"] += 1
        if success:
            metrics["successful_executions"] += 1
        
        # Update average execution time
        current_avg = metrics["average_execution_time"]
        total_executions = metrics["total_executions"]
        
        new_avg = ((current_avg * (total_executions - 1)) + execution_time) / total_executions
        metrics["average_execution_time"] = new_avg
        
        # Update error rate
        metrics["error_rate"] = 1 - (metrics["successful_executions"] / metrics["total_executions"])
        
        # Update timestamp
        import time
        metrics["last_updated"] = time.time()
    
    def get_framework_statistics(self) -> Dict[str, Any]:
        """Get comprehensive framework statistics"""
        return {
            "available_frameworks": [f.value for f in self.get_available_frameworks()],
            "performance_metrics": {
                f.value: metrics for f, metrics in self.performance_metrics.items()
            },
            "task_history_count": len(self.task_history),
            "framework_preferences": {
                task_type.value: {
                    framework.value: score 
                    for framework, score in preferences.items()
                }
                for task_type, preferences in self.framework_preferences.items()
            }
        }
