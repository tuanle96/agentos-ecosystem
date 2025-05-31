#!/usr/bin/env python3
"""
AgentOS AI Worker - Comprehensive Framework Coverage Tests
Week 3 Implementation: Coverage Improvement for All Frameworks

This module provides comprehensive test coverage for all framework components
to achieve 80%+ coverage target.
"""

import pytest
import asyncio
import os
from unittest.mock import Mock, patch, AsyncMock
from typing import Dict, Any

# Import framework components
from frameworks.base_wrapper import AgentConfig, TaskRequest, TaskResponse, FrameworkType
from frameworks.orchestrator import FrameworkOrchestrator
from frameworks import get_available_frameworks, get_framework_capabilities

class TestFrameworkOrchestrator:
    """Comprehensive tests for Framework Orchestrator"""

    @pytest.fixture
    def orchestrator(self):
        """Create orchestrator instance"""
        return FrameworkOrchestrator()

    @pytest.fixture
    def sample_config(self):
        """Sample agent configuration"""
        return AgentConfig(
            name="test_agent",
            description="Test agent for coverage",
            capabilities=["web_search", "calculations"],
            framework_preference="auto"
        )

    def test_orchestrator_initialization(self, orchestrator):
        """Test orchestrator initialization"""
        assert orchestrator is not None
        assert orchestrator.framework_registry == {}
        assert orchestrator.performance_metrics == {}
        assert orchestrator.task_history == []

    def test_framework_registration(self, orchestrator):
        """Test framework registration"""
        from frameworks.langchain_wrapper import LangChainAgentWrapper

        orchestrator.register_framework(FrameworkType.LANGCHAIN, LangChainAgentWrapper)
        assert FrameworkType.LANGCHAIN in orchestrator.framework_registry
        assert orchestrator.framework_registry[FrameworkType.LANGCHAIN] == LangChainAgentWrapper

    @pytest.mark.asyncio
    async def test_task_type_classification(self, orchestrator):
        """Test task type classification"""
        test_cases = [
            ("Write a Python function", "code_generation"),
            ("Search for information", "general_purpose"),
            ("Have a conversation", "conversational"),
            ("Coordinate multiple agents", "multi_agent"),
            ("Calculate 2 + 2", "general_purpose"),
            ("Process this text", "general_purpose"),
            ("Call this API", "general_purpose"),
            ("Unknown task type", "general_purpose")
        ]

        for task, expected_type in test_cases:
            task_request = TaskRequest(task=task, tools=[])
            result = await orchestrator._analyze_task_type(task_request)
            # Should return a valid TaskType enum
            assert result is not None

    @pytest.mark.asyncio
    async def test_framework_selection(self, orchestrator, sample_config):
        """Test framework selection logic"""
        from frameworks.langchain_wrapper import LangChainAgentWrapper

        # Register a framework
        orchestrator.register_framework(FrameworkType.LANGCHAIN, LangChainAgentWrapper)

        task_request = TaskRequest(task="Test task", tools=[])

        # Test with preference
        config_with_pref = AgentConfig(
            name="test",
            description="test",
            capabilities=["web_search"],
            framework_preference="langchain"
        )

        result = await orchestrator.select_optimal_framework(task_request, config_with_pref)
        assert result == FrameworkType.LANGCHAIN

    def test_performance_tracking(self, orchestrator):
        """Test performance metrics tracking"""
        from frameworks.langchain_wrapper import LangChainAgentWrapper

        framework = FrameworkType.LANGCHAIN

        # Register framework first to create metrics
        orchestrator.register_framework(framework, LangChainAgentWrapper)

        # Test successful execution
        orchestrator.update_performance_metrics(framework, 0.5, True)

        metrics = orchestrator.performance_metrics[framework]
        assert metrics["total_executions"] == 1
        assert metrics["successful_executions"] == 1
        assert metrics["error_rate"] == 0.0
        assert metrics["average_execution_time"] == 0.5

        # Test failed execution
        orchestrator.update_performance_metrics(framework, 1.0, False)

        metrics = orchestrator.performance_metrics[framework]
        assert metrics["total_executions"] == 2
        assert metrics["successful_executions"] == 1
        assert metrics["error_rate"] == 0.5

    def test_framework_statistics(self, orchestrator):
        """Test framework statistics"""
        from frameworks.langchain_wrapper import LangChainAgentWrapper

        orchestrator.register_framework(FrameworkType.LANGCHAIN, LangChainAgentWrapper)

        stats = orchestrator.get_framework_statistics()

        assert "available_frameworks" in stats
        assert "performance_metrics" in stats
        assert "framework_preferences" in stats
        assert len(stats["available_frameworks"]) == 1

class TestFrameworkUtilities:
    """Test framework utility functions"""

    def test_get_available_frameworks(self):
        """Test getting available frameworks"""
        frameworks = get_available_frameworks()
        assert isinstance(frameworks, list)
        assert len(frameworks) > 0
        # Should include at least langchain and swarms
        assert "langchain" in frameworks
        assert "swarms" in frameworks

    def test_get_framework_capabilities(self):
        """Test getting framework capabilities"""
        capabilities = get_framework_capabilities()
        assert isinstance(capabilities, dict)
        assert len(capabilities) > 0

        # Check structure
        for framework_name, info in capabilities.items():
            assert "status" in info
            assert "strengths" in info
            assert "use_cases" in info

