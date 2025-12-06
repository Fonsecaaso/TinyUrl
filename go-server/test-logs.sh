#!/bin/bash

# Test script to verify logs are flowing through OpenTelemetry to Loki

set -e

echo "🧪 Testing OpenTelemetry Logs Integration"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if services are running
echo -e "${BLUE}📊 Step 1: Checking if services are running${NC}"
echo ""

if ! docker ps | grep -q otel-collector; then
    echo -e "${YELLOW}⚠️  OpenTelemetry Collector is not running${NC}"
    echo "Please start it with: cd observability && docker-compose up -d"
    exit 1
fi
echo -e "${GREEN}✅ OpenTelemetry Collector is running${NC}"

if ! docker ps | grep -q loki; then
    echo -e "${YELLOW}⚠️  Loki is not running${NC}"
    echo "Please start it with: cd observability && docker-compose up -d"
    exit 1
fi
echo -e "${GREEN}✅ Loki is running${NC}"
echo ""

# Check if the Go application is running
echo -e "${BLUE}📊 Step 2: Checking if TinyURL API is running${NC}"
echo ""

if ! curl -s http://localhost:8080/api/health > /dev/null; then
    echo -e "${YELLOW}⚠️  TinyURL API is not running${NC}"
    echo "Please start it with: go run main.go"
    exit 1
fi
echo -e "${GREEN}✅ TinyURL API is running${NC}"
echo ""

# Generate some traffic to create logs
echo -e "${BLUE}📊 Step 3: Generating traffic to create logs${NC}"
echo ""

echo "Creating URLs..."
for i in {1..3}; do
    curl -s -X POST http://localhost:8080/api/ \
        -H "Content-Type: application/json" \
        -d "{\"url\":\"https://example.com/test-$i\"}" > /dev/null
    echo "  Created URL $i"
done
echo ""

echo "Making health checks..."
for i in {1..3}; do
    curl -s http://localhost:8080/api/health > /dev/null
    echo "  Health check $i"
done
echo ""

echo "Making some 404 requests..."
for i in {1..2}; do
    curl -s http://localhost:8080/api/notfound$i > /dev/null
    echo "  404 request $i"
done
echo ""

# Wait for logs to be processed
echo -e "${BLUE}📊 Step 4: Waiting for logs to be processed...${NC}"
sleep 3
echo ""

# Query Loki for logs
echo -e "${BLUE}📊 Step 5: Querying Loki for logs${NC}"
echo ""

LOKI_QUERY='query_range?query={service_name="tinyurl-api"}'
LOKI_URL="http://localhost:3100/loki/api/v1/${LOKI_QUERY}"

RESPONSE=$(curl -s "$LOKI_URL" | jq -r '.status')

if [ "$RESPONSE" == "success" ]; then
    echo -e "${GREEN}✅ Successfully queried Loki${NC}"

    LOG_COUNT=$(curl -s "$LOKI_URL" | jq '.data.result | length')
    echo -e "${GREEN}✅ Found $LOG_COUNT log streams${NC}"

    # Show a sample of recent logs
    echo ""
    echo -e "${BLUE}📋 Recent logs from Loki:${NC}"
    curl -s "$LOKI_URL" | jq -r '.data.result[0].values[-5:][] | .[1]' 2>/dev/null | head -5
else
    echo -e "${YELLOW}⚠️  Failed to query Loki or no logs found yet${NC}"
    echo "This might be normal if logs haven't been flushed yet."
    echo "Try running this script again in a few seconds."
fi

echo ""
echo -e "${GREEN}🎉 Test complete!${NC}"
echo ""
echo "You can view logs in Grafana:"
echo "  1. Open http://localhost:3000"
echo "  2. Go to Explore"
echo "  3. Select Loki as data source"
echo "  4. Query: {service_name=\"tinyurl-api\"}"
echo ""
echo "Or check OpenTelemetry Collector logs:"
echo "  docker logs otel-collector -f"
