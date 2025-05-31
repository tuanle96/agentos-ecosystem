#!/usr/bin/env python3
"""
Week 6 Day 3: Python AI Worker Performance Optimizer
AgentOS Performance Optimization Implementation
"""

import asyncio
import concurrent.futures
import functools
import gc
import logging
import multiprocessing
import os
import psutil
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Callable, Union
import weakref

# Performance monitoring
import cProfile
import pstats
from memory_profiler import profile
import tracemalloc

# Add the parent directory to Python path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from frameworks.orchestrator import FrameworkOrchestrator
from memory.mem0_memory_engine import Mem0MemoryEngine


@dataclass
class PerformanceConfig:
    """Configuration for performance optimization"""
    max_workers: int = multiprocessing.cpu_count() * 2
    thread_pool_size: int = 20
    process_pool_size: int = 4
    memory_limit_mb: int = 2048
    gc_threshold: int = 1000
    cache_size: int = 1000
    async_batch_size: int = 10
    framework_timeout: int = 30
    memory_cleanup_interval: int = 300  # 5 minutes


class PerformanceOptimizer:
    """Advanced performance optimizer for Python AI Worker"""
    
    def __init__(self, config: Optional[PerformanceConfig] = None):
        self.config = config or PerformanceConfig()
        self.logger = self._setup_logging()
        
        # Performance tracking
        self.metrics = {
            "requests_processed": 0,
            "total_processing_time": 0.0,
            "memory_usage_peak": 0,
            "cache_hits": 0,
            "cache_misses": 0,
            "framework_switches": 0,
            "errors": 0
        }
        
        # Thread and process pools
        self.thread_pool = ThreadPoolExecutor(max_workers=self.config.thread_pool_size)
        self.process_pool = ProcessPoolExecutor(max_workers=self.config.process_pool_size)
        
        # Caching system
        self.cache = {}
        self.cache_access_times = {}
        self.cache_lock = threading.RLock()
        
        # Framework instances with connection pooling
        self.framework_pool = {}
        self.framework_lock = threading.RLock()
        
        # Memory management
        self.memory_monitor = MemoryMonitor(self.config.memory_limit_mb)
        
        # Start background tasks
        self._start_background_tasks()
        
        self.logger.info(f"Performance optimizer initialized with {self.config.max_workers} workers")
    
    def _setup_logging(self) -> logging.Logger:
        """Setup optimized logging"""
        logger = logging.getLogger("performance_optimizer")
        logger.setLevel(logging.INFO)
        
        if not logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            )
            handler.setFormatter(formatter)
            logger.addHandler(handler)
        
        return logger
    
    def _start_background_tasks(self):
        """Start background optimization tasks"""
        # Memory cleanup task
        cleanup_thread = threading.Thread(
            target=self._memory_cleanup_loop,
            daemon=True
        )
        cleanup_thread.start()
        
        # Cache cleanup task
        cache_cleanup_thread = threading.Thread(
            target=self._cache_cleanup_loop,
            daemon=True
        )
        cache_cleanup_thread.start()
        
        # Metrics collection task
        metrics_thread = threading.Thread(
            target=self._metrics_collection_loop,
            daemon=True
        )
        metrics_thread.start()
    
    def optimize_framework_execution(self, framework_name: str, task: str, **kwargs) -> Any:
        """Optimized framework execution with caching and pooling"""
        start_time = time.time()
        
        try:
            # Check cache first
            cache_key = self._generate_cache_key(framework_name, task, kwargs)
            cached_result = self._get_from_cache(cache_key)
            
            if cached_result is not None:
                self.metrics["cache_hits"] += 1
                self.logger.debug(f"Cache hit for {framework_name}: {task[:50]}...")
                return cached_result
            
            self.metrics["cache_misses"] += 1
            
            # Get or create framework instance
            framework = self._get_framework_instance(framework_name)
            
            # Execute with timeout and optimization
            result = self._execute_with_optimization(framework, task, **kwargs)
            
            # Cache the result
            self._store_in_cache(cache_key, result)
            
            # Update metrics
            processing_time = time.time() - start_time
            self.metrics["requests_processed"] += 1
            self.metrics["total_processing_time"] += processing_time
            
            self.logger.debug(f"Framework {framework_name} executed in {processing_time:.3f}s")
            return result
            
        except Exception as e:
            self.metrics["errors"] += 1
            self.logger.error(f"Framework execution error: {e}")
            raise
    
    async def optimize_async_execution(self, tasks: List[Dict[str, Any]]) -> List[Any]:
        """Optimized async execution for multiple tasks"""
        if not tasks:
            return []
        
        # Batch tasks for optimal processing
        batches = [
            tasks[i:i + self.config.async_batch_size]
            for i in range(0, len(tasks), self.config.async_batch_size)
        ]
        
        results = []
        for batch in batches:
            batch_results = await self._process_batch_async(batch)
            results.extend(batch_results)
        
        return results
    
    async def _process_batch_async(self, batch: List[Dict[str, Any]]) -> List[Any]:
        """Process a batch of tasks asynchronously"""
        loop = asyncio.get_event_loop()
        
        # Create tasks for concurrent execution
        tasks = []
        for task_data in batch:
            framework_name = task_data.get("framework", "langchain")
            task_content = task_data.get("task", "")
            kwargs = task_data.get("kwargs", {})
            
            # Run in thread pool to avoid blocking
            task = loop.run_in_executor(
                self.thread_pool,
                self.optimize_framework_execution,
                framework_name,
                task_content,
                **kwargs
            )
            tasks.append(task)
        
        # Wait for all tasks to complete
        results = await asyncio.gather(*tasks, return_exceptions=True)
        return results
    
    def _get_framework_instance(self, framework_name: str):
        """Get or create optimized framework instance"""
        with self.framework_lock:
            if framework_name not in self.framework_pool:
                # Create new framework instance with optimization
                if framework_name == "langchain":
                    framework = self._create_optimized_langchain()
                elif framework_name == "swarms":
                    framework = self._create_optimized_swarms()
                elif framework_name == "crewai":
                    framework = self._create_optimized_crewai()
                elif framework_name == "autogen":
                    framework = self._create_optimized_autogen()
                else:
                    raise ValueError(f"Unknown framework: {framework_name}")
                
                self.framework_pool[framework_name] = framework
                self.metrics["framework_switches"] += 1
                self.logger.info(f"Created optimized {framework_name} instance")
            
            return self.framework_pool[framework_name]
    
    def _create_optimized_langchain(self):
        """Create optimized LangChain instance"""
        try:
            from langchain.llms import OpenAI
            from langchain.chains import LLMChain
            from langchain.prompts import PromptTemplate
            
            # Optimized LangChain configuration
            llm = OpenAI(
                temperature=0.7,
                max_tokens=1000,
                request_timeout=self.config.framework_timeout,
                max_retries=2
            )
            
            prompt = PromptTemplate(
                input_variables=["task"],
                template="Execute the following task efficiently: {task}"
            )
            
            chain = LLMChain(llm=llm, prompt=prompt)
            return chain
            
        except ImportError:
            self.logger.warning("LangChain not available, using mock implementation")
            return MockFramework("langchain")
    
    def _create_optimized_swarms(self):
        """Create optimized Swarms instance"""
        try:
            # Swarms optimization would go here
            # For now, return mock implementation
            return MockFramework("swarms")
        except ImportError:
            return MockFramework("swarms")
    
    def _create_optimized_crewai(self):
        """Create optimized CrewAI instance"""
        try:
            # CrewAI optimization would go here
            # For now, return mock implementation
            return MockFramework("crewai")
        except ImportError:
            return MockFramework("crewai")
    
    def _create_optimized_autogen(self):
        """Create optimized AutoGen instance"""
        try:
            # AutoGen optimization would go here
            # For now, return mock implementation
            return MockFramework("autogen")
        except ImportError:
            return MockFramework("autogen")
    
    def _execute_with_optimization(self, framework, task: str, **kwargs) -> Any:
        """Execute framework with performance optimization"""
        # Memory optimization
        gc.collect()
        
        # Execute with timeout
        try:
            if hasattr(framework, 'run'):
                result = framework.run(task, **kwargs)
            elif hasattr(framework, 'execute'):
                result = framework.execute(task, **kwargs)
            else:
                # Mock execution
                result = f"Optimized execution of: {task[:100]}..."
            
            return result
            
        except Exception as e:
            self.logger.error(f"Framework execution failed: {e}")
            raise
    
    def _generate_cache_key(self, framework_name: str, task: str, kwargs: Dict) -> str:
        """Generate cache key for task"""
        import hashlib
        
        # Create deterministic key
        key_data = f"{framework_name}:{task}:{sorted(kwargs.items())}"
        return hashlib.md5(key_data.encode()).hexdigest()
    
    def _get_from_cache(self, cache_key: str) -> Optional[Any]:
        """Get result from cache"""
        with self.cache_lock:
            if cache_key in self.cache:
                # Update access time
                self.cache_access_times[cache_key] = time.time()
                return self.cache[cache_key]
            return None
    
    def _store_in_cache(self, cache_key: str, result: Any):
        """Store result in cache"""
        with self.cache_lock:
            # Check cache size limit
            if len(self.cache) >= self.config.cache_size:
                self._evict_cache_entries()
            
            self.cache[cache_key] = result
            self.cache_access_times[cache_key] = time.time()
    
    def _evict_cache_entries(self):
        """Evict old cache entries (LRU)"""
        if not self.cache_access_times:
            return
        
        # Remove 20% of oldest entries
        sorted_entries = sorted(
            self.cache_access_times.items(),
            key=lambda x: x[1]
        )
        
        entries_to_remove = len(sorted_entries) // 5
        for cache_key, _ in sorted_entries[:entries_to_remove]:
            self.cache.pop(cache_key, None)
            self.cache_access_times.pop(cache_key, None)
    
    def _memory_cleanup_loop(self):
        """Background memory cleanup loop"""
        while True:
            try:
                time.sleep(self.config.memory_cleanup_interval)
                
                # Force garbage collection
                collected = gc.collect()
                
                # Check memory usage
                process = psutil.Process()
                memory_mb = process.memory_info().rss / 1024 / 1024
                
                if memory_mb > self.config.memory_limit_mb * 0.8:
                    self.logger.warning(f"High memory usage: {memory_mb:.1f}MB")
                    self._aggressive_cleanup()
                
                self.logger.debug(f"Memory cleanup: collected {collected} objects, using {memory_mb:.1f}MB")
                
            except Exception as e:
                self.logger.error(f"Memory cleanup error: {e}")
    
    def _cache_cleanup_loop(self):
        """Background cache cleanup loop"""
        while True:
            try:
                time.sleep(60)  # Check every minute
                
                current_time = time.time()
                expired_keys = []
                
                with self.cache_lock:
                    for cache_key, access_time in self.cache_access_times.items():
                        # Remove entries older than 30 minutes
                        if current_time - access_time > 1800:
                            expired_keys.append(cache_key)
                    
                    for cache_key in expired_keys:
                        self.cache.pop(cache_key, None)
                        self.cache_access_times.pop(cache_key, None)
                
                if expired_keys:
                    self.logger.debug(f"Cache cleanup: removed {len(expired_keys)} expired entries")
                
            except Exception as e:
                self.logger.error(f"Cache cleanup error: {e}")
    
    def _metrics_collection_loop(self):
        """Background metrics collection loop"""
        while True:
            try:
                time.sleep(30)  # Collect every 30 seconds
                
                # Update memory metrics
                process = psutil.Process()
                memory_mb = process.memory_info().rss / 1024 / 1024
                self.metrics["memory_usage_peak"] = max(
                    self.metrics["memory_usage_peak"],
                    memory_mb
                )
                
                # Log performance summary
                if self.metrics["requests_processed"] > 0:
                    avg_time = self.metrics["total_processing_time"] / self.metrics["requests_processed"]
                    cache_hit_rate = self.metrics["cache_hits"] / (
                        self.metrics["cache_hits"] + self.metrics["cache_misses"]
                    ) * 100 if (self.metrics["cache_hits"] + self.metrics["cache_misses"]) > 0 else 0
                    
                    self.logger.info(
                        f"Performance: {self.metrics['requests_processed']} requests, "
                        f"{avg_time:.3f}s avg, {cache_hit_rate:.1f}% cache hit rate"
                    )
                
            except Exception as e:
                self.logger.error(f"Metrics collection error: {e}")
    
    def _aggressive_cleanup(self):
        """Aggressive memory cleanup when usage is high"""
        # Clear half of the cache
        with self.cache_lock:
            cache_keys = list(self.cache.keys())
            keys_to_remove = cache_keys[:len(cache_keys) // 2]
            
            for cache_key in keys_to_remove:
                self.cache.pop(cache_key, None)
                self.cache_access_times.pop(cache_key, None)
        
        # Force garbage collection
        gc.collect()
        
        self.logger.info("Aggressive memory cleanup completed")
    
    def get_performance_metrics(self) -> Dict[str, Any]:
        """Get current performance metrics"""
        process = psutil.Process()
        memory_mb = process.memory_info().rss / 1024 / 1024
        
        metrics = self.metrics.copy()
        metrics.update({
            "current_memory_mb": memory_mb,
            "cache_size": len(self.cache),
            "thread_pool_active": self.thread_pool._threads,
            "framework_instances": len(self.framework_pool)
        })
        
        return metrics
    
    def shutdown(self):
        """Shutdown optimizer and cleanup resources"""
        self.logger.info("Shutting down performance optimizer...")
        
        # Shutdown thread pools
        self.thread_pool.shutdown(wait=True)
        self.process_pool.shutdown(wait=True)
        
        # Clear caches
        with self.cache_lock:
            self.cache.clear()
            self.cache_access_times.clear()
        
        # Clear framework pool
        with self.framework_lock:
            self.framework_pool.clear()
        
        self.logger.info("Performance optimizer shutdown completed")


class MemoryMonitor:
    """Memory usage monitoring and optimization"""
    
    def __init__(self, limit_mb: int):
        self.limit_mb = limit_mb
        self.logger = logging.getLogger("memory_monitor")
    
    def check_memory_usage(self) -> bool:
        """Check if memory usage is within limits"""
        process = psutil.Process()
        memory_mb = process.memory_info().rss / 1024 / 1024
        
        if memory_mb > self.limit_mb:
            self.logger.warning(f"Memory limit exceeded: {memory_mb:.1f}MB > {self.limit_mb}MB")
            return False
        
        return True


class MockFramework:
    """Mock framework for testing and fallback"""
    
    def __init__(self, name: str):
        self.name = name
    
    def run(self, task: str, **kwargs) -> str:
        """Mock run method"""
        return f"Mock {self.name} execution: {task[:100]}..."
    
    def execute(self, task: str, **kwargs) -> str:
        """Mock execute method"""
        return self.run(task, **kwargs)


# Global optimizer instance
_optimizer_instance = None


def get_optimizer() -> PerformanceOptimizer:
    """Get global optimizer instance"""
    global _optimizer_instance
    if _optimizer_instance is None:
        _optimizer_instance = PerformanceOptimizer()
    return _optimizer_instance


def optimize_framework_call(framework_name: str, task: str, **kwargs) -> Any:
    """Optimized framework call function"""
    optimizer = get_optimizer()
    return optimizer.optimize_framework_execution(framework_name, task, **kwargs)


async def optimize_async_calls(tasks: List[Dict[str, Any]]) -> List[Any]:
    """Optimized async framework calls"""
    optimizer = get_optimizer()
    return await optimizer.optimize_async_execution(tasks)
