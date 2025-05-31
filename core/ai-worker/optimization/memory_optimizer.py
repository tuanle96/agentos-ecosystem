#!/usr/bin/env python3
"""
Week 6 Day 3: Memory System Optimization
AgentOS Performance Optimization Implementation
"""

import asyncio
import json
import logging
import threading
import time
import weakref
from collections import defaultdict, OrderedDict
from dataclasses import dataclass, asdict
from typing import Any, Dict, List, Optional, Set, Tuple, Union
import hashlib
import pickle
import zlib

# Memory profiling
import psutil
import gc
from memory_profiler import profile

# Redis optimization
import redis
from redis.connection import ConnectionPool


@dataclass
class MemoryConfig:
    """Configuration for memory optimization"""
    max_memory_mb: int = 1024
    cache_ttl_seconds: int = 3600
    compression_threshold: int = 1024  # bytes
    batch_size: int = 100
    cleanup_interval: int = 300  # 5 minutes
    redis_pool_size: int = 50
    local_cache_size: int = 10000
    memory_pressure_threshold: float = 0.8


class OptimizedMemoryEngine:
    """Optimized memory engine with multi-level caching and compression"""
    
    def __init__(self, config: Optional[MemoryConfig] = None):
        self.config = config or MemoryConfig()
        self.logger = self._setup_logging()
        
        # Multi-level cache
        self.l1_cache = OrderedDict()  # In-memory cache
        self.l2_cache = {}  # Compressed cache
        self.cache_stats = defaultdict(int)
        self.cache_lock = threading.RLock()
        
        # Redis connection with optimization
        self.redis_pool = None
        self.redis_client = None
        self._setup_redis()
        
        # Memory tracking
        self.memory_tracker = MemoryTracker(self.config.max_memory_mb)
        
        # Background tasks
        self._start_background_tasks()
        
        self.logger.info("Optimized memory engine initialized")
    
    def _setup_logging(self) -> logging.Logger:
        """Setup optimized logging"""
        logger = logging.getLogger("memory_optimizer")
        logger.setLevel(logging.INFO)
        
        if not logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            )
            handler.setFormatter(formatter)
            logger.addHandler(handler)
        
        return logger
    
    def _setup_redis(self):
        """Setup optimized Redis connection"""
        try:
            # Create optimized connection pool
            self.redis_pool = ConnectionPool(
                host='localhost',
                port=6379,
                db=0,
                max_connections=self.config.redis_pool_size,
                retry_on_timeout=True,
                socket_keepalive=True,
                socket_keepalive_options={},
                health_check_interval=30
            )
            
            self.redis_client = redis.Redis(
                connection_pool=self.redis_pool,
                decode_responses=False  # Keep binary for compression
            )
            
            # Test connection
            self.redis_client.ping()
            self.logger.info(f"Redis connected with pool size {self.config.redis_pool_size}")
            
        except Exception as e:
            self.logger.warning(f"Redis connection failed: {e}, using local cache only")
            self.redis_client = None
    
    def _start_background_tasks(self):
        """Start background optimization tasks"""
        # Memory cleanup task
        cleanup_thread = threading.Thread(
            target=self._memory_cleanup_loop,
            daemon=True
        )
        cleanup_thread.start()
        
        # Cache optimization task
        cache_thread = threading.Thread(
            target=self._cache_optimization_loop,
            daemon=True
        )
        cache_thread.start()
    
    def store_memory(self, content: str, metadata: Optional[Dict] = None, 
                    agent_id: Optional[str] = None) -> str:
        """Store memory with optimization"""
        start_time = time.time()
        
        try:
            # Generate memory ID
            memory_id = self._generate_memory_id(content, metadata)
            
            # Prepare memory object
            memory_obj = {
                "id": memory_id,
                "content": content,
                "metadata": metadata or {},
                "agent_id": agent_id,
                "timestamp": time.time(),
                "access_count": 0
            }
            
            # Store in multi-level cache
            self._store_in_cache(memory_id, memory_obj)
            
            # Store in Redis if available
            if self.redis_client:
                self._store_in_redis(memory_id, memory_obj)
            
            # Update stats
            self.cache_stats["stores"] += 1
            store_time = time.time() - start_time
            self.cache_stats["total_store_time"] += store_time
            
            self.logger.debug(f"Memory stored: {memory_id} in {store_time:.3f}s")
            return memory_id
            
        except Exception as e:
            self.logger.error(f"Memory store error: {e}")
            raise
    
    def retrieve_memory(self, memory_id: str) -> Optional[Dict]:
        """Retrieve memory with optimization"""
        start_time = time.time()
        
        try:
            # Check L1 cache first
            memory_obj = self._get_from_l1_cache(memory_id)
            if memory_obj:
                self.cache_stats["l1_hits"] += 1
                self._update_access_stats(memory_obj)
                return memory_obj
            
            # Check L2 cache
            memory_obj = self._get_from_l2_cache(memory_id)
            if memory_obj:
                self.cache_stats["l2_hits"] += 1
                # Promote to L1 cache
                self._store_in_l1_cache(memory_id, memory_obj)
                self._update_access_stats(memory_obj)
                return memory_obj
            
            # Check Redis
            if self.redis_client:
                memory_obj = self._get_from_redis(memory_id)
                if memory_obj:
                    self.cache_stats["redis_hits"] += 1
                    # Store in local caches
                    self._store_in_cache(memory_id, memory_obj)
                    self._update_access_stats(memory_obj)
                    return memory_obj
            
            # Memory not found
            self.cache_stats["misses"] += 1
            return None
            
        except Exception as e:
            self.logger.error(f"Memory retrieve error: {e}")
            return None
        
        finally:
            retrieve_time = time.time() - start_time
            self.cache_stats["total_retrieve_time"] += retrieve_time
    
    def search_memories(self, query: str, limit: int = 10, 
                       agent_id: Optional[str] = None) -> List[Dict]:
        """Search memories with optimization"""
        start_time = time.time()
        
        try:
            results = []
            query_lower = query.lower()
            
            # Search in L1 cache first (most recent/frequent)
            l1_results = self._search_in_cache(self.l1_cache, query_lower, agent_id)
            results.extend(l1_results)
            
            # Search in L2 cache if needed
            if len(results) < limit:
                l2_results = self._search_in_cache(self.l2_cache, query_lower, agent_id)
                results.extend(l2_results)
            
            # Search in Redis if needed and available
            if len(results) < limit and self.redis_client:
                redis_results = self._search_in_redis(query_lower, agent_id, limit - len(results))
                results.extend(redis_results)
            
            # Sort by relevance and access count
            results = self._rank_search_results(results, query_lower)
            
            # Update access stats for returned results
            for result in results[:limit]:
                self._update_access_stats(result)
            
            search_time = time.time() - start_time
            self.cache_stats["searches"] += 1
            self.cache_stats["total_search_time"] += search_time
            
            self.logger.debug(f"Memory search: {len(results)} results in {search_time:.3f}s")
            return results[:limit]
            
        except Exception as e:
            self.logger.error(f"Memory search error: {e}")
            return []
    
    def consolidate_memories(self, agent_id: Optional[str] = None) -> Dict[str, Any]:
        """Consolidate memories with optimization"""
        start_time = time.time()
        
        try:
            # Get all memories for agent
            memories = self._get_all_memories(agent_id)
            
            if not memories:
                return {"status": "no_memories", "count": 0}
            
            # Group similar memories
            memory_groups = self._group_similar_memories(memories)
            
            # Consolidate each group
            consolidated_count = 0
            for group in memory_groups:
                if len(group) > 1:
                    consolidated_memory = self._consolidate_memory_group(group)
                    if consolidated_memory:
                        # Store consolidated memory
                        self.store_memory(
                            consolidated_memory["content"],
                            consolidated_memory["metadata"],
                            agent_id
                        )
                        
                        # Remove original memories
                        for memory in group:
                            self._remove_memory(memory["id"])
                        
                        consolidated_count += len(group) - 1
            
            consolidation_time = time.time() - start_time
            self.cache_stats["consolidations"] += 1
            
            result = {
                "status": "success",
                "memories_processed": len(memories),
                "groups_found": len(memory_groups),
                "memories_consolidated": consolidated_count,
                "time_seconds": consolidation_time
            }
            
            self.logger.info(f"Memory consolidation: {consolidated_count} memories consolidated")
            return result
            
        except Exception as e:
            self.logger.error(f"Memory consolidation error: {e}")
            return {"status": "error", "error": str(e)}
    
    def _store_in_cache(self, memory_id: str, memory_obj: Dict):
        """Store in multi-level cache"""
        with self.cache_lock:
            # Store in L1 cache
            self._store_in_l1_cache(memory_id, memory_obj)
            
            # Store compressed version in L2 cache
            compressed_obj = self._compress_memory(memory_obj)
            self.l2_cache[memory_id] = compressed_obj
    
    def _store_in_l1_cache(self, memory_id: str, memory_obj: Dict):
        """Store in L1 cache with LRU eviction"""
        with self.cache_lock:
            # Check cache size limit
            if len(self.l1_cache) >= self.config.local_cache_size:
                # Remove oldest item (LRU)
                oldest_key = next(iter(self.l1_cache))
                del self.l1_cache[oldest_key]
            
            # Store new item
            self.l1_cache[memory_id] = memory_obj
            # Move to end (most recently used)
            self.l1_cache.move_to_end(memory_id)
    
    def _get_from_l1_cache(self, memory_id: str) -> Optional[Dict]:
        """Get from L1 cache"""
        with self.cache_lock:
            if memory_id in self.l1_cache:
                # Move to end (most recently used)
                self.l1_cache.move_to_end(memory_id)
                return self.l1_cache[memory_id].copy()
            return None
    
    def _get_from_l2_cache(self, memory_id: str) -> Optional[Dict]:
        """Get from L2 cache (compressed)"""
        with self.cache_lock:
            if memory_id in self.l2_cache:
                compressed_obj = self.l2_cache[memory_id]
                return self._decompress_memory(compressed_obj)
            return None
    
    def _store_in_redis(self, memory_id: str, memory_obj: Dict):
        """Store in Redis with compression"""
        try:
            # Serialize and compress
            serialized = json.dumps(memory_obj, default=str)
            
            if len(serialized) > self.config.compression_threshold:
                compressed = zlib.compress(serialized.encode())
                self.redis_client.setex(
                    f"memory:{memory_id}",
                    self.config.cache_ttl_seconds,
                    compressed
                )
                # Mark as compressed
                self.redis_client.setex(
                    f"compressed:{memory_id}",
                    self.config.cache_ttl_seconds,
                    "1"
                )
            else:
                self.redis_client.setex(
                    f"memory:{memory_id}",
                    self.config.cache_ttl_seconds,
                    serialized
                )
                
        except Exception as e:
            self.logger.error(f"Redis store error: {e}")
    
    def _get_from_redis(self, memory_id: str) -> Optional[Dict]:
        """Get from Redis with decompression"""
        try:
            # Check if compressed
            is_compressed = self.redis_client.get(f"compressed:{memory_id}")
            data = self.redis_client.get(f"memory:{memory_id}")
            
            if data is None:
                return None
            
            if is_compressed:
                # Decompress
                decompressed = zlib.decompress(data)
                serialized = decompressed.decode()
            else:
                serialized = data.decode() if isinstance(data, bytes) else data
            
            return json.loads(serialized)
            
        except Exception as e:
            self.logger.error(f"Redis get error: {e}")
            return None
    
    def _compress_memory(self, memory_obj: Dict) -> bytes:
        """Compress memory object"""
        try:
            serialized = json.dumps(memory_obj, default=str)
            if len(serialized) > self.config.compression_threshold:
                return zlib.compress(serialized.encode())
            else:
                return serialized.encode()
        except Exception as e:
            self.logger.error(f"Compression error: {e}")
            return json.dumps(memory_obj, default=str).encode()
    
    def _decompress_memory(self, compressed_data: bytes) -> Dict:
        """Decompress memory object"""
        try:
            # Try decompression first
            try:
                decompressed = zlib.decompress(compressed_data)
                return json.loads(decompressed.decode())
            except zlib.error:
                # Not compressed, decode directly
                return json.loads(compressed_data.decode())
        except Exception as e:
            self.logger.error(f"Decompression error: {e}")
            return {}
    
    def _generate_memory_id(self, content: str, metadata: Optional[Dict]) -> str:
        """Generate unique memory ID"""
        content_hash = hashlib.md5(content.encode()).hexdigest()
        metadata_hash = hashlib.md5(str(sorted((metadata or {}).items())).encode()).hexdigest()
        timestamp = str(int(time.time() * 1000))
        return f"{content_hash[:8]}_{metadata_hash[:8]}_{timestamp}"
    
    def _update_access_stats(self, memory_obj: Dict):
        """Update memory access statistics"""
        memory_obj["access_count"] = memory_obj.get("access_count", 0) + 1
        memory_obj["last_accessed"] = time.time()
    
    def _search_in_cache(self, cache: Dict, query: str, agent_id: Optional[str]) -> List[Dict]:
        """Search in cache"""
        results = []
        
        for memory_obj in cache.values():
            if isinstance(memory_obj, bytes):
                memory_obj = self._decompress_memory(memory_obj)
            
            # Filter by agent_id if specified
            if agent_id and memory_obj.get("agent_id") != agent_id:
                continue
            
            # Simple text search
            content = memory_obj.get("content", "").lower()
            if query in content:
                results.append(memory_obj.copy())
        
        return results
    
    def _search_in_redis(self, query: str, agent_id: Optional[str], limit: int) -> List[Dict]:
        """Search in Redis"""
        results = []
        
        try:
            # Get all memory keys
            pattern = "memory:*"
            keys = self.redis_client.keys(pattern)
            
            for key in keys[:limit * 2]:  # Get more than needed for filtering
                memory_obj = self._get_from_redis(key.decode().split(":", 1)[1])
                if memory_obj:
                    # Filter by agent_id if specified
                    if agent_id and memory_obj.get("agent_id") != agent_id:
                        continue
                    
                    # Simple text search
                    content = memory_obj.get("content", "").lower()
                    if query in content:
                        results.append(memory_obj)
                        
                        if len(results) >= limit:
                            break
            
        except Exception as e:
            self.logger.error(f"Redis search error: {e}")
        
        return results
    
    def _rank_search_results(self, results: List[Dict], query: str) -> List[Dict]:
        """Rank search results by relevance"""
        def calculate_score(memory_obj: Dict) -> float:
            content = memory_obj.get("content", "").lower()
            
            # Base score from query matches
            score = content.count(query) * 10
            
            # Boost recent memories
            age_hours = (time.time() - memory_obj.get("timestamp", 0)) / 3600
            recency_score = max(0, 100 - age_hours)
            
            # Boost frequently accessed memories
            access_score = memory_obj.get("access_count", 0) * 5
            
            return score + recency_score + access_score
        
        return sorted(results, key=calculate_score, reverse=True)
    
    def _get_all_memories(self, agent_id: Optional[str]) -> List[Dict]:
        """Get all memories for consolidation"""
        memories = []
        
        # Get from L1 cache
        with self.cache_lock:
            for memory_obj in self.l1_cache.values():
                if not agent_id or memory_obj.get("agent_id") == agent_id:
                    memories.append(memory_obj.copy())
        
        # Get from Redis if available
        if self.redis_client:
            try:
                keys = self.redis_client.keys("memory:*")
                for key in keys:
                    memory_obj = self._get_from_redis(key.decode().split(":", 1)[1])
                    if memory_obj and (not agent_id or memory_obj.get("agent_id") == agent_id):
                        # Avoid duplicates
                        if not any(m["id"] == memory_obj["id"] for m in memories):
                            memories.append(memory_obj)
            except Exception as e:
                self.logger.error(f"Error getting memories from Redis: {e}")
        
        return memories
    
    def _group_similar_memories(self, memories: List[Dict]) -> List[List[Dict]]:
        """Group similar memories for consolidation"""
        # Simple grouping by content similarity
        groups = []
        used_indices = set()
        
        for i, memory1 in enumerate(memories):
            if i in used_indices:
                continue
            
            group = [memory1]
            used_indices.add(i)
            
            for j, memory2 in enumerate(memories[i+1:], i+1):
                if j in used_indices:
                    continue
                
                # Simple similarity check
                content1 = memory1.get("content", "").lower()
                content2 = memory2.get("content", "").lower()
                
                # Check for common words (simple similarity)
                words1 = set(content1.split())
                words2 = set(content2.split())
                
                if len(words1 & words2) / max(len(words1 | words2), 1) > 0.3:
                    group.append(memory2)
                    used_indices.add(j)
            
            groups.append(group)
        
        return groups
    
    def _consolidate_memory_group(self, group: List[Dict]) -> Optional[Dict]:
        """Consolidate a group of similar memories"""
        if len(group) <= 1:
            return None
        
        # Combine content
        combined_content = " ".join([m.get("content", "") for m in group])
        
        # Merge metadata
        combined_metadata = {}
        for memory in group:
            metadata = memory.get("metadata", {})
            for key, value in metadata.items():
                if key not in combined_metadata:
                    combined_metadata[key] = value
                elif isinstance(value, list):
                    if isinstance(combined_metadata[key], list):
                        combined_metadata[key].extend(value)
                    else:
                        combined_metadata[key] = [combined_metadata[key]] + value
        
        # Add consolidation info
        combined_metadata["consolidated_from"] = [m["id"] for m in group]
        combined_metadata["consolidation_timestamp"] = time.time()
        
        return {
            "content": combined_content,
            "metadata": combined_metadata
        }
    
    def _remove_memory(self, memory_id: str):
        """Remove memory from all caches"""
        with self.cache_lock:
            self.l1_cache.pop(memory_id, None)
            self.l2_cache.pop(memory_id, None)
        
        if self.redis_client:
            try:
                self.redis_client.delete(f"memory:{memory_id}")
                self.redis_client.delete(f"compressed:{memory_id}")
            except Exception as e:
                self.logger.error(f"Redis delete error: {e}")
    
    def _memory_cleanup_loop(self):
        """Background memory cleanup loop"""
        while True:
            try:
                time.sleep(self.config.cleanup_interval)
                
                # Check memory pressure
                if self.memory_tracker.is_memory_pressure():
                    self._aggressive_cleanup()
                else:
                    self._routine_cleanup()
                
            except Exception as e:
                self.logger.error(f"Memory cleanup error: {e}")
    
    def _cache_optimization_loop(self):
        """Background cache optimization loop"""
        while True:
            try:
                time.sleep(60)  # Run every minute
                
                # Optimize L1 cache
                self._optimize_l1_cache()
                
                # Cleanup expired Redis keys
                if self.redis_client:
                    self._cleanup_redis_expired()
                
            except Exception as e:
                self.logger.error(f"Cache optimization error: {e}")
    
    def _routine_cleanup(self):
        """Routine memory cleanup"""
        # Force garbage collection
        collected = gc.collect()
        
        # Clean old entries from L2 cache
        current_time = time.time()
        expired_keys = []
        
        with self.cache_lock:
            for key, memory_obj in list(self.l2_cache.items()):
                if isinstance(memory_obj, bytes):
                    memory_obj = self._decompress_memory(memory_obj)
                
                # Remove entries older than cache TTL
                if current_time - memory_obj.get("timestamp", 0) > self.config.cache_ttl_seconds:
                    expired_keys.append(key)
            
            for key in expired_keys:
                del self.l2_cache[key]
        
        self.logger.debug(f"Routine cleanup: {collected} objects, {len(expired_keys)} expired entries")
    
    def _aggressive_cleanup(self):
        """Aggressive cleanup when memory pressure is high"""
        # Clear half of L1 cache (keep most recent)
        with self.cache_lock:
            items_to_remove = len(self.l1_cache) // 2
            for _ in range(items_to_remove):
                if self.l1_cache:
                    self.l1_cache.popitem(last=False)  # Remove oldest
        
        # Clear quarter of L2 cache
        with self.cache_lock:
            items_to_remove = len(self.l2_cache) // 4
            keys_to_remove = list(self.l2_cache.keys())[:items_to_remove]
            for key in keys_to_remove:
                del self.l2_cache[key]
        
        # Force garbage collection
        gc.collect()
        
        self.logger.warning("Aggressive memory cleanup completed")
    
    def _optimize_l1_cache(self):
        """Optimize L1 cache based on access patterns"""
        with self.cache_lock:
            # Sort by access count and recency
            sorted_items = sorted(
                self.l1_cache.items(),
                key=lambda x: (x[1].get("access_count", 0), x[1].get("last_accessed", 0)),
                reverse=True
            )
            
            # Rebuild cache with optimized order
            self.l1_cache.clear()
            for key, value in sorted_items:
                self.l1_cache[key] = value
    
    def _cleanup_redis_expired(self):
        """Cleanup expired Redis keys"""
        try:
            # This would be handled by Redis TTL, but we can do additional cleanup
            pass
        except Exception as e:
            self.logger.error(f"Redis cleanup error: {e}")
    
    def get_memory_stats(self) -> Dict[str, Any]:
        """Get memory system statistics"""
        process = psutil.Process()
        memory_mb = process.memory_info().rss / 1024 / 1024
        
        stats = {
            "current_memory_mb": memory_mb,
            "max_memory_mb": self.config.max_memory_mb,
            "memory_usage_percent": (memory_mb / self.config.max_memory_mb) * 100,
            "l1_cache_size": len(self.l1_cache),
            "l2_cache_size": len(self.l2_cache),
            "cache_stats": dict(self.cache_stats),
            "redis_connected": self.redis_client is not None
        }
        
        if self.redis_client:
            try:
                redis_info = self.redis_client.info("memory")
                stats["redis_memory_mb"] = redis_info.get("used_memory", 0) / 1024 / 1024
            except Exception:
                pass
        
        return stats


class MemoryTracker:
    """Memory usage tracking and alerts"""
    
    def __init__(self, limit_mb: int):
        self.limit_mb = limit_mb
        self.logger = logging.getLogger("memory_tracker")
    
    def is_memory_pressure(self) -> bool:
        """Check if system is under memory pressure"""
        process = psutil.Process()
        memory_mb = process.memory_info().rss / 1024 / 1024
        
        pressure_threshold = self.limit_mb * 0.8
        return memory_mb > pressure_threshold


# Global optimized memory engine instance
_memory_engine_instance = None


def get_optimized_memory_engine() -> OptimizedMemoryEngine:
    """Get global optimized memory engine instance"""
    global _memory_engine_instance
    if _memory_engine_instance is None:
        _memory_engine_instance = OptimizedMemoryEngine()
    return _memory_engine_instance
