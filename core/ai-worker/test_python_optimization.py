#!/usr/bin/env python3
"""
Week 6 Day 3: Python AI Worker Optimization Validation Test
AgentOS Performance Optimization Validation
"""

import asyncio
import json
import os
import sys
import time
from typing import Dict, Any, List

# Add optimization modules to path
sys.path.append(os.path.join(os.path.dirname(__file__), 'optimization'))

def test_basic_optimization():
    """Test basic optimization functionality"""
    print("🔧 Testing Basic Optimization...")
    
    try:
        from performance_optimizer import PerformanceOptimizer, PerformanceConfig
        
        # Create optimizer with basic config
        config = PerformanceConfig(
            max_workers=2,
            thread_pool_size=4,
            memory_limit_mb=256
        )
        
        optimizer = PerformanceOptimizer(config)
        
        # Test framework execution
        start_time = time.time()
        result = optimizer.optimize_framework_execution(
            "langchain",
            "Test optimization task",
            context="Basic test"
        )
        execution_time = time.time() - start_time
        
        # Get metrics
        metrics = optimizer.get_performance_metrics()
        
        # Cleanup
        optimizer.shutdown()
        
        print(f"✅ Basic optimization test passed in {execution_time:.3f}s")
        print(f"📊 Requests processed: {metrics.get('requests_processed', 0)}")
        
        return True
        
    except Exception as e:
        print(f"❌ Basic optimization test failed: {e}")
        return False

def test_memory_optimization():
    """Test memory optimization functionality"""
    print("🧠 Testing Memory Optimization...")
    
    try:
        from memory_optimizer import OptimizedMemoryEngine, MemoryConfig
        
        # Create memory engine with basic config
        config = MemoryConfig(
            max_memory_mb=128,
            cache_ttl_seconds=300,
            local_cache_size=100
        )
        
        memory_engine = OptimizedMemoryEngine(config)
        
        # Test memory operations
        start_time = time.time()
        
        # Store memories
        memory_ids = []
        for i in range(5):
            memory_id = memory_engine.store_memory(
                f"Test memory content {i}",
                {"test_id": i, "category": "validation"},
                "test_agent"
            )
            memory_ids.append(memory_id)
        
        # Retrieve memories
        retrieved_count = 0
        for memory_id in memory_ids:
            memory = memory_engine.retrieve_memory(memory_id)
            if memory:
                retrieved_count += 1
        
        # Search memories
        search_results = memory_engine.search_memories("test memory", limit=3)
        
        execution_time = time.time() - start_time
        
        # Get stats
        stats = memory_engine.get_memory_stats()
        
        print(f"✅ Memory optimization test passed in {execution_time:.3f}s")
        print(f"📊 Memories stored: {len(memory_ids)}, retrieved: {retrieved_count}")
        print(f"🔍 Search results: {len(search_results)}")
        print(f"💾 L1 cache size: {stats.get('l1_cache_size', 0)}")
        
        return True
        
    except Exception as e:
        print(f"❌ Memory optimization test failed: {e}")
        return False

def test_framework_optimization():
    """Test framework optimization functionality"""
    print("🤖 Testing Framework Optimization...")
    
    try:
        from framework_optimizer import FrameworkOptimizerManager, FrameworkConfig
        
        # Create framework manager with basic config
        config = FrameworkConfig(
            max_concurrent_requests=3,
            request_timeout=5,
            cache_enabled=True
        )
        
        manager = FrameworkOptimizerManager(config)
        
        # Test single task execution
        start_time = time.time()
        result = manager.execute_task(
            "langchain",
            "Test framework optimization"
        )
        single_execution_time = time.time() - start_time
        
        # Test batch execution
        tasks = [
            {"framework": "langchain", "task": f"Batch task {i}", "kwargs": {}}
            for i in range(3)
        ]
        
        start_time = time.time()
        batch_results = asyncio.run(manager.execute_batch_tasks(tasks))
        batch_execution_time = time.time() - start_time
        
        # Get metrics
        metrics = manager.get_all_metrics()
        
        print(f"✅ Framework optimization test passed")
        print(f"⏱️  Single execution: {single_execution_time:.3f}s")
        print(f"⚡ Batch execution: {batch_execution_time:.3f}s")
        print(f"📊 Total requests: {metrics['global']['total_requests']}")
        
        return True
        
    except Exception as e:
        print(f"❌ Framework optimization test failed: {e}")
        return False

