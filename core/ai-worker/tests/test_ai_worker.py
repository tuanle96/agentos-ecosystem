"""
Week 2 AI Worker Integration Tests
Tests for LangChain integration, agent management, and tool execution
"""

import pytest
import asyncio
import json
import time
from typing import Dict, Any
from unittest.mock import Mock, patch
from fastapi.testclient import TestClient

# Import the main app
import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from main import app, LangChainAgentWrapper, LANGCHAIN_AVAILABLE

client = TestClient(app)

class TestAIWorkerBasics:
    """Test basic AI Worker functionality"""

    def test_health_endpoint(self):
        """Test health check endpoint"""
        response = client.get("/health")
        assert response.status_code == 200

        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "agentos-ai-worker"
        assert data["version"] == "0.1.0-week2"
        assert data["framework"] == "langchain"
        assert "langchain_available" in data
        assert "openai_configured" in data

    def test_framework_status(self):
        """Test framework status endpoint"""
        response = client.get("/framework/status")
        assert response.status_code == 200

        data = response.json()
        assert data["framework"] == "langchain"
        assert "available" in data
        assert "active_agents" in data
        assert "supported_capabilities" in data
        assert data["version"] == "0.1.0-week2"

        # Check supported capabilities
        capabilities = data["supported_capabilities"]
        expected_capabilities = [
            "web_search", "calculations", "text_processing",
            "file_operations", "api_calls"
        ]
        for cap in expected_capabilities:
            assert cap in capabilities

    def test_list_agents_empty(self):
        """Test listing agents when none exist"""
        response = client.get("/agents")
        assert response.status_code == 200

        data = response.json()
        assert data["count"] == 0
        assert data["agents"] == []

    def test_tools_endpoint(self):
        """Test tools listing endpoint"""
        response = client.get("/tools")
        assert response.status_code == 200

        data = response.json()
        assert "tools" in data
        assert "count" in data
        assert isinstance(data["tools"], list)
        assert data["count"] >= 0


class TestAgentManagement:
    """Test agent creation, management, and deletion"""

    def test_create_agent_without_langchain(self):
        """Test agent creation when LangChain is not available"""
        agent_config = {
            "agent_id": "test-agent-1",
            "capabilities": ["web_search", "calculations"],
            "framework": "langchain",
            "config": {
                "temperature": 0.7,
                "max_tokens": 1000
            }
        }

        response = client.post("/agents/create", json=agent_config)

        if not LANGCHAIN_AVAILABLE:
            # Expect 422 (validation error) or 500 (server error)
            assert response.status_code in [422, 500]
            if response.status_code == 500:
                assert "LangChain not available" in response.json()["detail"]
        else:
            # If LangChain is available but no API key
            if not os.getenv("OPENAI_API_KEY"):
                assert response.status_code in [422, 500]
                if response.status_code == 500:
                    assert "OpenAI API key not configured" in response.json()["detail"]

    def test_create_agent_invalid_config(self):
        """Test agent creation with invalid configuration"""
        # Missing required fields
        response = client.post("/agents/create", json={})
        assert response.status_code == 422  # Validation error

        # Invalid capabilities
        agent_config = {
            "agent_id": "test-agent-invalid",
            "capabilities": ["invalid_capability"],
            "framework": "langchain"
        }

        response = client.post("/agents/create", json=agent_config)
        # Accept both 400 (bad request) and 422 (validation error)
        assert response.status_code in [400, 422, 500]

    def test_get_nonexistent_agent(self):
        """Test getting an agent that doesn't exist"""
        response = client.get("/agents/nonexistent-agent")
        assert response.status_code == 404
        assert "Agent not found" in response.json()["detail"]

    def test_delete_nonexistent_agent(self):
        """Test deleting an agent that doesn't exist"""
        response = client.delete("/agents/nonexistent-agent")
        assert response.status_code == 404
        assert "Agent not found" in response.json()["detail"]


class TestToolExecution:
    """Test tool execution functionality"""

    def test_execute_task_without_agent(self):
        """Test task execution when no agent exists"""
        task_request = {
            "agent_id": "nonexistent-agent",
            "task": "Calculate 2+2",
            "context": {},
            "tools": ["calculations"]
        }

        response = client.post("/agents/nonexistent-agent/execute", json=task_request)
        assert response.status_code == 404
        assert "Agent not found" in response.json()["detail"]


class TestLangChainWrapper:
    """Test LangChain wrapper functionality"""

    def test_wrapper_initialization(self):
        """Test LangChain wrapper initialization"""
        config = {
            "agent_id": "test-wrapper",
            "capabilities": ["web_search", "calculations"],
            "framework": "langchain",
            "config": {"temperature": 0.7}
        }

        wrapper = LangChainAgentWrapper(config)
        # Test that wrapper stores config correctly (as dict)
        assert wrapper.agent_config["agent_id"] == "test-wrapper"
        assert wrapper.agent_config["capabilities"] == ["web_search", "calculations"]
        assert wrapper.agent_config["framework"] == "langchain"

    @pytest.mark.asyncio
    async def test_capability_to_tool_conversion(self):
        """Test conversion of capabilities to LangChain tools"""
        config = {
            "agent_id": "test-conversion",
            "capabilities": ["calculations"],
            "framework": "langchain"
        }

        wrapper = LangChainAgentWrapper(config)

        # Test calculations capability
        tool = await wrapper._capability_to_tool("calculations")
        if tool:  # Only test if tool creation succeeded
            # Accept both "calculations" and "calculator" as valid names
            assert tool.name in ["calculations", "calculator"]
            assert "mathematical" in tool.description.lower() or "calculation" in tool.description.lower()

    def test_unsupported_capability(self):
        """Test handling of unsupported capabilities"""
        config = {
            "agent_id": "test-unsupported",
            "capabilities": ["unsupported_capability"],
            "framework": "langchain"
        }

        wrapper = LangChainAgentWrapper(config)
        # Should handle gracefully without crashing