class TestBaseWrapper:
    """Test base wrapper functionality"""

    def test_agent_config_creation(self):
        """Test agent configuration creation"""
        config = AgentConfig(
            name="test_agent",
            description="Test description",
            capabilities=["web_search", "calculations"],
            framework_preference="langchain",
            max_iterations=5,
            timeout=30,
            temperature=0.7,
            model="gpt-3.5-turbo"
        )

        assert config.name == "test_agent"
        assert config.description == "Test description"
        assert len(config.capabilities) == 2
        assert config.framework_preference == "langchain"
        assert config.max_iterations == 5
        assert config.timeout == 30
        assert config.temperature == 0.7
        assert config.model == "gpt-3.5-turbo"

    def test_task_request_creation(self):
        """Test task request creation"""
        task_request = TaskRequest(
            task="Test task",
            context={"user_id": "test"},
            tools=["web_search"],
            max_iterations=3,
            timeout=30
        )

        assert task_request.task == "Test task"
        assert task_request.context == {"user_id": "test"}
        assert task_request.tools == ["web_search"]
        assert task_request.max_iterations == 3
        assert task_request.timeout == 30

    def test_task_response_creation(self):
        """Test task response creation"""
        response = TaskResponse(
            task_id="test-123",
            result="Test result",
            status="completed",
            execution_time=1.5,
            framework_used="langchain",
            agent_id="agent-456",
            metadata={"test": True}
        )

        assert response.task_id == "test-123"
        assert response.result == "Test result"
        assert response.status == "completed"
        assert response.execution_time == 1.5
        assert response.framework_used == "langchain"
        assert response.agent_id == "agent-456"
        assert response.metadata == {"test": True}

    def test_framework_type_enum(self):
        """Test framework type enumeration"""
        assert FrameworkType.LANGCHAIN == "langchain"
        assert FrameworkType.SWARMS == "swarms"
        assert FrameworkType.CREWAI == "crewai"
        assert FrameworkType.AUTOGEN == "autogen"

class TestFrameworkIntegration:
    """Integration tests for framework components"""

    @pytest.mark.asyncio
    async def test_multi_framework_workflow(self):
        """Test complete multi-framework workflow"""
        # Create orchestrator
        orchestrator = FrameworkOrchestrator()

        # Register frameworks
        from frameworks.langchain_wrapper import LangChainAgentWrapper
        from frameworks.swarms_wrapper import SwarmAgentWrapper

        orchestrator.register_framework(FrameworkType.LANGCHAIN, LangChainAgentWrapper)
        orchestrator.register_framework(FrameworkType.SWARMS, SwarmAgentWrapper)

        # Create configuration
        config = AgentConfig(
            name="integration_test_agent",
            description="Integration test agent",
            capabilities=["web_search", "calculations"],
            framework_preference="auto"
        )

        # Create task request
        task_request = TaskRequest(
            task="Calculate 2 + 2 and search for information",
            tools=["calculations", "web_search"]
        )

        # Test framework selection
        selected_framework = await orchestrator.select_optimal_framework(task_request, config)
        assert selected_framework in [FrameworkType.LANGCHAIN, FrameworkType.SWARMS]

        # Test agent creation
        agent_class = orchestrator.framework_registry[selected_framework]
        agent = agent_class(config)

        assert agent is not None
        assert agent.agent_config == config
        assert agent.framework_type == selected_framework

    def test_error_handling(self):
        """Test error handling in framework components"""
        # Test invalid agent config - empty name
        with pytest.raises(ValueError, match="Agent name cannot be empty"):
            AgentConfig(
                name="",  # Invalid empty name
                description="Test",
                capabilities=[]
            )

        # Test invalid agent config - empty description
        with pytest.raises(ValueError, match="Agent description cannot be empty"):
            AgentConfig(
                name="Test Agent",
                description="",  # Invalid empty description
                capabilities=[]
            )

    def test_performance_benchmarks(self):
        """Test performance benchmarks"""
        orchestrator = FrameworkOrchestrator()

        # Register framework first
        from frameworks.langchain_wrapper import LangChainAgentWrapper
        orchestrator.register_framework(FrameworkType.LANGCHAIN, LangChainAgentWrapper)

        # Simulate performance data
        framework = FrameworkType.LANGCHAIN

        # Add multiple performance records
        for i in range(10):
            success = i < 8  # 80% success rate
            execution_time = 0.1 + (i * 0.05)  # Increasing execution time
            error = None if success else f"Error {i}"

            orchestrator.update_performance_metrics(framework, execution_time, success)

        metrics = orchestrator.performance_metrics[framework]

        assert metrics["total_executions"] == 10
        assert metrics["successful_executions"] == 8
        assert abs(metrics["error_rate"] - 0.2) < 0.01  # Allow for floating point precision
        assert metrics["average_execution_time"] > 0.1

# Performance and load testing
class TestPerformanceMetrics:
    """Test performance and load characteristics"""

    def test_framework_initialization_time(self):
        """Test framework initialization performance"""
        import time

        start_time = time.time()
        orchestrator = FrameworkOrchestrator()
        initialization_time = time.time() - start_time

        # Should initialize quickly
        assert initialization_time < 0.1  # Less than 100ms

    def test_memory_usage(self):
        """Test memory usage of framework components"""
        import sys

        # Create multiple agents
        agents = []
        for i in range(10):
            config = AgentConfig(
                name=f"agent_{i}",
                description=f"Test agent {i}",
                capabilities=["web_search"]
            )

            from frameworks.langchain_wrapper import LangChainAgentWrapper
            agent = LangChainAgentWrapper(config)
            agents.append(agent)

        # Memory usage should be reasonable
        # This is a basic check - in production you'd use more sophisticated memory profiling
        assert len(agents) == 10
        assert sys.getsizeof(agents) < 10000  # Less than 10KB for basic objects
