# AgentOS AI Worker - Multi-Framework Support
# Week 3 Implementation: Framework Integration Package

"""
Multi-Framework Integration Package for AgentOS AI Worker

This package provides unified interfaces for multiple AI frameworks:
- LangChain: Comprehensive AI application framework
- Swarms: Distributed agent coordination
- CrewAI: Role-based multi-agent workflows
- AutoGen: Conversational AI and code generation

Architecture:
- Universal agent interface for framework abstraction
- Framework-specific wrappers for each AI framework
- Intelligent framework selection based on task requirements
- Performance optimization and caching
"""

from typing import Dict, Any, List, Optional
import asyncio
import logging

# Framework availability flags
LANGCHAIN_AVAILABLE = False
SWARMS_AVAILABLE = False
CREWAI_AVAILABLE = False
AUTOGEN_AVAILABLE = False

# Import framework wrappers with fallback handling
try:
    from .langchain_wrapper import LangChainAgentWrapper
    LANGCHAIN_AVAILABLE = True
except ImportError:
    logging.warning("LangChain not available")
    LangChainAgentWrapper = None

try:
    from .swarms_wrapper import SwarmAgentWrapper
    SWARMS_AVAILABLE = True
except ImportError:
    logging.warning("Swarms not available")
    SwarmAgentWrapper = None

try:
    from .crewai_wrapper import CrewAIAgentWrapper
    CREWAI_AVAILABLE = True
except ImportError:
    logging.warning("CrewAI not available")
    CrewAIAgentWrapper = None
    CREWAI_AVAILABLE = False

try:
    from .autogen_wrapper import AutoGenAgentWrapper
    AUTOGEN_AVAILABLE = True
except ImportError:
    logging.warning("AutoGen not available")
    AutoGenAgentWrapper = None
    AUTOGEN_AVAILABLE = False

from .orchestrator import FrameworkOrchestrator

# Framework registry
FRAMEWORK_REGISTRY = {
    'langchain': {
        'wrapper': LangChainAgentWrapper,
        'available': LANGCHAIN_AVAILABLE,
        'description': 'Comprehensive AI application framework',
        'strengths': ['tool_integration', 'memory_management', 'chain_composition'],
        'use_cases': ['general_purpose', 'tool_heavy', 'memory_intensive']
    },
    'swarms': {
        'wrapper': SwarmAgentWrapper,
        'available': SWARMS_AVAILABLE,
        'description': 'Distributed agent coordination framework',
        'strengths': ['multi_agent', 'distributed_processing', 'scalability'],
        'use_cases': ['parallel_processing', 'distributed_tasks', 'swarm_intelligence']
    },
    'crewai': {
        'wrapper': CrewAIAgentWrapper,
        'available': CREWAI_AVAILABLE,
        'description': 'Role-based multi-agent workflow framework',
        'strengths': ['role_based', 'workflow_management', 'collaboration'],
        'use_cases': ['team_workflows', 'role_specialization', 'sequential_tasks']
    },
    'autogen': {
        'wrapper': AutoGenAgentWrapper,
        'available': AUTOGEN_AVAILABLE,
        'description': 'Conversational AI and code generation framework',
        'strengths': ['conversation', 'code_generation', 'multi_turn'],
        'use_cases': ['conversations', 'code_tasks', 'iterative_refinement']
    }
}

def get_available_frameworks() -> List[str]:
    """Get list of available frameworks"""
    return [name for name, info in FRAMEWORK_REGISTRY.items() if info['available']]

def get_framework_info(framework_name: str) -> Optional[Dict[str, Any]]:
    """Get information about a specific framework"""
    return FRAMEWORK_REGISTRY.get(framework_name)

def get_framework_capabilities() -> Dict[str, Any]:
    """Get comprehensive framework capabilities matrix"""
    capabilities = {}
    for name, info in FRAMEWORK_REGISTRY.items():
        if info['available']:
            capabilities[name] = {
                'description': info['description'],
                'strengths': info['strengths'],
                'use_cases': info['use_cases'],
                'status': 'available'
            }
        else:
            capabilities[name] = {
                'description': info['description'],
                'status': 'unavailable',
                'reason': 'Framework not installed'
            }
    return capabilities

__all__ = [
    'LangChainAgentWrapper',
    'SwarmAgentWrapper',
    'CrewAIAgentWrapper',
    'AutoGenAgentWrapper',
    'FrameworkOrchestrator',
    'FRAMEWORK_REGISTRY',
    'get_available_frameworks',
    'get_framework_info',
    'get_framework_capabilities',
    'LANGCHAIN_AVAILABLE',
    'SWARMS_AVAILABLE',
    'CREWAI_AVAILABLE',
    'AUTOGEN_AVAILABLE'
]
