"""
Universal Memory Interface for AgentOS using mem0
Week 4: Advanced Memory System Implementation

This module provides a unified interface for memory operations across all AI frameworks,
using mem0 as the core memory engine for intelligent memory management.
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
except ImportError:
    # Fallback if mem0 is not installed
    Memory = None

from redis import Redis


class MemoryType(Enum):
    """Types of memory in the AgentOS system"""
    WORKING = "working"
    EPISODIC = "episodic"
    SEMANTIC = "semantic"
    PROCEDURAL = "procedural"


class FrameworkType(Enum):
    """Supported AI frameworks"""
    LANGCHAIN = "langchain"
    SWARMS = "swarms"
    CREWAI = "crewai"
    AUTOGEN = "autogen"
    UNIVERSAL = "universal"


@dataclass
class MemoryEntry:
    """Universal memory entry structure"""
    id: str
    content: str
    memory_type: MemoryType
    framework: FrameworkType
    concepts: List[str]
    importance: float
    embedding: Optional[List[float]] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None

    def __post_init__(self):
        if self.created_at is None:
            self.created_at = datetime.now()
        if self.updated_at is None:
            self.updated_at = datetime.now()


@dataclass
class ConsolidationResult:
    """Result of memory consolidation process"""
    consolidation_id: str
    framework: FrameworkType
    episodic_count: int
    semantic_count: int
    consolidation_score: float
    patterns_found: List[str]
    new_memories_created: int
    started_at: datetime
    completed_at: Optional[datetime] = None


class UniversalMemory:
    """
    Universal Memory Interface for AgentOS

    Provides unified memory operations across all AI frameworks with support for:
    - Semantic memory storage and retrieval
    - Memory consolidation (episodic to semantic)
    - Cross-framework memory sharing
    - Strategic forgetting based on importance and access patterns
    """

    def __init__(self,
                 api_base_url: str = "http://localhost:8000",
                 redis_host: str = "localhost",
                 redis_port: int = 6379,
                 redis_db: int = 0):
        """
        Initialize Universal Memory Interface

        Args:
            api_base_url: Base URL for AgentOS API
            redis_host: Redis host for caching
            redis_port: Redis port
            redis_db: Redis database number
        """
        self.api_base_url = api_base_url.rstrip('/')
        self.redis = Redis(host=redis_host, port=redis_port, db=redis_db, decode_responses=True)
        self.logger = logging.getLogger(__name__)

        # Framework-specific adapters (will be initialized lazily)
        self._framework_adapters = {}

        # Memory operation statistics
        self.stats = {
            "memories_stored": 0,
            "memories_retrieved": 0,
            "consolidations_performed": 0,
            "cache_hits": 0,
            "cache_misses": 0
        }

    async def store_memory(self,
                          content: str,
                          memory_type: MemoryType,
                          framework: FrameworkType,
                          concepts: List[str] = None,
                          importance: float = 0.5,
                          metadata: Dict[str, Any] = None) -> str:
        """
        Store a memory entry in the universal memory system

        Args:
            content: The memory content
            memory_type: Type of memory (working, episodic, semantic)
            framework: AI framework this memory belongs to
            concepts: List of concept tags
            importance: Importance score (0.0 to 1.0)
            metadata: Additional metadata

        Returns:
            Memory ID of the stored memory
        """
        if concepts is None:
            concepts = []
        if metadata is None:
            metadata = {}

        # Create memory entry
        memory_entry = MemoryEntry(
            id="",  # Will be assigned by API
            content=content,
            memory_type=memory_type,
            framework=framework,
            concepts=concepts,
            importance=importance,
            metadata=metadata
        )

        try:
            # Store based on memory type
            if memory_type == MemoryType.SEMANTIC:
                memory_id = await self._store_semantic_memory(memory_entry)
            elif memory_type == MemoryType.WORKING:
                memory_id = await self._store_working_memory(memory_entry)
            else:
                memory_id = await self._store_episodic_memory(memory_entry)

            # Update statistics
            self.stats["memories_stored"] += 1

            # Cache the memory for quick access
            cache_key = f"memory:{memory_id}"
            memory_entry.id = memory_id
            await self._cache_memory(cache_key, memory_entry)

            self.logger.info(f"Stored {memory_type.value} memory {memory_id} for {framework.value}")
            return memory_id

        except Exception as e:
            self.logger.error(f"Failed to store memory: {e}")
            raise

    async def retrieve_memory(self,
                             query: str,
                             framework: FrameworkType = None,
                             memory_type: MemoryType = None,
                             limit: int = 10,
                             similarity_threshold: float = 0.7) -> List[MemoryEntry]:
        """
        Retrieve memories using semantic search

        Args:
            query: Search query
            framework: Filter by specific framework (optional)
            memory_type: Filter by memory type (optional)
            limit: Maximum number of results
            similarity_threshold: Minimum similarity score

        Returns:
            List of matching memory entries
        """
        try:
            # Check cache first
            cache_key = f"search:{hash(query)}:{framework}:{memory_type}:{limit}"
            cached_result = await self._get_cached_search(cache_key)
            if cached_result:
                self.stats["cache_hits"] += 1
                return cached_result

            self.stats["cache_misses"] += 1

            # Perform semantic search via API
            search_params = {
                "query": query,
                "limit": limit,
                "threshold": similarity_threshold
            }

            if framework:
                search_params["framework"] = framework.value

            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.api_base_url}/api/v1/memory/semantic/search",
                    json=search_params
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        memories = [self._dict_to_memory_entry(mem) for mem in data.get("memories", [])]

                        # Cache the result
                        await self._cache_search_result(cache_key, memories)

                        self.stats["memories_retrieved"] += len(memories)
                        return memories
                    else:
                        raise Exception(f"API error: {response.status}")

        except Exception as e:
            self.logger.error(f"Failed to retrieve memories: {e}")
            raise

    async def consolidate_memories(self,
                                  framework: FrameworkType,
                                  time_window: timedelta = timedelta(hours=24)) -> ConsolidationResult:
        """
        Consolidate episodic memories into semantic knowledge

        Args:
            framework: Framework to consolidate memories for
            time_window: Time window for episodic memories to consolidate

        Returns:
            Consolidation result with statistics
        """
        try:
            consolidation_params = {
                "framework": framework.value,
                "time_window_hours": time_window.total_seconds() / 3600
            }

            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.api_base_url}/api/v1/memory/consolidation/trigger",
                    json=consolidation_params
                ) as response:
                    if response.status == 200:
                        data = await response.json()

                        result = ConsolidationResult(
                            consolidation_id=data["consolidation_id"],
                            framework=framework,
                            episodic_count=data["episodic_count"],
                            semantic_count=data["semantic_count"],
                            consolidation_score=data["consolidation_score"],
                            patterns_found=data.get("patterns_found", []),
                            new_memories_created=data.get("new_memories_created", 0),
                            started_at=datetime.fromisoformat(data["started_at"])
                        )

                        self.stats["consolidations_performed"] += 1
                        self.logger.info(f"Consolidated {result.episodic_count} episodic memories for {framework.value}")

                        return result
                    else:
                        raise Exception(f"Consolidation API error: {response.status}")

        except Exception as e:
            self.logger.error(f"Failed to consolidate memories: {e}")
            raise

    async def get_framework_memory(self, framework: FrameworkType) -> Dict[str, Any]:
        """
        Get framework-specific memory statistics and recent memories

        Args:
            framework: Framework to get memory for

        Returns:
            Framework memory information
        """
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(
                    f"{self.api_base_url}/api/v1/memory/frameworks/{framework.value}"
                ) as response:
                    if response.status == 200:
                        return await response.json()
                    else:
                        raise Exception(f"Framework memory API error: {response.status}")

        except Exception as e:
            self.logger.error(f"Failed to get framework memory: {e}")
            raise

    # ===================================
    # PRIVATE HELPER METHODS
    # ===================================

    async def _store_semantic_memory(self, memory_entry: MemoryEntry) -> str:
        """Store semantic memory via API"""
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.api_base_url}/api/v1/memory/semantic/store",
                json={
                    "content": memory_entry.content,
                    "concepts": memory_entry.concepts,
                    "framework": memory_entry.framework.value,
                    "source_type": memory_entry.metadata.get("source_type", "user_input"),
                    "importance": memory_entry.importance
                }
            ) as response:
                if response.status == 201:
                    data = await response.json()
                    return data["memory_id"]
                else:
                    raise Exception(f"Semantic memory API error: {response.status}")

    async def _store_working_memory(self, memory_entry: MemoryEntry) -> str:
        """Store working memory in Redis"""
        memory_id = f"working_{int(time.time() * 1000)}"
        cache_key = f"working_memory:{memory_entry.framework.value}:{memory_id}"

        memory_data = asdict(memory_entry)
        memory_data["id"] = memory_id

        self.redis.setex(cache_key, 3600, json.dumps(memory_data, default=str))  # 1 hour TTL
        return memory_id

    async def _store_episodic_memory(self, memory_entry: MemoryEntry) -> str:
        """Store episodic memory in database"""
        # For now, store as semantic memory with episodic type
        # In production, would have separate episodic storage
        return await self._store_semantic_memory(memory_entry)

    async def _cache_memory(self, cache_key: str, memory_entry: MemoryEntry):
        """Cache memory entry in Redis"""
        memory_data = asdict(memory_entry)
        self.redis.setex(cache_key, 300, json.dumps(memory_data, default=str))  # 5 min TTL

    async def _get_cached_search(self, cache_key: str) -> Optional[List[MemoryEntry]]:
        """Get cached search results"""
        cached_data = self.redis.get(cache_key)
        if cached_data:
            try:
                data = json.loads(cached_data)
                return [self._dict_to_memory_entry(mem) for mem in data]
            except:
                return None
        return None

    async def _cache_search_result(self, cache_key: str, memories: List[MemoryEntry]):
        """Cache search results"""
        memory_data = [asdict(mem) for mem in memories]
        self.redis.setex(cache_key, 60, json.dumps(memory_data, default=str))  # 1 min TTL

    def _dict_to_memory_entry(self, data: Dict[str, Any]) -> MemoryEntry:
        """Convert dictionary to MemoryEntry"""
        return MemoryEntry(
            id=data.get("id", ""),
            content=data.get("content", ""),
            memory_type=MemoryType(data.get("memory_type", "semantic")),
            framework=FrameworkType(data.get("framework", "universal")),
            concepts=data.get("concepts", []),
            importance=data.get("importance", 0.5),
            embedding=data.get("embedding"),
            metadata=data.get("metadata", {}),
            created_at=datetime.fromisoformat(data["created_at"]) if data.get("created_at") else datetime.now(),
            updated_at=datetime.fromisoformat(data["updated_at"]) if data.get("updated_at") else datetime.now()
        )

    def get_statistics(self) -> Dict[str, Any]:
        """Get memory operation statistics"""
        return {
            **self.stats,
            "cache_hit_rate": self.stats["cache_hits"] / max(self.stats["cache_hits"] + self.stats["cache_misses"], 1),
            "total_operations": sum(self.stats.values())
        }
