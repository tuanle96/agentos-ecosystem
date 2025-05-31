"""
Memory Consolidation Engine for AgentOS
Week 4: Advanced Memory System Implementation

This module implements memory consolidation algorithms that convert episodic memories
into semantic knowledge, extract patterns, and optimize memory storage.
"""

import asyncio
import json
import logging
import re
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Tuple
from dataclasses import dataclass
from collections import defaultdict, Counter

import aiohttp
from universal_memory import MemoryEntry, MemoryType, FrameworkType, ConsolidationResult


@dataclass
class ConsolidationRule:
    """Rule for memory consolidation"""
    trigger_type: str  # time_based, frequency_based, importance_based
    threshold: float
    importance_boost: float
    pattern_types: List[str]
    consolidation_strategy: str


@dataclass
class MemoryPattern:
    """Identified pattern in episodic memories"""
    pattern_type: str  # causal, temporal, conceptual, behavioral
    description: str
    confidence: float
    supporting_memories: List[str]
    extracted_knowledge: str
    concepts: List[str]


class MemoryConsolidationEngine:
    """
    Advanced Memory Consolidation Engine

    Implements sophisticated algorithms for:
    - Pattern recognition in episodic memories
    - Knowledge extraction and generalization
    - Semantic memory creation from patterns
    - Strategic forgetting based on importance decay
    """

    def __init__(self,
                 api_base_url: str = "http://localhost:8000",
                 llm_endpoint: str = None):
        """
        Initialize Memory Consolidation Engine

        Args:
            api_base_url: Base URL for AgentOS API
            llm_endpoint: Optional LLM endpoint for pattern analysis
        """
        self.api_base_url = api_base_url.rstrip('/')
        self.llm_endpoint = llm_endpoint
        self.logger = logging.getLogger(__name__)

        # Consolidation rules
        self.consolidation_rules = [
            ConsolidationRule(
                trigger_type="time_based",
                threshold=24.0,  # hours
                importance_boost=0.1,
                pattern_types=["temporal", "causal"],
                consolidation_strategy="pattern_extraction"
            ),
            ConsolidationRule(
                trigger_type="frequency_based",
                threshold=5.0,  # number of similar memories
                importance_boost=0.2,
                pattern_types=["conceptual", "behavioral"],
                consolidation_strategy="concept_clustering"
            ),
            ConsolidationRule(
                trigger_type="importance_based",
                threshold=0.8,  # importance score
                importance_boost=0.3,
                pattern_types=["causal", "conceptual"],
                consolidation_strategy="knowledge_synthesis"
            )
        ]

        # Pattern recognition templates
        self.pattern_templates = {
            "causal": [
                r"because of (.+), (.+) happened",
                r"(.+) led to (.+)",
                r"as a result of (.+), (.+)",
                r"(.+) caused (.+)"
            ],
            "temporal": [
                r"after (.+), (.+) occurred",
                r"before (.+), (.+) was",
                r"during (.+), (.+) happened",
                r"(.+) then (.+)"
            ],
            "conceptual": [
                r"(.+) is similar to (.+)",
                r"(.+) relates to (.+)",
                r"(.+) and (.+) share (.+)",
                r"(.+) belongs to (.+)"
            ],
            "behavioral": [
                r"when (.+), I (.+)",
                r"if (.+), then (.+)",
                r"(.+) always results in (.+)",
                r"(.+) pattern: (.+)"
            ]
        }

    async def consolidate_framework_memories(self,
                                           framework: FrameworkType,
                                           time_window: timedelta = timedelta(hours=24)) -> ConsolidationResult:
        """
        Consolidate memories for a specific framework

        Args:
            framework: Framework to consolidate memories for
            time_window: Time window for episodic memories

        Returns:
            Consolidation result with statistics and new knowledge
        """
        consolidation_id = f"consolidation_{framework.value}_{int(datetime.now().timestamp())}"
        started_at = datetime.now()

        try:
            self.logger.info(f"Starting memory consolidation for {framework.value}")

            # Step 1: Retrieve episodic memories within time window
            episodic_memories = await self._get_episodic_memories(framework, time_window)
            self.logger.info(f"Retrieved {len(episodic_memories)} episodic memories")

            if len(episodic_memories) < 2:
                return ConsolidationResult(
                    consolidation_id=consolidation_id,
                    framework=framework,
                    episodic_count=len(episodic_memories),
                    semantic_count=0,
                    consolidation_score=0.0,
                    patterns_found=[],
                    new_memories_created=0,
                    started_at=started_at,
                    completed_at=datetime.now()
                )

            # Step 2: Identify patterns in episodic memories
            patterns = await self._identify_patterns(episodic_memories)
            self.logger.info(f"Identified {len(patterns)} patterns")

            # Step 3: Extract knowledge from patterns
            semantic_memories = await self._extract_semantic_knowledge(patterns, framework)
            self.logger.info(f"Extracted {len(semantic_memories)} semantic memories")

            # Step 4: Store new semantic memories
            new_memory_ids = []
            for semantic_memory in semantic_memories:
                memory_id = await self._store_consolidated_memory(semantic_memory, framework)
                new_memory_ids.append(memory_id)

            # Step 5: Calculate consolidation score
            consolidation_score = self._calculate_consolidation_score(patterns, semantic_memories)

            # Step 6: Update memory importance scores
            await self._update_memory_importance(episodic_memories, patterns)

            completed_at = datetime.now()

            result = ConsolidationResult(
                consolidation_id=consolidation_id,
                framework=framework,
                episodic_count=len(episodic_memories),
                semantic_count=len(semantic_memories),
                consolidation_score=consolidation_score,
                patterns_found=[p.description for p in patterns],
                new_memories_created=len(new_memory_ids),
                started_at=started_at,
                completed_at=completed_at
            )

            # Store consolidation record
            await self._store_consolidation_record(result)

            self.logger.info(f"Consolidation completed: {consolidation_score:.2f} score, {len(semantic_memories)} new memories")
            return result

        except Exception as e:
            self.logger.error(f"Consolidation failed: {e}")
            raise

    async def _get_episodic_memories(self,
                                   framework: FrameworkType,
                                   time_window: timedelta) -> List[MemoryEntry]:
        """Retrieve episodic memories within time window"""
        cutoff_time = datetime.now() - time_window

        # Query episodic memories from API
        async with aiohttp.ClientSession() as session:
            params = {
                "framework": framework.value,
                "memory_type": "episodic",
                "since": cutoff_time.isoformat(),
                "limit": 100
            }

            async with session.get(
                f"{self.api_base_url}/api/v1/memory/episodic",
                params=params
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    return [self._dict_to_memory_entry(mem) for mem in data.get("memories", [])]
                else:
                    self.logger.warning(f"Failed to retrieve episodic memories: {response.status}")
                    return []

    async def _identify_patterns(self, memories: List[MemoryEntry]) -> List[MemoryPattern]:
        """Identify patterns in episodic memories"""
        patterns = []

        # Group memories by concepts for pattern analysis
        concept_groups = defaultdict(list)
        for memory in memories:
            for concept in memory.concepts:
                concept_groups[concept].append(memory)

        # Identify temporal patterns
        temporal_patterns = await self._find_temporal_patterns(memories)
        patterns.extend(temporal_patterns)

        # Identify causal patterns
        causal_patterns = await self._find_causal_patterns(memories)
        patterns.extend(causal_patterns)

        # Identify conceptual patterns
        conceptual_patterns = await self._find_conceptual_patterns(concept_groups)
        patterns.extend(conceptual_patterns)

        # Identify behavioral patterns
        behavioral_patterns = await self._find_behavioral_patterns(memories)
        patterns.extend(behavioral_patterns)

        # Filter and rank patterns by confidence
        patterns = [p for p in patterns if p.confidence >= 0.6]
        patterns.sort(key=lambda x: x.confidence, reverse=True)

        return patterns[:10]  # Top 10 patterns

    async def _find_temporal_patterns(self, memories: List[MemoryEntry]) -> List[MemoryPattern]:
        """Find temporal patterns in memories"""
        patterns = []

        # Sort memories by timestamp
        sorted_memories = sorted(memories, key=lambda m: m.created_at)

        # Look for sequences
        for i in range(len(sorted_memories) - 1):
            current = sorted_memories[i]
            next_mem = sorted_memories[i + 1]

            # Check if memories are related and sequential
            time_diff = (next_mem.created_at - current.created_at).total_seconds()
            if time_diff < 3600:  # Within 1 hour
                # Check for concept overlap
                common_concepts = set(current.concepts) & set(next_mem.concepts)
                if common_concepts:
                    pattern = MemoryPattern(
                        pattern_type="temporal",
                        description=f"Sequential pattern: {current.content[:50]}... → {next_mem.content[:50]}...",
                        confidence=0.7 + len(common_concepts) * 0.1,
                        supporting_memories=[current.id, next_mem.id],
                        extracted_knowledge=f"When {list(common_concepts)[0]} occurs, it often leads to related activities",
                        concepts=list(common_concepts)
                    )
                    patterns.append(pattern)

        return patterns

    async def _find_causal_patterns(self, memories: List[MemoryEntry]) -> List[MemoryPattern]:
        """Find causal patterns using text analysis"""
        patterns = []

        for template_type, templates in self.pattern_templates.items():
            if template_type == "causal":
                for memory in memories:
                    for template in templates:
                        matches = re.findall(template, memory.content.lower())
                        if matches:
                            for match in matches:
                                if len(match) == 2:  # Cause and effect
                                    cause, effect = match
                                    pattern = MemoryPattern(
                                        pattern_type="causal",
                                        description=f"Causal relationship: {cause} → {effect}",
                                        confidence=0.8,
                                        supporting_memories=[memory.id],
                                        extracted_knowledge=f"Understanding: {cause} typically results in {effect}",
                                        concepts=[cause.strip(), effect.strip()]
                                    )
                                    patterns.append(pattern)

        return patterns

    async def _find_conceptual_patterns(self, concept_groups: Dict[str, List[MemoryEntry]]) -> List[MemoryPattern]:
        """Find conceptual patterns in grouped memories"""
        patterns = []

        for concept, group_memories in concept_groups.items():
            if len(group_memories) >= 3:  # Need multiple instances
                # Extract common themes
                all_concepts = []
                for memory in group_memories:
                    all_concepts.extend(memory.concepts)

                concept_counts = Counter(all_concepts)
                common_concepts = [c for c, count in concept_counts.most_common(5) if count >= 2]

                if len(common_concepts) >= 2:
                    pattern = MemoryPattern(
                        pattern_type="conceptual",
                        description=f"Conceptual cluster around '{concept}' with related concepts: {', '.join(common_concepts[:3])}",
                        confidence=0.6 + min(len(group_memories) * 0.1, 0.3),
                        supporting_memories=[m.id for m in group_memories],
                        extracted_knowledge=f"Knowledge domain: {concept} is associated with {', '.join(common_concepts[:3])}",
                        concepts=common_concepts
                    )
                    patterns.append(pattern)

        return patterns

    async def _find_behavioral_patterns(self, memories: List[MemoryEntry]) -> List[MemoryPattern]:
        """Find behavioral patterns in memories"""
        patterns = []

        # Look for repeated actions or decisions
        action_patterns = defaultdict(list)

        for memory in memories:
            # Simple action extraction (would be more sophisticated in production)
            content_lower = memory.content.lower()
            if any(word in content_lower for word in ["decided", "chose", "selected", "did"]):
                # Extract the action context
                for concept in memory.concepts:
                    action_patterns[concept].append(memory)

        for action, action_memories in action_patterns.items():
            if len(action_memories) >= 2:
                pattern = MemoryPattern(
                    pattern_type="behavioral",
                    description=f"Behavioral pattern: repeated actions related to '{action}'",
                    confidence=0.7,
                    supporting_memories=[m.id for m in action_memories],
                    extracted_knowledge=f"Behavioral insight: tendency to engage with {action} in similar contexts",
                    concepts=[action]
                )
                patterns.append(pattern)

        return patterns

    async def _extract_semantic_knowledge(self,
                                        patterns: List[MemoryPattern],
                                        framework: FrameworkType) -> List[Dict[str, Any]]:
        """Extract semantic knowledge from identified patterns"""
        semantic_memories = []

        for pattern in patterns:
            # Create semantic memory from pattern
            semantic_memory = {
                "content": pattern.extracted_knowledge,
                "concepts": pattern.concepts,
                "importance": min(pattern.confidence, 1.0),
                "framework": framework.value,
                "source_type": "consolidation",
                "metadata": {
                    "pattern_type": pattern.pattern_type,
                    "confidence": pattern.confidence,
                    "supporting_memories": pattern.supporting_memories,
                    "consolidation_timestamp": datetime.now().isoformat()
                }
            }
            semantic_memories.append(semantic_memory)

        return semantic_memories

    async def _store_consolidated_memory(self,
                                       semantic_memory: Dict[str, Any],
                                       framework: FrameworkType) -> str:
        """Store consolidated semantic memory"""
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.api_base_url}/api/v1/memory/semantic/store",
                json=semantic_memory
            ) as response:
                if response.status == 201:
                    data = await response.json()
                    return data["memory_id"]
                else:
                    raise Exception(f"Failed to store consolidated memory: {response.status}")

    def _calculate_consolidation_score(self,
                                     patterns: List[MemoryPattern],
                                     semantic_memories: List[Dict[str, Any]]) -> float:
        """Calculate overall consolidation quality score"""
        if not patterns:
            return 0.0

        # Base score from pattern confidence
        avg_confidence = sum(p.confidence for p in patterns) / len(patterns)

        # Bonus for pattern diversity
        pattern_types = set(p.pattern_type for p in patterns)
        diversity_bonus = len(pattern_types) * 0.1

        # Bonus for knowledge extraction
        extraction_bonus = len(semantic_memories) * 0.05

        score = avg_confidence + diversity_bonus + extraction_bonus
        return min(score, 1.0)

    async def _update_memory_importance(self,
                                      episodic_memories: List[MemoryEntry],
                                      patterns: List[MemoryPattern]):
        """Update importance scores of episodic memories based on patterns"""
        memory_pattern_map = defaultdict(list)

        # Map memories to patterns they support
        for pattern in patterns:
            for memory_id in pattern.supporting_memories:
                memory_pattern_map[memory_id].append(pattern)

        # Update importance scores
        for memory in episodic_memories:
            if memory.id in memory_pattern_map:
                supporting_patterns = memory_pattern_map[memory.id]
                importance_boost = sum(p.confidence * 0.1 for p in supporting_patterns)
                new_importance = min(memory.importance + importance_boost, 1.0)

                # Update memory importance via real API call
                await self._update_memory_importance_api(memory.id, new_importance)
                self.logger.debug(f"Updated memory {memory.id} importance: {memory.importance:.2f} → {new_importance:.2f}")

    async def _store_consolidation_record(self, result: ConsolidationResult):
        """Store consolidation record for tracking"""
        record = {
            "consolidation_id": result.consolidation_id,
            "framework": result.framework.value,
            "episodic_count": result.episodic_count,
            "semantic_count": result.semantic_count,
            "consolidation_score": result.consolidation_score,
            "patterns_found": result.patterns_found,
            "new_memories_created": result.new_memories_created,
            "started_at": result.started_at.isoformat(),
            "completed_at": result.completed_at.isoformat() if result.completed_at else None
        }

        # Store consolidation record via real API call
        await self._store_consolidation_record_api(record)
        self.logger.info(f"Stored consolidation record: {result.consolidation_id}")

    def _dict_to_memory_entry(self, data: Dict[str, Any]) -> MemoryEntry:
        """Convert dictionary to MemoryEntry"""
        return MemoryEntry(
            id=data.get("id", ""),
            content=data.get("content", ""),
            memory_type=MemoryType(data.get("memory_type", "episodic")),
            framework=FrameworkType(data.get("framework", "universal")),
            concepts=data.get("concepts", []),
            importance=data.get("importance", 0.5),
            metadata=data.get("metadata", {}),
            created_at=datetime.fromisoformat(data["created_at"]) if data.get("created_at") else datetime.now(),
            updated_at=datetime.fromisoformat(data["updated_at"]) if data.get("updated_at") else datetime.now()
        )

    async def _update_memory_importance_api(self, memory_id: str, new_importance: float):
        """Update memory importance via AgentOS API"""
        try:
            async with aiohttp.ClientSession() as session:
                update_data = {
                    "importance": new_importance,
                    "updated_at": datetime.now().isoformat()
                }

                async with session.patch(
                    f"{self.api_base_url}/api/v1/memory/{memory_id}",
                    json=update_data
                ) as response:
                    if response.status == 200:
                        self.logger.debug(f"Successfully updated memory {memory_id} importance to {new_importance:.2f}")
                    else:
                        self.logger.warning(f"Failed to update memory {memory_id} importance: {response.status}")

        except Exception as e:
            self.logger.error(f"Error updating memory importance for {memory_id}: {str(e)}")

    async def _store_consolidation_record_api(self, record: Dict[str, Any]):
        """Store consolidation record via AgentOS API"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.api_base_url}/api/v1/memory/consolidation/records",
                    json=record
                ) as response:
                    if response.status == 201:
                        self.logger.debug(f"Successfully stored consolidation record: {record['consolidation_id']}")
                    else:
                        self.logger.warning(f"Failed to store consolidation record: {response.status}")

        except Exception as e:
            self.logger.error(f"Error storing consolidation record: {str(e)}")
