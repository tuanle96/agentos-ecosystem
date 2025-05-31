#!/bin/bash

# Week 6: Setup Performance Profiling Tools
# AgentOS Performance Optimization Implementation

set -e

echo "🔧 Setting up Performance Profiling Tools for Week 6"
echo "===================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install k6 (load testing tool)
install_k6() {
    echo -e "${BLUE}📦 Installing k6 load testing tool...${NC}"
    
    if command_exists k6; then
        echo -e "${GREEN}✅ k6 is already installed${NC}"
        k6 version
        return 0
    fi
    
    # Detect OS and install k6
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command_exists brew; then
            brew install k6
        else
            echo -e "${YELLOW}⚠️  Homebrew not found. Installing k6 manually...${NC}"
            curl -s https://github.com/grafana/k6/releases/latest/download/k6-v0.47.0-macos-amd64.tar.gz | tar -xz
            sudo mv k6-v0.47.0-macos-amd64/k6 /usr/local/bin/
            rm -rf k6-v0.47.0-macos-amd64
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        sudo gpg -k
        sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
        echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
        sudo apt-get update
        sudo apt-get install k6
    else
        echo -e "${RED}❌ Unsupported OS for automatic k6 installation${NC}"
        echo "Please install k6 manually from: https://k6.io/docs/getting-started/installation/"
        return 1
    fi
    
    echo -e "${GREEN}✅ k6 installed successfully${NC}"
    k6 version
}

# Function to install Python performance profiling dependencies
install_python_deps() {
    echo -e "${BLUE}🐍 Installing Python performance profiling dependencies...${NC}"
    
    cd agentos-ecosystem/core/ai-worker
    
    # Check if virtual environment exists
    if [ ! -d "venv" ]; then
        echo "📦 Creating Python virtual environment..."
        python3 -m venv venv
    fi
    
    # Activate virtual environment
    source venv/bin/activate
    
    # Install performance profiling packages
    echo "📦 Installing performance profiling packages..."
    pip install --upgrade pip
    pip install memory-profiler psutil cProfile-tools py-spy
    
    # Install additional monitoring tools
    pip install prometheus-client grafana-api
    
    echo -e "${GREEN}✅ Python dependencies installed${NC}"
    
    cd - > /dev/null
}

# Function to install Go profiling tools
install_go_tools() {
    echo -e "${BLUE}🔧 Installing Go profiling tools...${NC}"
    
    # Install pprof tool
    if ! command_exists pprof; then
        echo "📦 Installing pprof..."
        go install github.com/google/pprof@latest
    else
        echo -e "${GREEN}✅ pprof is already installed${NC}"
    fi
    
    # Install go-torch for flame graphs
    if ! command_exists go-torch; then
        echo "📦 Installing go-torch..."
        go install github.com/uber/go-torch@latest
    else
        echo -e "${GREEN}✅ go-torch is already installed${NC}"
    fi
    
    echo -e "${GREEN}✅ Go profiling tools installed${NC}"
}

