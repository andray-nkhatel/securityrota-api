#!/bin/bash
# Seed script to populate test data - 19 officers total

BASE_URL="http://localhost:8080/api/v1"

echo "Creating Sergeant..."
curl -s -X POST "$BASE_URL/officers" \
  -H "Content-Type: application/json" \
  -d '{"name":"Sgt. Kalongana","role":"sergeant","team":1}'
echo

echo "Creating Female Officers (2)..."
curl -s -X POST "$BASE_URL/officers" \
  -H "Content-Type: application/json" \
  -d '{"name":"Faides","role":"female","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" \
  -H "Content-Type: application/json" \
  -d '{"name":"Abigail","role":"female","team":1}'
echo

echo "Creating Team A Regular Officers (8)..."
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Alexander","role":"regular","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Levy","role":"regular","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Anderson","role":"regular","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Julius","role":"regular","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Richard","role":"regular","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Rodger","role":"regular","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Chipoya","role":"regular","team":1}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Mubiana","role":"regular","team":1}'
echo

echo "Creating Team B Regular Officers (8)..."
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Moses","role":"regular","team":2}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Michael","role":"regular","team":2}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Gift","role":"regular","team":2}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Gibson","role":"regular","team":2}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Sinoya","role":"regular","team":2}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Hendrix","role":"regular","team":2}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Arthur","role":"regular","team":2}'
echo
curl -s -X POST "$BASE_URL/officers" -H "Content-Type: application/json" \
  -d '{"name":"Paul","role":"regular","team":2}'
echo

echo ""
echo "Total: 1 Sergeant + 2 Female + 8 Team A + 8 Team B = 19 Officers"
echo ""

echo "Generating rota for week starting 2025-11-30 (Sunday)..."
curl -s -X POST "$BASE_URL/shifts/generate" \
  -H "Content-Type: application/json" \
  -d '{"week_start":"2025-11-30"}'
echo

echo ""
echo "All officers:"
curl -s "$BASE_URL/officers" | jq .
echo ""
echo "Sunday shifts (2025-11-30):"
curl -s "$BASE_URL/shifts?date=2025-11-30" | jq .
