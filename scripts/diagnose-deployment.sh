#!/bin/bash
set -e

echo "🔍 Diagnosing Deployment Issues..."
echo "=================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check status
check_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

# 1. Check Backend Container
echo "1️⃣  Backend Container Status:"
echo "----------------------------"
CONTAINER_STATUS=$(sudo docker ps -a | grep dms-backend-prod || echo "")
if [ -z "$CONTAINER_STATUS" ]; then
    echo -e "${RED}❌ Container dms-backend-prod not found${NC}"
else
    echo "$CONTAINER_STATUS"
    if echo "$CONTAINER_STATUS" | grep -q "Up"; then
        echo -e "${GREEN}✅ Container is running${NC}"
    else
        echo -e "${RED}❌ Container is not running${NC}"
        echo ""
        echo "Container logs (last 30 lines):"
        sudo docker logs dms-backend-prod --tail 30 2>&1 || echo "Cannot get logs"
    fi
fi
echo ""

# 2. Check Backend Port
echo "2️⃣  Backend Port 8080:"
echo "---------------------"
if sudo ss -tlnp | grep -q ":8080"; then
    echo -e "${GREEN}✅ Port 8080 is listening${NC}"
    sudo ss -tlnp | grep ":8080"
else
    echo -e "${RED}❌ Port 8080 is not listening${NC}"
fi
echo ""

# 3. Test Backend Health
echo "3️⃣  Backend Health Check:"
echo "------------------------"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" http://127.0.0.1:8080/health 2>/dev/null || echo -e "\n000")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -1)
BODY=$(echo "$HEALTH_RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Backend is healthy (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
else
    echo -e "${RED}❌ Backend health check failed (HTTP $HTTP_CODE)${NC}"
fi
echo ""

# 4. Check Nginx (Backend)
echo "4️⃣  Nginx Status (Backend):"
echo "--------------------------"
if systemctl is-active --quiet nginx; then
    echo -e "${GREEN}✅ Nginx is running${NC}"
else
    echo -e "${RED}❌ Nginx is not running${NC}"
    echo "Attempting to start Nginx..."
    sudo systemctl start nginx || echo "Failed to start Nginx"
fi

# Check Nginx config
if sudo nginx -t 2>&1 | grep -q "successful"; then
    echo -e "${GREEN}✅ Nginx config is valid${NC}"
else
    echo -e "${RED}❌ Nginx config has errors${NC}"
    sudo nginx -t
fi

# Check ports
echo ""
echo "Nginx ports:"
if sudo ss -tlnp | grep -q ":80"; then
    echo -e "${GREEN}✅ Port 80 is listening${NC}"
else
    echo -e "${RED}❌ Port 80 is not listening${NC}"
fi

if sudo ss -tlnp | grep -q ":443"; then
    echo -e "${GREEN}✅ Port 443 is listening${NC}"
else
    echo -e "${YELLOW}⚠️  Port 443 is not listening (SSL may not be configured)${NC}"
fi
echo ""

# 5. Test via Nginx
echo "5️⃣  Testing via Nginx:"
echo "---------------------"
NGINX_HTTP=$(curl -s -w "\n%{http_code}" http://127.0.0.1/health 2>/dev/null || echo -e "\n000")
NGINX_HTTP_CODE=$(echo "$NGINX_HTTP" | tail -1)

if [ "$NGINX_HTTP_CODE" = "200" ] || [ "$NGINX_HTTP_CODE" = "301" ]; then
    echo -e "${GREEN}✅ HTTP via Nginx works (HTTP $NGINX_HTTP_CODE)${NC}"
else
    echo -e "${RED}❌ HTTP via Nginx failed (HTTP $NGINX_HTTP_CODE)${NC}"
fi

NGINX_HTTPS=$(curl -s -k -w "\n%{http_code}" https://127.0.0.1/health 2>/dev/null || echo -e "\n000")
NGINX_HTTPS_CODE=$(echo "$NGINX_HTTPS" | tail -1)

if [ "$NGINX_HTTPS_CODE" = "200" ]; then
    echo -e "${GREEN}✅ HTTPS via Nginx works (HTTP $NGINX_HTTPS_CODE)${NC}"
else
    echo -e "${YELLOW}⚠️  HTTPS via Nginx failed (HTTP $NGINX_HTTPS_CODE)${NC}"
fi
echo ""

# 6. Check Cloud SQL Proxy
echo "6️⃣  Cloud SQL Proxy:"
echo "------------------"
if ps aux | grep -q "[c]loud-sql-proxy"; then
    echo -e "${GREEN}✅ Cloud SQL Proxy is running${NC}"
    ps aux | grep "[c]loud-sql-proxy" | head -1
else
    echo -e "${RED}❌ Cloud SQL Proxy is not running${NC}"
fi
echo ""

# 7. Check Database Connection
echo "7️⃣  Database Connection:"
echo "-----------------------"
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID} 2>/dev/null || echo "")
if [ -z "$DB_PASSWORD" ]; then
    echo -e "${RED}❌ Cannot retrieve database password${NC}"
else
    DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))" 2>/dev/null || echo "")
    DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"
    
    if psql "${DATABASE_URL}" -c "SELECT 1;" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Database connection successful${NC}"
    else
        echo -e "${RED}❌ Database connection failed${NC}"
    fi
fi
echo ""

# 8. Check Environment Variables
echo "8️⃣  Backend Environment Variables:"
echo "----------------------------------"
if sudo docker ps | grep -q dms-backend-prod; then
    echo "CORS_ORIGIN:"
    sudo docker exec dms-backend-prod env | grep CORS_ORIGIN || echo "Not set"
    echo ""
    echo "DISABLE_RATE_LIMIT:"
    sudo docker exec dms-backend-prod env | grep DISABLE_RATE_LIMIT || echo "Not set"
    echo ""
    echo "ENV:"
    sudo docker exec dms-backend-prod env | grep "^ENV=" || echo "Not set"
else
    echo -e "${RED}❌ Cannot check env vars (container not running)${NC}"
fi
echo ""

# 9. Summary
echo "=================================="
echo "📊 Summary:"
echo "=================================="
echo ""

ISSUES=0

if ! sudo docker ps | grep -q dms-backend-prod; then
    echo -e "${RED}❌ Backend container is not running${NC}"
    ISSUES=$((ISSUES + 1))
fi

if ! systemctl is-active --quiet nginx; then
    echo -e "${RED}❌ Nginx is not running${NC}"
    ISSUES=$((ISSUES + 1))
fi

if ! sudo ss -tlnp | grep -q ":8080"; then
    echo -e "${RED}❌ Backend port 8080 is not listening${NC}"
    ISSUES=$((ISSUES + 1))
fi

if ! sudo ss -tlnp | grep -q ":80"; then
    echo -e "${RED}❌ Nginx port 80 is not listening${NC}"
    ISSUES=$((ISSUES + 1))
fi

if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✅ All critical services are running${NC}"
    echo ""
    echo "🧪 Quick test commands:"
    echo "  curl http://127.0.0.1:8080/health"
    echo "  curl http://127.0.0.1/health"
    echo "  curl https://api-pedeve-dev.aretaamany.com/health"
else
    echo -e "${RED}❌ Found $ISSUES critical issue(s)${NC}"
    echo ""
    echo "🔧 Next steps:"
    echo "  1. Check container logs: sudo docker logs dms-backend-prod"
    echo "  2. Check Nginx logs: sudo journalctl -u nginx --tail 50"
    echo "  3. Restart services if needed"
fi

