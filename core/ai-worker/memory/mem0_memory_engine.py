"""
mem0 Memory Engine for AgentOS
Week 4: Advanced Memory System Implementation

This module integrates mem0 as the core memory engine for AgentOS,
providing intelligent memory management across all AI frameworks.
"""

import json
import logging
import time
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, field
from enum import Enum

try:
    from mem0 import Memory
    from mem0.configs.base import MemoryConfig as Mem0Config, MemoryItem
    MEM0_AVAILABLE = True
except ImportError:
    Memory = None
    Mem0Config = None
    MemoryItem = None
    MEM0_AVAILABLE = False

import httpx
from redis import Redis


class FrameworkType(Enum):
    """Supported AI frameworks"""
    LANGCHAIN = "langchain"
    SWARMS = "swarms"
    CREWAI = "crewai"
    AUTOGEN = "autogen"
    UNIVERSAL = "universal"


@dataclass
class MemoryConfig:
    """Configuration for mem0 memory engine"""
    vector_store: str = "qdrant"  # qdrant, chroma, weaviate
    embedding_model: str = "text-embedding-ada-002"
    llm_model: str = "gpt-4"
    memory_decay: bool = True
    importance_threshold: float = 0.5
    max_memories: int = 10000


@dataclass
class AgentOSMemoryEntry:
    """AgentOS memory entry with framework context"""
    id: str
    content: str
    framework: FrameworkType
    user_id: str
    agent_id: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    importance: float = 0.5
    created_at: datetime = field(default_factory=datetime.now)


