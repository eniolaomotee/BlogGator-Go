#!/bin/bash

BASE_URL="http://localhost:8080/api"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Gator API Client Example ===${NC}\n"

# 1. Register a new user
echo -e "${GREEN}1. Registering new user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }')

echo "$REGISTER_RESPONSE" | jq .

# Extract token
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token')

if [ "$TOKEN" == "null" ]; then
  echo -e "${RED}Registration failed!${NC}"
  exit 1
fi

echo -e "\n${GREEN}Token: $TOKEN${NC}\n"

# 2. Get current user info
echo -e "${GREEN}2. Getting current user info...${NC}"
curl -s -X GET "$BASE_URL/me" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 3. Add a feed
echo -e "\n${GREEN}3. Adding a feed...${NC}"
FEED_RESPONSE=$(curl -s -X POST "$BASE_URL/feeds" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TechCrunch",
    "url": "https://techcrunch.com/feed/"
  }')

echo "$FEED_RESPONSE" | jq .

FEED_ID=$(echo "$FEED_RESPONSE" | jq -r '.id')

# 4. Get all feeds
echo -e "\n${GREEN}4. Getting all feeds...${NC}"
curl -s -X GET "$BASE_URL/feeds" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 5. Get posts
echo -e "\n${GREEN}5. Getting posts (may be empty if aggregator hasn't run)...${NC}"
curl -s -X GET "$BASE_URL/posts?limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 6. Get posts with filters
echo -e "\n${GREEN}6. Getting posts with filters...${NC}"
curl -s -X GET "$BASE_URL/posts?limit=10&sort=title&order=asc" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 7. Login (test authentication)
echo -e "\n${GREEN}7. Testing login...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }')

echo "$LOGIN_RESPONSE" | jq .

# 8. Health check
echo -e "\n${GREEN}8. Health check...${NC}"
curl -s -X GET "$BASE_URL/health" | jq .

echo -e "\n${BLUE}=== API Client Test Complete ===${NC}"