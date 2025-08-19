#!/bin/bash

# Test script for Spectro Lab Backend API

BASE_URL="http://localhost:8080"

echo "🧪 Testing Spectro Lab Backend API"
echo "=================================="

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s -X GET "$BASE_URL/health" | jq .

echo -e "\n2. Testing login endpoint..."
# Test login (creates user)
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}')

echo $LOGIN_RESPONSE | jq .

# Extract token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to get token"
    exit 1
fi

echo -e "\n3. Testing lab creation..."
# Test lab creation
LAB_RESPONSE=$(curl -s -X POST "$BASE_URL/api/labs" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Test Lab",
    "duration": 60
  }')

echo $LAB_RESPONSE | jq .

# Extract lab ID
LAB_ID=$(echo $LAB_RESPONSE | jq -r '.id')

if [ "$LAB_ID" = "null" ] || [ -z "$LAB_ID" ]; then
    echo "❌ Failed to get lab ID"
    exit 1
fi

echo -e "\n4. Testing get lab..."
# Test get lab
curl -s -X GET "$BASE_URL/api/labs/$LAB_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n5. Testing get user labs..."
# Test get user labs
curl -s -X GET "$BASE_URL/api/labs" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n6. Testing admin endpoints..."
# Test admin endpoints (should fail for regular user)
echo "Testing admin access with regular user (should fail):"
curl -s -X GET "$BASE_URL/api/admin/labs" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n7. Testing admin login..."
# Login as admin
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@spectrocloud.com"}')

echo $ADMIN_RESPONSE | jq .

# Extract admin token
ADMIN_TOKEN=$(echo $ADMIN_RESPONSE | jq -r '.token')

if [ "$ADMIN_TOKEN" = "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Failed to get admin token"
    exit 1
fi

echo -e "\n8. Testing admin endpoints with admin token..."
# Test admin endpoints with admin token
curl -s -X GET "$BASE_URL/api/admin/labs" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

curl -s -X GET "$BASE_URL/api/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

echo -e "\n9. Testing user management..."
# Test user creation
curl -s -X POST "$BASE_URL/api/admin/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "email": "newuser@example.com",
    "name": "New User",
    "role": "user"
  }' | jq .

echo -e "\n10. Testing role update..."
# Get user ID for role update
USERS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

USER_ID=$(echo $USERS_RESPONSE | jq -r '.[] | select(.email == "newuser@example.com") | .id')

if [ "$USER_ID" != "null" ] && [ -n "$USER_ID" ]; then
    curl -s -X PUT "$BASE_URL/api/admin/users/$USER_ID/role" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d '{"role": "admin"}' | jq .
fi

echo -e "\n✅ All tests completed!"
