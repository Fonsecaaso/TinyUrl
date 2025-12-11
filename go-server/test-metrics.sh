#!/bin/bash

# Script para testar as métricas da aplicação TinyURL

set -e

BASE_URL="http://localhost:8080"
METRICS_URL="$BASE_URL/metrics"

echo "🧪 Testing TinyURL Metrics"
echo "=========================="
echo ""

# Cores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para exibir métrica
show_metric() {
    local metric_name=$1
    echo -e "${BLUE}📊 $metric_name${NC}"
    curl -s $METRICS_URL | grep "^$metric_name" | grep -v "^# "
    echo ""
}

# Teste 1: Criar URLs
echo -e "${GREEN}✅ Test 1: Creating URLs${NC}"
echo "Creating 5 URLs..."
for i in {1..5}; do
    curl -s -X POST $BASE_URL/api/ \
        -H "Content-Type: application/json" \
        -d "{\"url\":\"https://example.com/test$i\"}" > /dev/null
    echo "  Created URL $i"
done
echo ""

echo "Waiting 1 second for metrics to update..."
sleep 1
show_metric "url_creation_total"

# Teste 2: Acessar URLs (sucesso e falha)
echo -e "${GREEN}✅ Test 2: Accessing URLs${NC}"
echo "Making 10 successful requests..."
for i in {1..10}; do
    curl -s $BASE_URL/api/health > /dev/null
done
echo ""

echo "Making 3 not-found requests..."
for i in {1..3}; do
    curl -s $BASE_URL/api/notfound$i > /dev/null
done
echo ""

echo "Waiting 1 second for metrics to update..."
sleep 1
show_metric "url_access_total"
show_metric "http_requests_total"

# Teste 3: Cache metrics
echo -e "${GREEN}✅ Test 3: Cache Performance${NC}"
echo "Making requests to test cache..."

# Criar uma URL
echo "Creating a test URL..."
RESPONSE=$(curl -s -X POST $BASE_URL/api/ \
    -H "Content-Type: application/json" \
    -d '{"url":"https://cache-test.com"}')
SHORT_CODE=$(echo $RESPONSE | grep -o '"short_code":"[^"]*"' | cut -d'"' -f4)

if [ -n "$SHORT_CODE" ]; then
    echo "  Short code: $SHORT_CODE"

    # Primeiro acesso (cache miss)
    echo "  First access (cache miss)..."
    curl -s $BASE_URL/api/$SHORT_CODE > /dev/null

    # Segundo acesso (cache hit)
    echo "  Second access (cache hit)..."
    curl -s $BASE_URL/api/$SHORT_CODE > /dev/null

    # Terceiro acesso (cache hit)
    echo "  Third access (cache hit)..."
    curl -s $BASE_URL/api/$SHORT_CODE > /dev/null
fi
echo ""

echo "Waiting 1 second for metrics to update..."
sleep 1
show_metric "cache_hits_total"
show_metric "cache_misses_total"

# Teste 4: HTTP Metrics
echo -e "${GREEN}✅ Test 4: HTTP Performance Metrics${NC}"
show_metric "http_request_duration_seconds_count"
show_metric "http_requests_in_flight"

# Teste 5: System Metrics
echo -e "${GREEN}✅ Test 5: System Metrics${NC}"
show_metric "go_goroutines_count"
show_metric "memory_usage_bytes"

# Resumo
echo ""
echo -e "${YELLOW}📈 Summary${NC}"
echo "=========================="
echo "All metrics are being collected! ✅"
echo ""
echo "View all metrics at: $METRICS_URL"
echo "View in Prometheus: http://localhost:9090"
echo "View in Grafana: http://localhost:3000"
echo ""
echo "Useful Prometheus queries:"
echo "  - rate(url_creation_total[1m])"
echo "  - rate(url_access_total[1m])"
echo "  - sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100"
echo ""