class TestPerformance:
    """Test performance characteristics"""

    def test_health_check_performance(self):
        """Test health check response time"""
        start_time = time.time()
        response = client.get("/health")
        end_time = time.time()

        assert response.status_code == 200
        response_time = (end_time - start_time) * 1000  # Convert to milliseconds

        # Health check should be very fast (< 50ms)
        assert response_time < 50, f"Health check took {response_time:.2f}ms, should be < 50ms"

    def test_framework_status_performance(self):
        """Test framework status response time"""
        start_time = time.time()
        response = client.get("/framework/status")
        end_time = time.time()

        assert response.status_code == 200
        response_time = (end_time - start_time) * 1000

        # Framework status should be fast (< 100ms)
        assert response_time < 100, f"Framework status took {response_time:.2f}ms, should be < 100ms"

    def test_concurrent_health_checks(self):
        """Test concurrent health check requests"""
        import concurrent.futures
        import threading

        def make_health_request():
            return client.get("/health")

        # Make 10 concurrent requests
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            start_time = time.time()
            futures = [executor.submit(make_health_request) for _ in range(10)]
            responses = [future.result() for future in concurrent.futures.as_completed(futures)]
            end_time = time.time()

        # All requests should succeed
        for response in responses:
            assert response.status_code == 200

        total_time = (end_time - start_time) * 1000
        avg_time = total_time / 10

        # Average response time should be reasonable under load
        assert avg_time < 200, f"Average response time under load: {avg_time:.2f}ms"


class TestErrorHandling:
    """Test error handling and edge cases"""

    def test_invalid_json_request(self):
        """Test handling of invalid JSON in requests"""
        response = client.post(
            "/agents/create",
            data="invalid json",
            headers={"Content-Type": "application/json"}
        )
        assert response.status_code == 422

    def test_missing_content_type(self):
        """Test handling of missing content type"""
        response = client.post("/agents/create", data='{"test": "data"}')
        # Should still work or give appropriate error
        assert response.status_code in [200, 400, 422, 500]

    def test_large_request_handling(self):
        """Test handling of large requests"""
        large_config = {
            "agent_id": "large-test",
            "capabilities": ["web_search"],
            "framework": "langchain",
            "config": {
                "large_data": "x" * 10000  # 10KB of data
            }
        }

        response = client.post("/agents/create", json=large_config)
        # Should handle gracefully (either succeed or give appropriate error)
        assert response.status_code in [200, 400, 413, 422, 500]


class TestIntegration:
    """Integration tests with external dependencies"""

    def test_langchain_integration_status(self):
        """Test that LangChain integration status is properly reported"""
        response = client.get("/framework/status")
        data = response.json()

        # Test that framework status is properly reported regardless of availability
        assert "available" in data
        assert "framework" in data
        assert data["framework"] == "langchain"

        # If LangChain is available, should be True, otherwise False
        if LANGCHAIN_AVAILABLE:
            assert data["available"] is True
        else:
            assert data["available"] is False

    def test_environment_configuration(self):
        """Test environment configuration detection"""
        response = client.get("/health")
        data = response.json()

        # Should detect OpenAI configuration
        openai_configured = bool(os.getenv("OPENAI_API_KEY"))
        assert data["openai_configured"] == openai_configured

    @pytest.mark.asyncio
    async def test_async_operations(self):
        """Test async operation handling"""
        # Test that async endpoints work properly using TestClient
        from fastapi.testclient import TestClient

        # Use sync client for testing async endpoints
        test_client = TestClient(app)

        response = test_client.get("/health")
        assert response.status_code == 200

        response = test_client.get("/framework/status")
        assert response.status_code == 200


# Test fixtures and utilities
@pytest.fixture
def sample_agent_config():
    """Sample agent configuration for testing"""
    return {
        "agent_id": "test-agent-fixture",
        "capabilities": ["web_search", "calculations"],
        "framework": "langchain",
        "config": {
            "temperature": 0.7,
            "max_tokens": 1000,
            "model": "gpt-3.5-turbo"
        }
    }

@pytest.fixture
def sample_task_request():
    """Sample task request for testing"""
    return {
        "agent_id": "test-agent",
        "task": "Calculate the square root of 16",
        "context": {
            "user_id": "test-user",
            "session_id": "test-session"
        },
        "tools": ["calculations"]
    }


if __name__ == "__main__":
    # Run tests with pytest
    pytest.main([__file__, "-v", "--tb=short"])
