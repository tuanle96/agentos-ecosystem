#!/usr/bin/env python3
"""
Week 6 Day 1: Python AI Worker Performance Profiler
AgentOS Performance Optimization Implementation
"""

import cProfile
import pstats
import io
import time
import psutil
import os
import sys
import tracemalloc
import asyncio
import json
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, List
import requests
from memory_profiler import profile
import gc

# Add the current directory to Python path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from frameworks.orchestrator import FrameworkOrchestrator
from memory.mem0_memory_engine import Mem0MemoryEngine


class AIWorkerProfiler:
    """Comprehensive performance profiler for Python AI Worker"""
    
    def __init__(self):
        self.results_dir = Path("../../performance_results/week6_day1")
        self.results_dir.mkdir(parents=True, exist_ok=True)
        
        self.orchestrator = None
        self.memory_engine = None
        
        # Performance metrics
        self.metrics = {
            "timestamp": datetime.now().isoformat(),
            "service": "agentos-ai-worker",
            "version": "0.1.0-mvp-week6",
            "python_version": sys.version,
            "process_id": os.getpid(),
            "tests": {}
        }
    
    def setup_components(self):
        """Initialize AI Worker components"""
        try:
            print("🔧 Initializing AI Worker components...")
            
            # Initialize Framework Orchestrator
            self.orchestrator = FrameworkOrchestrator()
            print(f"✅ Framework Orchestrator initialized with {len(self.orchestrator.available_frameworks)} frameworks")
            
            # Initialize Memory Engine
            self.memory_engine = Mem0MemoryEngine()
            print("✅ Memory Engine initialized")
            
            return True
        except Exception as e:
            print(f"❌ Failed to initialize components: {e}")
            return False
    
    def profile_framework_operations(self) -> Dict[str, Any]:
        """Profile framework operations performance"""
        print("🧠 Profiling Framework Operations...")
        
        framework_results = {}
        
        for framework_name in self.orchestrator.available_frameworks:
            print(f"  Testing {framework_name} framework...")
            
            start_time = time.time()
            
            try:
                # Test agent creation
                agent_start = time.time()
                agent = self.orchestrator.create_agent(
                    framework=framework_name,
                    name=f"test_agent_{framework_name}",
                    capabilities=["web_search", "text_processing"]
                )
                agent_creation_time = (time.time() - agent_start) * 1000
                
                # Test agent execution
                exec_start = time.time()
                result = self.orchestrator.execute_agent(
                    agent_id=agent.id if hasattr(agent, 'id') else f"test_{framework_name}",
                    task="Analyze performance metrics for testing",
                    context={"test": "performance_profiling"}
                )
                execution_time = (time.time() - exec_start) * 1000
                
                total_time = (time.time() - start_time) * 1000
                
                framework_results[framework_name] = {
                    "status": "success",
                    "agent_creation_ms": round(agent_creation_time, 2),
                    "execution_ms": round(execution_time, 2),
                    "total_ms": round(total_time, 2),
                    "result_length": len(str(result)) if result else 0
                }
                
                print(f"    ✅ {framework_name}: {total_time:.2f}ms")
                
            except Exception as e:
                framework_results[framework_name] = {
                    "status": "failed",
                    "error": str(e),
                    "total_ms": (time.time() - start_time) * 1000
                }
                print(f"    ❌ {framework_name}: {e}")
        
        return framework_results
    
    def profile_memory_operations(self) -> Dict[str, Any]:
        """Profile memory system performance"""
        print("🧠 Profiling Memory Operations...")
        
        memory_results = {}
        
        try:
            # Test memory storage
            store_start = time.time()
            memory_id = self.memory_engine.store_memory(
                content="Performance testing memory content",
                metadata={"test": "performance", "framework": "all"}
            )
            store_time = (time.time() - store_start) * 1000
            
            # Test memory search
            search_start = time.time()
            search_results = self.memory_engine.search_memory(
                query="performance testing",
                limit=10
            )
            search_time = (time.time() - search_start) * 1000
            
            # Test memory consolidation
            consolidate_start = time.time()
            consolidation_result = self.memory_engine.consolidate_memories()
            consolidate_time = (time.time() - consolidate_start) * 1000
            
            memory_results = {
                "status": "success",
                "store_ms": round(store_time, 2),
                "search_ms": round(search_time, 2),
                "consolidate_ms": round(consolidate_time, 2),
                "search_results_count": len(search_results) if search_results else 0,
                "memory_id": memory_id
            }
            
            print(f"    ✅ Memory operations completed in {store_time + search_time + consolidate_time:.2f}ms")
            
        except Exception as e:
            memory_results = {
                "status": "failed",
                "error": str(e)
            }
            print(f"    ❌ Memory operations failed: {e}")
        
        return memory_results
    
    def profile_system_resources(self) -> Dict[str, Any]:
        """Profile system resource usage"""
        print("💻 Profiling System Resources...")
        
        # Get current process
        process = psutil.Process(os.getpid())
        
        # Memory info
        memory_info = process.memory_info()
        memory_percent = process.memory_percent()
        
        # CPU info
        cpu_percent = process.cpu_percent(interval=1)
        
        # System info
        system_memory = psutil.virtual_memory()
        system_cpu = psutil.cpu_percent(interval=1)
        
        return {
            "process": {
                "memory_rss_mb": round(memory_info.rss / 1024 / 1024, 2),
                "memory_vms_mb": round(memory_info.vms / 1024 / 1024, 2),
                "memory_percent": round(memory_percent, 2),
                "cpu_percent": round(cpu_percent, 2),
                "num_threads": process.num_threads(),
                "num_fds": process.num_fds() if hasattr(process, 'num_fds') else 0
            },
            "system": {
                "memory_total_gb": round(system_memory.total / 1024 / 1024 / 1024, 2),
                "memory_available_gb": round(system_memory.available / 1024 / 1024 / 1024, 2),
                "memory_percent": round(system_memory.percent, 2),
                "cpu_percent": round(system_cpu, 2),
                "cpu_count": psutil.cpu_count()
            }
        }
    
    def run_cpu_profiling(self) -> str:
        """Run CPU profiling and save results"""
        print("⚡ Running CPU Profiling...")
        
        # Create profiler
        profiler = cProfile.Profile()
        
        # Start profiling
        profiler.enable()
        
        try:
            # Run performance-intensive operations
            self.profile_framework_operations()
            self.profile_memory_operations()
            
            # Simulate some additional work
            for i in range(1000):
                data = [x**2 for x in range(100)]
                sum(data)
        
        finally:
            # Stop profiling
            profiler.disable()
        
        # Save results
        cpu_profile_path = self.results_dir / "python_cpu_profile.prof"
        profiler.dump_stats(str(cpu_profile_path))
        
        # Generate text report
        s = io.StringIO()
        ps = pstats.Stats(profiler, stream=s)
        ps.sort_stats('cumulative')
        ps.print_stats(50)  # Top 50 functions
        
        cpu_report_path = self.results_dir / "python_cpu_report.txt"
        with open(cpu_report_path, 'w') as f:
            f.write(s.getvalue())
        
        print(f"    ✅ CPU profile saved to {cpu_profile_path}")
        return str(cpu_profile_path)
    
    def run_memory_profiling(self) -> Dict[str, Any]:
        """Run memory profiling and save results"""
        print("🧠 Running Memory Profiling...")
        
        # Start memory tracing
        tracemalloc.start()
        
        # Get initial memory
        initial_memory = tracemalloc.get_traced_memory()
        
        try:
            # Run memory-intensive operations
            framework_results = self.profile_framework_operations()
            memory_results = self.profile_memory_operations()
            
            # Simulate memory usage
            large_data = []
            for i in range(10000):
                large_data.append(f"Memory test data item {i}" * 10)
            
            # Force garbage collection
            gc.collect()
            
            # Get peak memory
            current_memory, peak_memory = tracemalloc.get_traced_memory()
            
        finally:
            tracemalloc.stop()
        
        memory_stats = {
            "initial_mb": round(initial_memory[0] / 1024 / 1024, 2),
            "current_mb": round(current_memory / 1024 / 1024, 2),
            "peak_mb": round(peak_memory / 1024 / 1024, 2),
            "growth_mb": round((current_memory - initial_memory[0]) / 1024 / 1024, 2)
        }
        
        # Save memory stats
        memory_stats_path = self.results_dir / "python_memory_stats.json"
        with open(memory_stats_path, 'w') as f:
            json.dump(memory_stats, f, indent=2)
        
        print(f"    ✅ Memory stats saved to {memory_stats_path}")
        return memory_stats
    
    def run_comprehensive_profiling(self):
        """Run comprehensive performance profiling"""
        print("🚀 Starting Comprehensive AI Worker Profiling...")
        print("=" * 60)
        
        # Setup components
        if not self.setup_components():
            print("❌ Failed to setup components, aborting profiling")
            return
        
        # Run profiling tests
        start_time = time.time()
        
        # 1. System resources
        self.metrics["tests"]["system_resources"] = self.profile_system_resources()
        
        # 2. Framework operations
        self.metrics["tests"]["framework_operations"] = self.profile_framework_operations()
        
        # 3. Memory operations
        self.metrics["tests"]["memory_operations"] = self.profile_memory_operations()
        
        # 4. CPU profiling
        cpu_profile_path = self.run_cpu_profiling()
        self.metrics["tests"]["cpu_profiling"] = {
            "status": "completed",
            "profile_path": cpu_profile_path
        }
        
        # 5. Memory profiling
        memory_stats = self.run_memory_profiling()
        self.metrics["tests"]["memory_profiling"] = memory_stats
        
        # Calculate total time
        total_time = time.time() - start_time
        self.metrics["total_profiling_time_seconds"] = round(total_time, 2)
        
        # Save comprehensive results
        results_path = self.results_dir / "python_profiling_results.json"
        with open(results_path, 'w') as f:
            json.dump(self.metrics, f, indent=2)
        
        print("=" * 60)
        print(f"🎉 Profiling completed in {total_time:.2f} seconds")
        print(f"📊 Results saved to: {results_path}")
        
        # Print summary
        self.print_summary()
    
    def print_summary(self):
        """Print profiling summary"""
        print("\n📋 PROFILING SUMMARY")
        print("-" * 40)
        
        # System resources
        if "system_resources" in self.metrics["tests"]:
            sys_res = self.metrics["tests"]["system_resources"]
            print(f"💻 System Resources:")
            print(f"   Process Memory: {sys_res['process']['memory_rss_mb']} MB")
            print(f"   Process CPU: {sys_res['process']['cpu_percent']}%")
            print(f"   System Memory: {sys_res['system']['memory_percent']}%")
            print(f"   System CPU: {sys_res['system']['cpu_percent']}%")
        
        # Framework operations
        if "framework_operations" in self.metrics["tests"]:
            fw_ops = self.metrics["tests"]["framework_operations"]
            print(f"\n🧠 Framework Operations:")
            for framework, result in fw_ops.items():
                if result["status"] == "success":
                    print(f"   {framework}: {result['total_ms']}ms")
                else:
                    print(f"   {framework}: FAILED")
        
        # Memory operations
        if "memory_operations" in self.metrics["tests"]:
            mem_ops = self.metrics["tests"]["memory_operations"]
            if mem_ops["status"] == "success":
                print(f"\n🧠 Memory Operations:")
                print(f"   Store: {mem_ops['store_ms']}ms")
                print(f"   Search: {mem_ops['search_ms']}ms")
                print(f"   Consolidate: {mem_ops['consolidate_ms']}ms")
        
        # Memory profiling
        if "memory_profiling" in self.metrics["tests"]:
            mem_prof = self.metrics["tests"]["memory_profiling"]
            print(f"\n🧠 Memory Profiling:")
            print(f"   Peak Memory: {mem_prof['peak_mb']} MB")
            print(f"   Memory Growth: {mem_prof['growth_mb']} MB")


def main():
    """Main profiling function"""
    profiler = AIWorkerProfiler()
    profiler.run_comprehensive_profiling()


if __name__ == "__main__":
    main()
