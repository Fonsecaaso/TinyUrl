#!/bin/bash

echo "🔍 Testing log pipeline: Go App → OTel Collector → Loki"
echo ""

# Check if services are running
echo "1️⃣ Checking if services are running..."
echo ""

if ! docker ps | grep -q "otel-collector"; then
    echo "❌ OTel Collector is not running"
    echo "   Run: cd observability && docker-compose up -d"
    exit 1
fi

if ! docker ps | grep -q "loki"; then
    echo "❌ Loki is not running"
    echo "   Run: cd observability && docker-compose up -d"
    exit 1
fi

echo "✅ OTel Collector is running"
echo "✅ Loki is running"
echo ""

# Check OTel Collector logs
echo "2️⃣ Checking OTel Collector logs (last 10 lines)..."
echo ""
docker logs otel-collector --tail 10
echo ""

# Check Loki status
echo "3️⃣ Checking Loki status..."
echo ""
curl -s http://localhost:3100/ready
echo ""
echo ""

# Query Loki for logs
echo "4️⃣ Querying Loki for recent logs..."
echo ""
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={service_name="tinyurl-api"}' \
  --data-urlencode 'limit=10' | jq '.'
echo ""

echo "✅ Test completed!"
echo ""
echo "📊 To view logs in Grafana:"
echo "   1. Open http://localhost:3000"
echo "   2. Go to Explore"
echo "   3. Select Loki as data source"
echo "   4. Use query: {service_name=\"tinyurl-api\"}"
