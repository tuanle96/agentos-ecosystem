"""
Framework Memory Adapters for AgentOS
Week 4: Advanced Memory System Implementation

This module provides memory adapters for different AI frameworks,
integrating them with the mem0 memory engine.
"""

import json
import logging
from abc import ABC, abstractmethod
from typing import Dict, List, Optional, Any
from datetime import datetime

from mem0_memory_engine import Mem0MemoryEngine, FrameworkType


class BaseMemoryAdapter(ABC):
    """Base class for framework memory adapters"""
    
    def __init__(self, memory_engine: Mem0MemoryEngine, framework: FrameworkType):
        self.memory_engine = memory_engine
        self.framework = framework
        self.logger = logging.getLogger(f"{__name__}.{framework.value}")
    
    @abstractmethod
    async def store_conversation(self, user_id: str, messages: List[Dict[str, str]], 
                               agent_id: Optional[str] = None) -> str:
        """Store conversation in memory"""
        pass
    
    @abstractmethod
    async def retrieve_context(self, user_id: str, query: str, 
                             limit: int = 5) -> List[Dict[str, Any]]:
        """Retrieve relevant context for query"""
        pass
    
    @abstractmethod
    async def get_agent_memory(self, user_id: str, agent_id: str) -> Dict[str, Any]:
        """Get memory specific to an agent"""
        pass


class LangChainMemoryAdapter(BaseMemoryAdapter):
    """Memory adapter for LangChain framework"""
    
    def __init__(self, memory_engine: Mem0MemoryEngine):
        super().__init__(memory_engine, FrameworkType.LANGCHAIN)
    
    async def store_conversation(self, user_id: str, messages: List[Dict[str, str]], 
                               agent_id: Optional[str] = None) -> str:
        """
        Store LangChain conversation in mem0
        
        Args:
            user_id: User identifier
            messages: List of messages with role and content
            agent_id: LangChain agent identifier
            
        Returns:
            Memory ID
        """
        try:
            # Combine messages into conversation context
            conversation = "\n".join([
                f"{msg.get('role', 'user')}: {msg.get('content', '')}"
                for msg in messages
            ])
            
            metadata = {
                "type": "conversation",
                "message_count": len(messages),
                "agent_type": "langchain_agent",
                "conversation_id": agent_id or f"langchain_{int(datetime.now().timestamp())}"
            }
            
            memory_id = await self.memory_engine.store_memory(
                content=conversation,
                framework=self.framework,
                user_id=user_id,
                agent_id=agent_id,
                metadata=metadata
            )
            
            self.logger.info(f"Stored LangChain conversation: {memory_id}")
            return memory_id
            
        except Exception as e:
            self.logger.error(f"Failed to store LangChain conversation: {e}")
            raise
    
    async def retrieve_context(self, user_id: str, query: str, 
                             limit: int = 5) -> List[Dict[str, Any]]:
        """Retrieve relevant LangChain context"""
        try:
            memories = await self.memory_engine.retrieve_memories(
                query=query,
                user_id=user_id,
                framework=self.framework,
                limit=limit
            )
            
            # Format for LangChain usage
            context = []
            for memory in memories:
                context.append({
                    "content": memory.get("content", ""),
                    "metadata": memory.get("metadata", {}),
                    "relevance_score": memory.get("score", 0.0),
                    "timestamp": memory.get("metadata", {}).get("timestamp")
                })
            
            return context
            
        except Exception as e:
            self.logger.error(f"Failed to retrieve LangChain context: {e}")
            return []
    
    async def get_agent_memory(self, user_id: str, agent_id: str) -> Dict[str, Any]:
        """Get LangChain agent-specific memory"""
        try:
            memories = await self.memory_engine.get_framework_memories(
                framework=self.framework,
                user_id=user_id,
                limit=50
            )
            
            # Filter by agent_id
            agent_memories = [
                mem for mem in memories
                if mem.get("metadata", {}).get("conversation_id") == agent_id
            ]
            
            return {
                "agent_id": agent_id,
                "framework": "langchain",
                "memory_count": len(agent_memories),
                "memories": agent_memories,
                "last_interaction": agent_memories[0].get("metadata", {}).get("timestamp") if agent_memories else None
            }
            
        except Exception as e:
            self.logger.error(f"Failed to get LangChain agent memory: {e}")
            return {"agent_id": agent_id, "framework": "langchain", "memory_count": 0, "memories": []}