async def test_async_optimization():
    """Test async optimization functionality"""
    print("⚡ Testing Async Optimization...")
    
    try:
        from performance_optimizer import PerformanceOptimizer, PerformanceConfig
        
        # Create optimizer for async testing
        config = PerformanceConfig(
            max_workers=4,
            async_batch_size=3
        )
        
        optimizer = PerformanceOptimizer(config)
        
        # Create async tasks
        tasks = [
            {
                "framework": "langchain",
                "task": f"Async task {i}",
                "kwargs": {"priority": i % 2}
            }
            for i in range(6)
        ]
        
        # Execute async batch
        start_time = time.time()
        results = await optimizer.optimize_async_execution(tasks)
        execution_time = time.time() - start_time
        
        # Calculate throughput
        throughput = len(tasks) / execution_time if execution_time > 0 else 0
        
        # Cleanup
        optimizer.shutdown()
        
        print(f"✅ Async optimization test passed in {execution_time:.3f}s")
        print(f"📊 Tasks: {len(tasks)}, Results: {len(results)}")
        print(f"⚡ Throughput: {throughput:.1f} tasks/second")
        
        return True
        
    except Exception as e:
        print(f"❌ Async optimization test failed: {e}")
        return False

def run_validation_tests():
    """Run all validation tests"""
    print("🚀 Starting Python AI Worker Optimization Validation...")
    print("=" * 60)
    
    test_results = {
        "timestamp": time.time(),
        "tests": {}
    }
    
    # Run tests
    test_results["tests"]["basic_optimization"] = test_basic_optimization()
    test_results["tests"]["memory_optimization"] = test_memory_optimization()
    test_results["tests"]["framework_optimization"] = test_framework_optimization()
    test_results["tests"]["async_optimization"] = asyncio.run(test_async_optimization())
    
    # Calculate success rate
    passed_tests = sum(1 for result in test_results["tests"].values() if result)
    total_tests = len(test_results["tests"])
    success_rate = (passed_tests / total_tests) * 100 if total_tests > 0 else 0
    
    print("\n" + "=" * 60)
    print("📊 Validation Results Summary:")
    print("=" * 60)
    
    for test_name, result in test_results["tests"].items():
        status = "✅ PASSED" if result else "❌ FAILED"
        print(f"{test_name}: {status}")
    
    print(f"\n🎯 Overall Success Rate: {success_rate:.1f}% ({passed_tests}/{total_tests})")
    
    if success_rate == 100:
        print("🎉 All Python optimization tests passed successfully!")
        print("✅ Python AI Worker optimization is working correctly")
    elif success_rate >= 75:
        print("⚠️  Most Python optimization tests passed")
        print("🔧 Some optimizations may need attention")
    else:
        print("❌ Multiple Python optimization tests failed")
        print("🚨 Python optimization needs investigation")
    
    # Save results
    try:
        results_dir = "../../performance_results/week6_day3"
        os.makedirs(results_dir, exist_ok=True)
        
        with open(f"{results_dir}/python_validation_results.json", "w") as f:
            json.dump(test_results, f, indent=2, default=str)
        
        print(f"\n📁 Results saved to: {results_dir}/python_validation_results.json")
        
    except Exception as e:
        print(f"⚠️  Could not save results: {e}")
    
    return test_results

if __name__ == "__main__":
    results = run_validation_tests()
    
    # Exit with appropriate code
    passed_tests = sum(1 for result in results["tests"].values() if result)
    total_tests = len(results["tests"])
    
    if passed_tests == total_tests:
        sys.exit(0)  # All tests passed
    else:
        sys.exit(1)  # Some tests failed
