"""
Unit tests for Framework Memory Adapters
Week 4: Advanced Memory System Testing

This module provides comprehensive unit tests for framework-specific
memory adapters.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock, patch
from datetime import datetime

import sys
import os
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'memory'))

from framework_adapters import (
    BaseMemoryAdapter,
    LangChainMemoryAdapter,
    SwarmsMemoryAdapter,
    CrewAIMemoryAdapter,
    AutoGenMemoryAdapter,
    create_memory_adapter
)
from mem0_memory_engine import Mem0MemoryEngine, FrameworkType


class TestBaseMemoryAdapter:
    """Test suite for BaseMemoryAdapter abstract class"""

    @pytest.fixture
    def mock_memory_engine(self):
        """Create mock memory engine"""
        engine = Mock(spec=Mem0MemoryEngine)
        engine.store_memory = AsyncMock(return_value="test_memory_id")
        engine.retrieve_memories = AsyncMock(return_value=[])
        engine.get_framework_memories = AsyncMock(return_value=[])
        return engine

    def test_base_adapter_initialization(self, mock_memory_engine):
        """Test BaseMemoryAdapter initialization"""
        # Cannot instantiate abstract class directly
        with pytest.raises(TypeError):
            BaseMemoryAdapter(mock_memory_engine, FrameworkType.LANGCHAIN)


class TestLangChainMemoryAdapter:
    """Test suite for LangChainMemoryAdapter"""

    @pytest.fixture
    def mock_memory_engine(self):
        """Create mock memory engine"""
        engine = Mock(spec=Mem0MemoryEngine)
        engine.store_memory = AsyncMock(return_value="langchain_memory_123")
        engine.retrieve_memories = AsyncMock(return_value=[
            {
                "content": "LangChain conversation about AI",
                "metadata": {"framework": "langchain", "timestamp": "2024-12-27T10:00:00"},
                "score": 0.85
            }
        ])
        engine.get_framework_memories = AsyncMock(return_value=[
            {
                "content": "Previous LangChain interaction",
                "metadata": {"conversation_id": "test_agent", "timestamp": "2024-12-27T09:00:00"}
            }
        ])
        return engine

    @pytest.fixture
    def langchain_adapter(self, mock_memory_engine):
        """Create LangChainMemoryAdapter instance"""
        return LangChainMemoryAdapter(mock_memory_engine)

    @pytest.mark.asyncio
    async def test_store_conversation(self, langchain_adapter, mock_memory_engine):
        """Test storing LangChain conversation"""
        messages = [
            {"role": "user", "content": "What is machine learning?"},
            {"role": "assistant", "content": "Machine learning is a subset of AI..."},
            {"role": "user", "content": "Can you give me examples?"}
        ]

        memory_id = await langchain_adapter.store_conversation(
            user_id="test_user",
            messages=messages,
            agent_id="langchain_agent_1"
        )

        assert memory_id == "langchain_memory_123"

        # Verify store_memory was called with correct parameters
        mock_memory_engine.store_memory.assert_called_once()
        call_args = mock_memory_engine.store_memory.call_args
        assert call_args[1]["framework"] == FrameworkType.LANGCHAIN
        assert call_args[1]["user_id"] == "test_user"
        assert call_args[1]["agent_id"] == "langchain_agent_1"
        assert "user: What is machine learning?" in call_args[1]["content"]

    @pytest.mark.asyncio
    async def test_retrieve_context(self, langchain_adapter, mock_memory_engine):
        """Test retrieving LangChain context"""
        context = await langchain_adapter.retrieve_context(
            user_id="test_user",
            query="machine learning examples",
            limit=5
        )

        assert len(context) == 1
        assert context[0]["content"] == "LangChain conversation about AI"
        assert context[0]["relevance_score"] == 0.85
        assert "timestamp" in context[0]

        # Verify retrieve_memories was called
        mock_memory_engine.retrieve_memories.assert_called_once_with(
            query="machine learning examples",
            user_id="test_user",
            framework=FrameworkType.LANGCHAIN,
            limit=5
        )

    @pytest.mark.asyncio
    async def test_get_agent_memory(self, langchain_adapter, mock_memory_engine):
        """Test getting LangChain agent memory"""
        agent_memory = await langchain_adapter.get_agent_memory(
            user_id="test_user",
            agent_id="test_agent"
        )

        assert agent_memory["agent_id"] == "test_agent"
        assert agent_memory["framework"] == "langchain"
        assert agent_memory["memory_count"] == 1
        assert len(agent_memory["memories"]) == 1

        # Verify get_framework_memories was called
        mock_memory_engine.get_framework_memories.assert_called_once_with(
            framework=FrameworkType.LANGCHAIN,
            user_id="test_user",
            limit=50
        )

    @pytest.mark.asyncio
    async def test_empty_messages_handling(self, langchain_adapter):
        """Test handling empty messages list"""
        memory_id = await langchain_adapter.store_conversation(
            user_id="test_user",
            messages=[],
            agent_id="test_agent"
        )

        assert memory_id == "langchain_memory_123"


class TestSwarmsMemoryAdapter:
    """Test suite for SwarmsMemoryAdapter"""

    @pytest.fixture
    def mock_memory_engine(self):
        """Create mock memory engine"""
        engine = Mock(spec=Mem0MemoryEngine)
        engine.store_memory = AsyncMock(return_value="swarms_memory_456")
        engine.retrieve_memories = AsyncMock(return_value=[
            {
                "content": "Swarm collaboration with 3 agents: Agent coordination | Task distribution | Result aggregation",
                "metadata": {"framework": "swarms", "agent_count": 3},
                "score": 0.78
            }
        ])
        engine.get_framework_memories = AsyncMock(return_value=[
            {
                "content": "Previous swarm interaction",
                "metadata": {"swarm_id": "test_swarm", "agent_count": 2}
            }
        ])
        return engine

    @pytest.fixture
    def swarms_adapter(self, mock_memory_engine):
        """Create SwarmsMemoryAdapter instance"""
        return SwarmsMemoryAdapter(mock_memory_engine)

    @pytest.mark.asyncio
    async def test_store_conversation(self, swarms_adapter, mock_memory_engine):
        """Test storing Swarms interaction"""
        messages = [
            {"agent_id": "agent_1", "content": "Coordinating task distribution"},
            {"agent_id": "agent_2", "content": "Processing assigned subtask"},
            {"agent_id": "agent_3", "content": "Aggregating results"}
        ]

        memory_id = await swarms_adapter.store_conversation(
            user_id="test_user",
            messages=messages,
            agent_id="swarm_123"
        )

        assert memory_id == "swarms_memory_456"

        # Verify store_memory was called with swarm-specific metadata
        call_args = mock_memory_engine.store_memory.call_args
        assert call_args[1]["framework"] == FrameworkType.SWARMS
        assert call_args[1]["metadata"]["type"] == "swarm_interaction"
        assert call_args[1]["metadata"]["agent_count"] == 3
        assert call_args[1]["metadata"]["interaction_type"] == "collaborative"

    @pytest.mark.asyncio
    async def test_retrieve_context(self, swarms_adapter, mock_memory_engine):
        """Test retrieving Swarms context"""
        context = await swarms_adapter.retrieve_context(
            user_id="test_user",
            query="agent coordination",
            limit=3
        )

        assert len(context) == 1
        assert "interaction_summary" in context[0]
        assert "collaboration_score" in context[0]
        assert "agents_involved" in context[0]
        assert context[0]["agents_involved"] == 3

    @pytest.mark.asyncio
    async def test_get_agent_memory(self, swarms_adapter, mock_memory_engine):
        """Test getting Swarms agent memory"""
        swarm_memory = await swarms_adapter.get_agent_memory(
            user_id="test_user",
            agent_id="test_swarm"
        )

        assert swarm_memory["swarm_id"] == "test_swarm"
        assert swarm_memory["framework"] == "swarms"
        assert "collaboration_patterns" in swarm_memory
        assert swarm_memory["collaboration_patterns"]["collaboration_intensity"] in ["low", "medium", "high"]

    def test_create_swarm_summary(self, swarms_adapter):
        """Test creating swarm interaction summary"""
        messages = [
            {"agent_id": "agent_1", "content": "Task coordination message"},
            {"agent_id": "agent_2", "content": "Processing update"}
        ]

        summary = swarms_adapter._create_swarm_summary(messages)

        assert "Swarm collaboration with 2 agents" in summary
        assert "Task coordination message" in summary

    def test_analyze_collaboration_patterns(self, swarms_adapter):
        """Test analyzing collaboration patterns"""
        memories = [
            {"metadata": {"agent_count": 3}},
            {"metadata": {"agent_count": 2}},
            {"metadata": {"agent_count": 4}}
        ]

        patterns = swarms_adapter._analyze_collaboration_patterns(memories)

        assert patterns["average_agents_per_interaction"] == 3.0
        assert patterns["total_interactions"] == 3
        assert patterns["collaboration_intensity"] == "medium"  # 3.0 average is medium (>1 but <=3)


class TestCrewAIMemoryAdapter:
    """Test suite for CrewAIMemoryAdapter"""

    @pytest.fixture
    def mock_memory_engine(self):
        """Create mock memory engine"""
        engine = Mock(spec=Mem0MemoryEngine)
        engine.store_memory = AsyncMock(return_value="crewai_memory_789")
        engine.retrieve_memories = AsyncMock(return_value=[
            {
                "content": "Crew task with roles researcher, writer: Research AI trends | Write comprehensive report",
                "metadata": {"framework": "crewai", "roles_involved": ["researcher", "writer"]},
                "score": 0.82
            }
        ])
        engine.get_framework_memories = AsyncMock(return_value=[
            {
                "content": "Previous crew task",
                "metadata": {"crew_id": "test_crew", "roles_involved": ["analyst", "reviewer"]}
            }
        ])
        return engine

    @pytest.fixture
    def crewai_adapter(self, mock_memory_engine):
        """Create CrewAIMemoryAdapter instance"""
        return CrewAIMemoryAdapter(mock_memory_engine)

    @pytest.mark.asyncio
    async def test_store_conversation(self, crewai_adapter, mock_memory_engine):
        """Test storing CrewAI task execution"""
        messages = [
            {"role": "researcher", "content": "Conducting research on AI trends"},
            {"role": "writer", "content": "Drafting comprehensive report"},
            {"role": "reviewer", "content": "Reviewing and finalizing content"}
        ]

        memory_id = await crewai_adapter.store_conversation(
            user_id="test_user",
            messages=messages,
            agent_id="crew_456"
        )

        assert memory_id == "crewai_memory_789"

        # Verify store_memory was called with crew-specific metadata
        call_args = mock_memory_engine.store_memory.call_args
        assert call_args[1]["framework"] == FrameworkType.CREWAI
        assert call_args[1]["metadata"]["type"] == "crew_task"
        assert "researcher" in call_args[1]["metadata"]["roles_involved"]
        assert "writer" in call_args[1]["metadata"]["roles_involved"]
        assert "reviewer" in call_args[1]["metadata"]["roles_involved"]

    @pytest.mark.asyncio
    async def test_retrieve_context(self, crewai_adapter, mock_memory_engine):
        """Test retrieving CrewAI context"""
        context = await crewai_adapter.retrieve_context(
            user_id="test_user",
            query="research and writing tasks",
            limit=5
        )

        assert len(context) == 1
        assert "task_summary" in context[0]
        assert "crew_metadata" in context[0]
        assert "roles_involved" in context[0]
        assert "researcher" in context[0]["roles_involved"]
        assert "writer" in context[0]["roles_involved"]

    @pytest.mark.asyncio
    async def test_get_agent_memory(self, crewai_adapter, mock_memory_engine):
        """Test getting CrewAI crew memory"""
        crew_memory = await crewai_adapter.get_agent_memory(
            user_id="test_user",
            agent_id="test_crew"
        )

        assert crew_memory["crew_id"] == "test_crew"
        assert crew_memory["framework"] == "crewai"
        assert "role_analysis" in crew_memory
        assert "most_active_roles" in crew_memory["role_analysis"]

    def test_extract_roles(self, crewai_adapter):
        """Test extracting roles from messages"""
        messages = [
            {"role": "researcher", "content": "Research content"},
            {"role": "writer", "content": "Writing content"},
            {"role": "user", "content": "User input"}  # Should be filtered out
        ]

        roles = crewai_adapter._extract_roles(messages)

        assert "researcher" in roles
        assert "writer" in roles
        assert "user" not in roles
        assert len(roles) == 2

    def test_analyze_role_performance(self, crewai_adapter):
        """Test analyzing role performance"""
        memories = [
            {"metadata": {"roles_involved": ["researcher", "writer"]}},
            {"metadata": {"roles_involved": ["researcher", "analyst"]}},
            {"metadata": {"roles_involved": ["writer", "reviewer"]}}
        ]

        analysis = crewai_adapter._analyze_role_performance(memories)

        assert "most_active_roles" in analysis
        assert "total_unique_roles" in analysis
        assert analysis["total_unique_roles"] == 4  # researcher, writer, analyst, reviewer


class TestAutoGenMemoryAdapter:
    """Test suite for AutoGenMemoryAdapter"""

    @pytest.fixture
    def mock_memory_engine(self):
        """Create mock memory engine"""
        engine = Mock(spec=Mem0MemoryEngine)
        engine.store_memory = AsyncMock(return_value="autogen_memory_101")
        engine.retrieve_memories = AsyncMock(return_value=[
            {
                "content": "AutoGen conversation with 3 agent types over 5 turns",
                "metadata": {"framework": "autogen", "agent_types": ["user", "assistant", "code_executor"]},
                "score": 0.75
            }
        ])
        engine.get_framework_memories = AsyncMock(return_value=[
            {
                "content": "Previous AutoGen conversation",
                "metadata": {"conversation_id": "test_conversation", "turn_count": 8}
            }
        ])
        return engine

    @pytest.fixture
    def autogen_adapter(self, mock_memory_engine):
        """Create AutoGenMemoryAdapter instance"""
        return AutoGenMemoryAdapter(mock_memory_engine)

    @pytest.mark.asyncio
    async def test_store_conversation(self, autogen_adapter, mock_memory_engine):
        """Test storing AutoGen conversation"""
        messages = [
            {"role": "user", "content": "Write a Python function"},
            {"role": "assistant", "content": "Here's a Python function..."},
            {"role": "code_executor", "content": "Executing code..."},
            {"role": "assistant", "content": "The code executed successfully"}
        ]

        memory_id = await autogen_adapter.store_conversation(
            user_id="test_user",
            messages=messages,
            agent_id="autogen_conversation_1"
        )

        assert memory_id == "autogen_memory_101"

        # Verify store_memory was called with AutoGen-specific metadata
        call_args = mock_memory_engine.store_memory.call_args
        assert call_args[1]["framework"] == FrameworkType.AUTOGEN
        assert call_args[1]["metadata"]["type"] == "autogen_conversation"
        assert call_args[1]["metadata"]["turn_count"] == 4
        assert "user" in call_args[1]["metadata"]["agent_types"]
        assert "assistant" in call_args[1]["metadata"]["agent_types"]
        assert "code_executor" in call_args[1]["metadata"]["agent_types"]

    @pytest.mark.asyncio
    async def test_retrieve_context(self, autogen_adapter, mock_memory_engine):
        """Test retrieving AutoGen context"""
        context = await autogen_adapter.retrieve_context(
            user_id="test_user",
            query="code generation and execution",
            limit=5
        )

        assert len(context) == 1
        assert "conversation_summary" in context[0]
        assert "autogen_metadata" in context[0]
        assert "agent_types" in context[0]
        assert "code_executor" in context[0]["agent_types"]

    @pytest.mark.asyncio
    async def test_get_agent_memory(self, autogen_adapter, mock_memory_engine):
        """Test getting AutoGen conversation memory"""
        conversation_memory = await autogen_adapter.get_agent_memory(
            user_id="test_user",
            agent_id="test_conversation"
        )

        assert conversation_memory["conversation_id"] == "test_conversation"
        assert conversation_memory["framework"] == "autogen"
        assert "conversation_patterns" in conversation_memory
        assert "conversation_complexity" in conversation_memory["conversation_patterns"]

    def test_analyze_autogen_conversation(self, autogen_adapter):
        """Test analyzing AutoGen conversation"""
        messages = [
            {"role": "user", "content": "User message"},
            {"role": "assistant", "content": "Assistant response"},
            {"role": "code_executor", "content": "Code execution"}
        ]

        analysis = autogen_adapter._analyze_autogen_conversation(messages)

        assert "AutoGen conversation with 3 agent types over 3 turns" in analysis["summary"]
        assert set(analysis["agent_types"]) == {"user", "assistant", "code_executor"}
        assert analysis["turn_count"] == 3

    def test_analyze_conversation_patterns(self, autogen_adapter):
        """Test analyzing conversation patterns"""
        memories = [
            {"metadata": {"turn_count": 10}},
            {"metadata": {"turn_count": 5}},
            {"metadata": {"turn_count": 15}}
        ]

        patterns = autogen_adapter._analyze_conversation_patterns(memories)

        assert patterns["average_turns_per_conversation"] == 10.0
        assert patterns["total_conversations"] == 3
        assert patterns["conversation_complexity"] == "medium"


class TestMemoryAdapterFactory:
    """Test suite for memory adapter factory function"""

    @pytest.fixture
    def mock_memory_engine(self):
        """Create mock memory engine"""
        return Mock(spec=Mem0MemoryEngine)

    def test_create_langchain_adapter(self, mock_memory_engine):
        """Test creating LangChain adapter"""
        adapter = create_memory_adapter(FrameworkType.LANGCHAIN, mock_memory_engine)
        assert isinstance(adapter, LangChainMemoryAdapter)
        assert adapter.framework == FrameworkType.LANGCHAIN

    def test_create_swarms_adapter(self, mock_memory_engine):
        """Test creating Swarms adapter"""
        adapter = create_memory_adapter(FrameworkType.SWARMS, mock_memory_engine)
        assert isinstance(adapter, SwarmsMemoryAdapter)
        assert adapter.framework == FrameworkType.SWARMS

    def test_create_crewai_adapter(self, mock_memory_engine):
        """Test creating CrewAI adapter"""
        adapter = create_memory_adapter(FrameworkType.CREWAI, mock_memory_engine)
        assert isinstance(adapter, CrewAIMemoryAdapter)
        assert adapter.framework == FrameworkType.CREWAI

    def test_create_autogen_adapter(self, mock_memory_engine):
        """Test creating AutoGen adapter"""
        adapter = create_memory_adapter(FrameworkType.AUTOGEN, mock_memory_engine)
        assert isinstance(adapter, AutoGenMemoryAdapter)
        assert adapter.framework == FrameworkType.AUTOGEN

    def test_create_adapter_invalid_framework(self, mock_memory_engine):
        """Test creating adapter with invalid framework"""
        with pytest.raises(ValueError, match="No adapter available for framework"):
            create_memory_adapter(FrameworkType.UNIVERSAL, mock_memory_engine)


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
