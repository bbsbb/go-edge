#!/usr/bin/env bash
set -euo pipefail

# Runs a sequence of API requests against the sweetshop.
# Usage: ./scripts/requests.sh

BASE_URL="${BASE_URL:-http://localhost:8080}"
ORG="${ORG:-dev-shop}"

header=(-H "X-Organization-Slug: $ORG" -H "Content-Type: application/json")

echo "=== Create product ==="
product=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/products" \
  "${header[@]}" \
  -d '{"name":"Chocolate Cake","category":"ice_cream","price_cents":1500}')
body=$(echo "$product" | sed '$d')
code=$(echo "$product" | tail -1)
echo "Status: $code"
echo "$body" | jq .
product_id=$(echo "$body" | jq -r '.id')

echo ""
echo "=== Create second product (for delete test) ==="
product2=$(curl -s -X POST "$BASE_URL/products" \
  "${header[@]}" \
  -d '{"name":"Marshmallow Puff","category":"marshmallow","price_cents":500}')
echo "$product2" | jq .
product2_id=$(echo "$product2" | jq -r '.id')

echo ""
echo "=== List products ==="
curl -s -X GET "$BASE_URL/products" "${header[@]}" | jq .

echo ""
echo "=== Get product ==="
curl -s -X GET "$BASE_URL/products/$product_id" "${header[@]}" | jq .

echo ""
echo "=== Update product ==="
curl -s -X PUT "$BASE_URL/products/$product_id" \
  "${header[@]}" \
  -d '{"name":"Chocolate Cake Deluxe","category":"ice_cream","price_cents":1800}' | jq .

echo ""
echo "=== Open order ==="
order=$(curl -s -X POST "$BASE_URL/orders" "${header[@]}")
echo "$order" | jq .
order_id=$(echo "$order" | jq -r '.id')

echo ""
echo "=== Add item to order ==="
curl -s -X POST "$BASE_URL/orders/$order_id/items" \
  "${header[@]}" \
  -d "{\"product_id\":\"$product_id\",\"quantity\":2}" | jq .

echo ""
echo "=== Get order ==="
curl -s -X GET "$BASE_URL/orders/$order_id" "${header[@]}" | jq .

echo ""
echo "=== Close order ==="
curl -s -X POST "$BASE_URL/orders/$order_id/close" "${header[@]}" | jq .

echo ""
echo "=== Delete product (no orders) ==="
code=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/products/$product2_id" "${header[@]}")
echo "Status: $code"

echo ""
echo "=== Verify deletion ==="
curl -s -X GET "$BASE_URL/products" "${header[@]}" | jq .

echo ""
echo "Done."