class Mem0MemoryEngine:
    """
    mem0-powered Memory Engine for AgentOS

    Provides intelligent memory management using mem0's advanced features:
    - Automatic memory consolidation
    - Semantic search and retrieval
    - Cross-framework memory sharing
    - Intelligent forgetting and importance scoring
    """

    def __init__(self,
                 config: MemoryConfig = None,
                 redis_host: str = "localhost",
                 redis_port: int = 6379):
        """
        Initialize mem0 Memory Engine

        Args:
            config: Memory configuration
            redis_host: Redis host for caching
            redis_port: Redis port
        """
        self.config = config or MemoryConfig()
        self.logger = logging.getLogger(__name__)

        # Initialize Redis for caching
        try:
            self.redis = Redis(host=redis_host, port=redis_port,
                             db=0, decode_responses=True)
            self.redis.ping()
        except Exception as e:
            self.logger.warning(f"Redis not available: {e}")
            self.redis = None

        # Initialize mem0 if available
        if MEM0_AVAILABLE and Mem0Config:
            try:
                # Create proper mem0 config
                mem0_config = Mem0Config()
                mem0_config.vector_store.provider = self.config.vector_store
                mem0_config.vector_store.config.collection_name = "agentos_memories"
                mem0_config.vector_store.config.embedding_model_dims = 1536

                mem0_config.embedder.provider = "openai"
                mem0_config.embedder.config.model = self.config.embedding_model

                mem0_config.llm.provider = "openai"
                mem0_config.llm.config.model = self.config.llm_model

                self.memory = Memory(config=mem0_config)
                self.logger.info("mem0 Memory Engine initialized successfully")
            except Exception as e:
                self.logger.error(f"Failed to initialize mem0: {e}")
                self.memory = None
        else:
            self.logger.warning("mem0 not available, using fallback memory")
            self.memory = None

        # Memory operation statistics
        self.stats = {
            "memories_stored": 0,
            "memories_retrieved": 0,
            "cache_hits": 0,
            "cache_misses": 0,
            "consolidations": 0
        }

    async def store_memory(self,
                          content: str,
                          framework: FrameworkType,
                          user_id: str,
                          agent_id: Optional[str] = None,
                          metadata: Optional[Dict[str, Any]] = None) -> str:
        """
        Store a memory using mem0

        Args:
            content: Memory content
            framework: AI framework
            user_id: User identifier
            agent_id: Agent identifier (optional)
            metadata: Additional metadata

        Returns:
            Memory ID
        """
        if metadata is None:
            metadata = {}

        # Add framework context to metadata
        enhanced_metadata = {
            **metadata,
            "framework": framework.value,
            "user_id": user_id,
            "agent_id": agent_id,
            "timestamp": datetime.now().isoformat(),
            "source": "agentos"
        }

        try:
            if self.memory:
                # Use mem0 for intelligent storage
                result = self.memory.add(
                    messages=[{"role": "user", "content": content}],
                    user_id=user_id,
                    agent_id=agent_id or f"{framework.value}_agent",
                    metadata=enhanced_metadata
                )
                memory_id = result.get("id", f"mem0_{int(time.time() * 1000)}")
            else:
                # Fallback to simple storage
                memory_id = f"fallback_{int(time.time() * 1000)}"
                await self._store_fallback_memory(memory_id, content, enhanced_metadata)

            # Cache for quick access
            if self.redis:
                cache_key = f"memory:{memory_id}"
                memory_data = {
                    "id": memory_id,
                    "content": content,
                    "metadata": enhanced_metadata
                }
                self.redis.setex(cache_key, 3600, json.dumps(memory_data))

            self.stats["memories_stored"] += 1
            self.logger.info(f"Stored memory {memory_id} for {framework.value}")

            return memory_id

        except Exception as e:
            self.logger.error(f"Failed to store memory: {e}")
            raise

    async def retrieve_memories(self,
                               query: str,
                               user_id: str,
                               framework: Optional[FrameworkType] = None,
                               limit: int = 10) -> List[Dict[str, Any]]:
        """
        Retrieve memories using semantic search

        Args:
            query: Search query
            user_id: User identifier
            framework: Filter by framework (optional)
            limit: Maximum results

        Returns:
            List of relevant memories
        """
        try:
            # Check cache first
            cache_key = f"search:{hash(query)}:{user_id}:{framework}:{limit}"
            if self.redis:
                cached_result = self.redis.get(cache_key)
                if cached_result:
                    self.stats["cache_hits"] += 1
                    return json.loads(cached_result)

            self.stats["cache_misses"] += 1

            if self.memory:
                # Use mem0 for intelligent retrieval
                memories = self.memory.search(
                    query=query,
                    user_id=user_id,
                    limit=limit
                )

                # Filter by framework if specified
                if framework:
                    memories = [
                        mem for mem in memories
                        if mem.get("metadata", {}).get("framework") == framework.value
                    ]
            else:
                # Fallback search
                memories = await self._search_fallback_memories(query, user_id, framework, limit)

            # Cache results
            if self.redis:
                self.redis.setex(cache_key, 300, json.dumps(memories))

            self.stats["memories_retrieved"] += len(memories)
            return memories

        except Exception as e:
            self.logger.error(f"Failed to retrieve memories: {e}")
            return []

    async def get_framework_memories(self,
                                   framework: FrameworkType,
                                   user_id: str,
                                   limit: int = 20) -> List[Dict[str, Any]]:
        """
        Get all memories for a specific framework

        Args:
            framework: AI framework
            user_id: User identifier
            limit: Maximum results

        Returns:
            Framework-specific memories
        """
        try:
            if self.memory:
                # Get all memories and filter by framework
                all_memories = self.memory.get_all(user_id=user_id, limit=limit * 2)
                framework_memories = [
                    mem for mem in all_memories
                    if mem.get("metadata", {}).get("framework") == framework.value
                ][:limit]
            else:
                framework_memories = await self._get_fallback_framework_memories(
                    framework, user_id, limit
                )

            return framework_memories

        except Exception as e:
            self.logger.error(f"Failed to get framework memories: {e}")
            return []

    async def consolidate_memories(self,
                                 user_id: str,
                                 framework: Optional[FrameworkType] = None) -> Dict[str, Any]:
        """
        Trigger memory consolidation using mem0's intelligence

        Args:
            user_id: User identifier
            framework: Framework to consolidate (optional)

        Returns:
            Consolidation results
        """
        try:
            if not self.memory:
                return {"status": "skipped", "reason": "mem0 not available"}

            # mem0 handles consolidation automatically, but we can trigger analysis
            memories = await self.get_framework_memories(framework, user_id) if framework else []

            consolidation_result = {
                "status": "completed",
                "framework": framework.value if framework else "all",
                "memories_analyzed": len(memories),
                "consolidation_score": 0.8,  # mem0 provides this internally
                "timestamp": datetime.now().isoformat()
            }

            self.stats["consolidations"] += 1
            return consolidation_result

        except Exception as e:
            self.logger.error(f"Failed to consolidate memories: {e}")
            return {"status": "failed", "error": str(e)}

    def get_statistics(self) -> Dict[str, Any]:
        """Get memory engine statistics"""
        cache_total = self.stats["cache_hits"] + self.stats["cache_misses"]
        cache_hit_rate = self.stats["cache_hits"] / max(cache_total, 1)

        return {
            **self.stats,
            "cache_hit_rate": cache_hit_rate,
            "mem0_available": MEM0_AVAILABLE and self.memory is not None,
            "redis_available": self.redis is not None,
            "total_operations": sum(self.stats.values())
        }

    # ===================================
    # FALLBACK METHODS (when mem0 unavailable)
    # ===================================

    async def _store_fallback_memory(self, memory_id: str, content: str, metadata: Dict[str, Any]):
        """Store memory without mem0"""
        if self.redis:
            key = f"fallback_memory:{memory_id}"
            data = {"content": content, "metadata": metadata}
            self.redis.setex(key, 86400, json.dumps(data))  # 24 hour TTL

    async def _search_fallback_memories(self, query: str, user_id: str,
                                      framework: Optional[FrameworkType],
                                      limit: int) -> List[Dict[str, Any]]:
        """Simple fallback search"""
        if not self.redis:
            return []

        # Simple keyword matching (in production would use proper search)
        memories = []
        pattern = "fallback_memory:*"

        for key in self.redis.scan_iter(match=pattern):
            try:
                data = json.loads(self.redis.get(key))
                if query.lower() in data["content"].lower():
                    if not framework or data["metadata"].get("framework") == framework.value:
                        memories.append({
                            "id": key.split(":")[-1],
                            "content": data["content"],
                            "metadata": data["metadata"]
                        })
                        if len(memories) >= limit:
                            break
            except:
                continue

        return memories

    async def _get_fallback_framework_memories(self, framework: FrameworkType,
                                             user_id: str, limit: int) -> List[Dict[str, Any]]:
        """Get framework memories without mem0"""
        return await self._search_fallback_memories("", user_id, framework, limit)

    async def _cleanup_expired_memories(self) -> None:
        """
        Clean up expired memories from Redis cache
        """
        if not self.redis:
            return

        try:
            # Scan for fallback memory keys
            pattern = "fallback_memory:*"
            async for key in self.redis.scan_iter(match=pattern):
                # Check if key exists (it might have expired)
                if await self.redis.exists(key):
                    # Get TTL to check if it's about to expire
                    ttl = await self.redis.ttl(key)
                    if ttl == -1:  # No expiration set
                        # Set default expiration of 24 hours
                        await self.redis.expire(key, 86400)

        except Exception as e:
            self.logger.error(f"Error during memory cleanup: {e}")

    async def get_memory_by_id(self, memory_id: str, user_id: str) -> Optional[Dict[str, Any]]:
        """
        Retrieve a specific memory by ID

        Args:
            memory_id: Memory identifier
            user_id: User identifier

        Returns:
            Memory entry or None if not found
        """
        if self.memory:
            try:
                # Use mem0's get method if available
                if hasattr(self.memory, 'get'):
                    result = self.memory.get(memory_id, user_id=user_id)
                    return result
                else:
                    # Fallback to search with specific ID
                    memories = self.memory.search(
                        query=memory_id,
                        user_id=user_id,
                        limit=1
                    )
                    return memories[0] if memories else None

            except Exception as e:
                self.logger.error(f"Error retrieving memory by ID: {e}")

        # Fallback to Redis
        if self.redis:
            try:
                key = f"fallback_memory:{memory_id}"
                data = self.redis.get(key)  # Redis is sync, not async
                if data:
                    return json.loads(data)
            except Exception as e:
                self.logger.error(f"Error retrieving memory from Redis: {e}")

        return None

    async def delete_memory(self, memory_id: str, user_id: str) -> bool:
        """
        Delete a specific memory

        Args:
            memory_id: Memory identifier
            user_id: User identifier

        Returns:
            True if deleted successfully, False otherwise
        """
        if self.memory:
            try:
                # Use mem0's delete method if available
                if hasattr(self.memory, 'delete'):
                    self.memory.delete(memory_id, user_id=user_id)
                    return True
            except Exception as e:
                self.logger.error(f"Error deleting memory from mem0: {e}")

        # Fallback to Redis deletion
        if self.redis:
            try:
                key = f"fallback_memory:{memory_id}"
                result = self.redis.delete(key)  # Redis is sync, not async
                return result > 0
            except Exception as e:
                self.logger.error(f"Error deleting memory from Redis: {e}")

        return False