class SwarmsMemoryAdapter(BaseMemoryAdapter):
    """Memory adapter for Swarms framework"""
    
    def __init__(self, memory_engine: Mem0MemoryEngine):
        super().__init__(memory_engine, FrameworkType.SWARMS)
    
    async def store_conversation(self, user_id: str, messages: List[Dict[str, str]], 
                               agent_id: Optional[str] = None) -> str:
        """Store Swarms agent interaction"""
        try:
            # Swarms focuses on agent collaboration
            interaction_summary = self._create_swarm_summary(messages)
            
            metadata = {
                "type": "swarm_interaction",
                "agent_count": len(set(msg.get("agent_id") for msg in messages if msg.get("agent_id"))),
                "interaction_type": "collaborative",
                "swarm_id": agent_id or f"swarm_{int(datetime.now().timestamp())}"
            }
            
            memory_id = await self.memory_engine.store_memory(
                content=interaction_summary,
                framework=self.framework,
                user_id=user_id,
                agent_id=agent_id,
                metadata=metadata
            )
            
            self.logger.info(f"Stored Swarms interaction: {memory_id}")
            return memory_id
            
        except Exception as e:
            self.logger.error(f"Failed to store Swarms interaction: {e}")
            raise
    
    async def retrieve_context(self, user_id: str, query: str, 
                             limit: int = 5) -> List[Dict[str, Any]]:
        """Retrieve Swarms collaboration context"""
        try:
            memories = await self.memory_engine.retrieve_memories(
                query=query,
                user_id=user_id,
                framework=self.framework,
                limit=limit
            )
            
            # Format for Swarms usage with collaboration focus
            context = []
            for memory in memories:
                context.append({
                    "interaction_summary": memory.get("content", ""),
                    "swarm_metadata": memory.get("metadata", {}),
                    "collaboration_score": memory.get("score", 0.0),
                    "agents_involved": memory.get("metadata", {}).get("agent_count", 1)
                })
            
            return context
            
        except Exception as e:
            self.logger.error(f"Failed to retrieve Swarms context: {e}")
            return []
    
    async def get_agent_memory(self, user_id: str, agent_id: str) -> Dict[str, Any]:
        """Get Swarms agent memory with collaboration history"""
        try:
            memories = await self.memory_engine.get_framework_memories(
                framework=self.framework,
                user_id=user_id,
                limit=30
            )
            
            # Filter by swarm_id
            swarm_memories = [
                mem for mem in memories
                if mem.get("metadata", {}).get("swarm_id") == agent_id
            ]
            
            return {
                "swarm_id": agent_id,
                "framework": "swarms",
                "interaction_count": len(swarm_memories),
                "memories": swarm_memories,
                "collaboration_patterns": self._analyze_collaboration_patterns(swarm_memories)
            }
            
        except Exception as e:
            self.logger.error(f"Failed to get Swarms agent memory: {e}")
            return {"swarm_id": agent_id, "framework": "swarms", "interaction_count": 0, "memories": []}
    
    def _create_swarm_summary(self, messages: List[Dict[str, str]]) -> str:
        """Create summary of swarm interaction"""
        if not messages:
            return "Empty swarm interaction"
        
        agents = set(msg.get("agent_id", "unknown") for msg in messages)
        content_summary = " | ".join([msg.get("content", "")[:100] for msg in messages[:3]])
        
        return f"Swarm collaboration with {len(agents)} agents: {content_summary}"
    
    def _analyze_collaboration_patterns(self, memories: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Analyze collaboration patterns in swarm memories"""
        if not memories:
            return {"pattern": "no_data"}
        
        total_agents = sum(mem.get("metadata", {}).get("agent_count", 1) for mem in memories)
        avg_collaboration = total_agents / len(memories) if memories else 0
        
        return {
            "average_agents_per_interaction": avg_collaboration,
            "total_interactions": len(memories),
            "collaboration_intensity": "high" if avg_collaboration > 3 else "medium" if avg_collaboration > 1 else "low"
        }


class CrewAIMemoryAdapter(BaseMemoryAdapter):
    """Memory adapter for CrewAI framework"""
    
    def __init__(self, memory_engine: Mem0MemoryEngine):
        super().__init__(memory_engine, FrameworkType.CREWAI)
    
    async def store_conversation(self, user_id: str, messages: List[Dict[str, str]], 
                               agent_id: Optional[str] = None) -> str:
        """Store CrewAI crew task execution"""
        try:
            # CrewAI focuses on role-based task execution
            task_summary = self._create_crew_task_summary(messages)
            
            metadata = {
                "type": "crew_task",
                "roles_involved": self._extract_roles(messages),
                "task_complexity": len(messages),
                "crew_id": agent_id or f"crew_{int(datetime.now().timestamp())}"
            }
            
            memory_id = await self.memory_engine.store_memory(
                content=task_summary,
                framework=self.framework,
                user_id=user_id,
                agent_id=agent_id,
                metadata=metadata
            )
            
            self.logger.info(f"Stored CrewAI task: {memory_id}")
            return memory_id
            
        except Exception as e:
            self.logger.error(f"Failed to store CrewAI task: {e}")
            raise
    
    async def retrieve_context(self, user_id: str, query: str, 
                             limit: int = 5) -> List[Dict[str, Any]]:
        """Retrieve CrewAI task context"""
        try:
            memories = await self.memory_engine.retrieve_memories(
                query=query,
                user_id=user_id,
                framework=self.framework,
                limit=limit
            )
            
            # Format for CrewAI usage with role focus
            context = []
            for memory in memories:
                context.append({
                    "task_summary": memory.get("content", ""),
                    "crew_metadata": memory.get("metadata", {}),
                    "task_relevance": memory.get("score", 0.0),
                    "roles_involved": memory.get("metadata", {}).get("roles_involved", [])
                })
            
            return context
            
        except Exception as e:
            self.logger.error(f"Failed to retrieve CrewAI context: {e}")
            return []
    
    async def get_agent_memory(self, user_id: str, agent_id: str) -> Dict[str, Any]:
        """Get CrewAI crew memory with role analysis"""
        try:
            memories = await self.memory_engine.get_framework_memories(
                framework=self.framework,
                user_id=user_id,
                limit=40
            )
            
            # Filter by crew_id
            crew_memories = [
                mem for mem in memories
                if mem.get("metadata", {}).get("crew_id") == agent_id
            ]
            
            return {
                "crew_id": agent_id,
                "framework": "crewai",
                "task_count": len(crew_memories),
                "memories": crew_memories,
                "role_analysis": self._analyze_role_performance(crew_memories)
            }
            
        except Exception as e:
            self.logger.error(f"Failed to get CrewAI crew memory: {e}")
            return {"crew_id": agent_id, "framework": "crewai", "task_count": 0, "memories": []}
    
    def _create_crew_task_summary(self, messages: List[Dict[str, str]]) -> str:
        """Create summary of crew task execution"""
        if not messages:
            return "Empty crew task"
        
        roles = self._extract_roles(messages)
        task_content = " | ".join([msg.get("content", "")[:80] for msg in messages[:2]])
        
        return f"Crew task with roles {', '.join(roles)}: {task_content}"
    
    def _extract_roles(self, messages: List[Dict[str, str]]) -> List[str]:
        """Extract roles from messages"""
        roles = set()
        for msg in messages:
            role = msg.get("role", "")
            if role and role not in ["user", "assistant", "system"]:
                roles.add(role)
        return list(roles) or ["default_agent"]
    
    def _analyze_role_performance(self, memories: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Analyze role performance in crew tasks"""
        if not memories:
            return {"analysis": "no_data"}
        
        all_roles = []
        for mem in memories:
            roles = mem.get("metadata", {}).get("roles_involved", [])
            all_roles.extend(roles)
        
        role_frequency = {}
        for role in all_roles:
            role_frequency[role] = role_frequency.get(role, 0) + 1
        
        return {
            "most_active_roles": sorted(role_frequency.items(), key=lambda x: x[1], reverse=True)[:3],
            "total_unique_roles": len(role_frequency),
            "role_distribution": role_frequency
        }


class AutoGenMemoryAdapter(BaseMemoryAdapter):
    """Memory adapter for AutoGen framework"""
    
    def __init__(self, memory_engine: Mem0MemoryEngine):
        super().__init__(memory_engine, FrameworkType.AUTOGEN)
    
    async def store_conversation(self, user_id: str, messages: List[Dict[str, str]], 
                               agent_id: Optional[str] = None) -> str:
        """Store AutoGen multi-agent conversation"""
        try:
            # AutoGen focuses on multi-agent conversations
            conversation_analysis = self._analyze_autogen_conversation(messages)
            
            metadata = {
                "type": "autogen_conversation",
                "turn_count": len(messages),
                "agent_types": conversation_analysis["agent_types"],
                "conversation_id": agent_id or f"autogen_{int(datetime.now().timestamp())}"
            }
            
            memory_id = await self.memory_engine.store_memory(
                content=conversation_analysis["summary"],
                framework=self.framework,
                user_id=user_id,
                agent_id=agent_id,
                metadata=metadata
            )
            
            self.logger.info(f"Stored AutoGen conversation: {memory_id}")
            return memory_id
            
        except Exception as e:
            self.logger.error(f"Failed to store AutoGen conversation: {e}")
            raise
    
    async def retrieve_context(self, user_id: str, query: str, 
                             limit: int = 5) -> List[Dict[str, Any]]:
        """Retrieve AutoGen conversation context"""
        try:
            memories = await self.memory_engine.retrieve_memories(
                query=query,
                user_id=user_id,
                framework=self.framework,
                limit=limit
            )
            
            # Format for AutoGen usage
            context = []
            for memory in memories:
                context.append({
                    "conversation_summary": memory.get("content", ""),
                    "autogen_metadata": memory.get("metadata", {}),
                    "conversation_relevance": memory.get("score", 0.0),
                    "agent_types": memory.get("metadata", {}).get("agent_types", [])
                })
            
            return context
            
        except Exception as e:
            self.logger.error(f"Failed to retrieve AutoGen context: {e}")
            return []
    
    async def get_agent_memory(self, user_id: str, agent_id: str) -> Dict[str, Any]:
        """Get AutoGen conversation memory"""
        try:
            memories = await self.memory_engine.get_framework_memories(
                framework=self.framework,
                user_id=user_id,
                limit=35
            )
            
            # Filter by conversation_id
            conversation_memories = [
                mem for mem in memories
                if mem.get("metadata", {}).get("conversation_id") == agent_id
            ]
            
            return {
                "conversation_id": agent_id,
                "framework": "autogen",
                "conversation_count": len(conversation_memories),
                "memories": conversation_memories,
                "conversation_patterns": self._analyze_conversation_patterns(conversation_memories)
            }
            
        except Exception as e:
            self.logger.error(f"Failed to get AutoGen conversation memory: {e}")
            return {"conversation_id": agent_id, "framework": "autogen", "conversation_count": 0, "memories": []}
    
    def _analyze_autogen_conversation(self, messages: List[Dict[str, str]]) -> Dict[str, Any]:
        """Analyze AutoGen conversation"""
        if not messages:
            return {"summary": "Empty conversation", "agent_types": []}
        
        agent_types = list(set(msg.get("role", "user") for msg in messages))
        summary = f"AutoGen conversation with {len(agent_types)} agent types over {len(messages)} turns"
        
        return {
            "summary": summary,
            "agent_types": agent_types,
            "turn_count": len(messages)
        }
    
    def _analyze_conversation_patterns(self, memories: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Analyze conversation patterns"""
        if not memories:
            return {"pattern": "no_data"}
        
        total_turns = sum(mem.get("metadata", {}).get("turn_count", 0) for mem in memories)
        avg_turns = total_turns / len(memories) if memories else 0
        
        return {
            "average_turns_per_conversation": avg_turns,
            "total_conversations": len(memories),
            "conversation_complexity": "high" if avg_turns > 10 else "medium" if avg_turns > 5 else "low"
        }


# Factory function to create appropriate adapter
def create_memory_adapter(framework: FrameworkType, memory_engine: Mem0MemoryEngine) -> BaseMemoryAdapter:
    """Create appropriate memory adapter for framework"""
    adapters = {
        FrameworkType.LANGCHAIN: LangChainMemoryAdapter,
        FrameworkType.SWARMS: SwarmsMemoryAdapter,
        FrameworkType.CREWAI: CrewAIMemoryAdapter,
        FrameworkType.AUTOGEN: AutoGenMemoryAdapter
    }
    
    adapter_class = adapters.get(framework)
    if not adapter_class:
        raise ValueError(f"No adapter available for framework: {framework}")
    
    return adapter_class(memory_engine)
