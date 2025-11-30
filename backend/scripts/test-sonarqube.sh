#!/bin/bash
# Script untuk test koneksi SonarCloud

echo "Testing SonarCloud Configuration..."
echo ""

# Check environment variables
if [ -z "$SONARCLOUD_URL" ]; then
    echo "❌ SONARCLOUD_URL is not set"
    exit 1
else
    echo "✅ SONARCLOUD_URL: $SONARCLOUD_URL"
fi

if [ -z "$SONARCLOUD_TOKEN" ]; then
    echo "❌ SONARCLOUD_TOKEN is not set"
    exit 1
else
    echo "✅ SONARCLOUD_TOKEN: ${SONARCLOUD_TOKEN:0:10}... (hidden)"
fi

if [ -z "$SONARCLOUD_PROJECT_KEY" ]; then
    echo "❌ SONARCLOUD_PROJECT_KEY is not set"
    exit 1
else
    echo "✅ SONARCLOUD_PROJECT_KEY: $SONARCLOUD_PROJECT_KEY"
fi

echo ""
echo "Testing SonarCloud API connection..."

# Test API call
AUTH=$(echo -n "$SONARCLOUD_TOKEN:" | base64)
RESPONSE=$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Basic $AUTH" \
    "$SONARCLOUD_URL/api/issues/search?componentKeys=$SONARCLOUD_PROJECT_KEY&ps=1")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ SonarCloud API connection successful!"
    echo "Response preview: ${BODY:0:200}..."
else
    echo "❌ SonarCloud API connection failed!"
    echo "HTTP Status: $HTTP_CODE"
    echo "Response: $BODY"
    exit 1
fi