# Function to setup monitoring tools
setup_monitoring() {
    echo -e "${BLUE}📊 Setting up monitoring infrastructure...${NC}"
    
    # Create monitoring directory
    mkdir -p monitoring/{prometheus,grafana,alertmanager}
    
    # Create Prometheus configuration
    cat > monitoring/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'agentos-api'
    static_configs:
      - targets: ['localhost:8000']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'agentos-ai-worker'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['localhost:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['localhost:9121']
EOF

    # Create alert rules
    cat > monitoring/prometheus/alert_rules.yml << 'EOF'
groups:
  - name: agentos_alerts
    rules:
      - alert: HighResponseTime
        expr: http_request_duration_seconds{quantile="0.95"} > 0.015
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes / 1024 / 1024 > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage detected"
          description: "Memory usage is {{ $value }}MB"
EOF

    # Create Docker Compose for monitoring stack
    cat > monitoring/docker-compose.monitoring.yml << 'EOF'
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: agentos-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus:/etc/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    container_name: agentos-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:latest
    container_name: agentos-postgres-exporter
    ports:
      - "9187:9187"
    environment:
      - DATA_SOURCE_NAME=postgresql://postgres:postgres@host.docker.internal:5432/agentos?sslmode=disable

  redis-exporter:
    image: oliver006/redis_exporter:latest
    container_name: agentos-redis-exporter
    ports:
      - "9121:9121"
    environment:
      - REDIS_ADDR=redis://host.docker.internal:6379

volumes:
  grafana-storage:
EOF

    echo -e "${GREEN}✅ Monitoring infrastructure configured${NC}"
}

# Function to install system monitoring tools
install_system_tools() {
    echo -e "${BLUE}💻 Installing system monitoring tools...${NC}"
    
    # Install htop, iotop, and other system tools
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command_exists brew; then
            brew install htop iotop
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        sudo apt-get update
        sudo apt-get install -y htop iotop sysstat
    fi
    
    echo -e "${GREEN}✅ System monitoring tools installed${NC}"
}

# Function to create performance results directory
setup_results_directory() {
    echo -e "${BLUE}📁 Setting up performance results directory...${NC}"
    
    mkdir -p performance_results/week6_day1
    mkdir -p performance_results/week6_day2
    mkdir -p performance_results/week6_day3
    mkdir -p performance_results/week6_day4
    mkdir -p performance_results/week6_day5
    mkdir -p performance_results/week6_day6
    mkdir -p performance_results/week6_day7
    
    # Create README for results directory
    cat > performance_results/README.md << 'EOF'
# AgentOS Performance Results

This directory contains performance profiling and optimization results for Week 6.

## Directory Structure

- `week6_day1/` - Initial performance profiling and baseline establishment
- `week6_day2/` - Go backend optimization results
- `week6_day3/` - Python AI worker optimization results
- `week6_day4/` - Advanced caching implementation results
- `week6_day5/` - Advanced features implementation results
- `week6_day6/` - Monitoring and observability setup results
- `week6_day7/` - Final testing and validation results

## File Types

- `*.prof` - Go pprof profile files
- `*.json` - JSON performance data and k6 results
- `*.txt` - Text reports and analysis
- `*.md` - Markdown summaries and reports
- `*.sql` - Database analysis queries
- `*.sh` - Shell scripts for analysis

## Tools Used

- **Go Profiling**: pprof, go-torch
- **Python Profiling**: cProfile, memory-profiler, py-spy
- **Load Testing**: k6, Apache Bench
- **Database Analysis**: PostgreSQL built-in tools
- **Monitoring**: Prometheus, Grafana
- **System Monitoring**: htop, iotop, sysstat
EOF

    echo -e "${GREEN}✅ Performance results directory created${NC}"
}

# Main setup function
main() {
    echo -e "${BLUE}🚀 Starting Week 6 Performance Tools Setup...${NC}"
    
    # Check prerequisites
    if ! command_exists go; then
        echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
        exit 1
    fi
    
    if ! command_exists python3; then
        echo -e "${RED}❌ Python 3 is not installed. Please install Python 3 first.${NC}"
        exit 1
    fi
    
    if ! command_exists docker; then
        echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
        exit 1
    fi
    
    # Install tools
    install_k6
    install_python_deps
    install_go_tools
    setup_monitoring
    install_system_tools
    setup_results_directory
    
    echo ""
    echo -e "${GREEN}🎉 Week 6 Performance Tools Setup Completed!${NC}"
    echo -e "${BLUE}📋 Summary:${NC}"
    echo "  ✅ k6 load testing tool"
    echo "  ✅ Python profiling dependencies"
    echo "  ✅ Go profiling tools"
    echo "  ✅ Monitoring infrastructure (Prometheus + Grafana)"
    echo "  ✅ System monitoring tools"
    echo "  ✅ Performance results directory structure"
    echo ""
    echo -e "${YELLOW}🔧 Next Steps:${NC}"
    echo "  1. Start monitoring stack: cd monitoring && docker-compose -f docker-compose.monitoring.yml up -d"
    echo "  2. Run performance profiling: ./scripts/week6_performance_profiling.sh"
    echo "  3. Access Grafana dashboard: http://localhost:3000 (admin/admin)"
    echo "  4. Access Prometheus: http://localhost:9090"
    echo ""
}

# Execute main function
main "$@"
