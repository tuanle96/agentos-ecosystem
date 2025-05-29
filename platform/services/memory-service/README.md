# Memory Service

> Advanced memory management with vector databases

## Overview

The Memory Service provides sophisticated memory capabilities including working memory, long-term memory, and semantic search using vector databases.

## Features

- Working memory management
- Long-term memory storage
- Vector embeddings and semantic search
- Memory consolidation and retrieval
- Multi-modal memory support

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Vector DBs**: Pinecone, Weaviate, Qdrant
- **Search**: Elasticsearch
- **Embeddings**: OpenAI, Sentence Transformers

## Development

```bash
# Build service
make build-memory-service

# Run tests
make test-memory-service

# Start with hot reload
cd services/memory-service && air
```

## Key Components

- **Memory Manager**: Core memory operations
- **Vector Store**: Vector database operations
- **Embedding Service**: Generate embeddings
- **Search Engine**: Semantic search capabilities
- **Memory Consolidation**: Background processing

## Environment Variables

```bash
ELASTICSEARCH_URL=http://elasticsearch:9200
PINECONE_API_KEY=your-key
OPENAI_API_KEY=your-key
WEAVIATE_URL=http://weaviate:8080
```