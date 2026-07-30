#!/bin/bash
set -e

echo "=== Building Statping ==="

# Kill any running instance
echo "Stopping any running statping..."
taskkill //F //IM statping.exe 2>/dev/null || true

# Build frontend (outputs to frontend/dist)
echo "Building frontend..."
cd frontend
npm run build
cd ..

# Sync built assets to source/dist (Go embeds from here)
echo "Syncing to source/dist..."
rm -rf source/dist/assets source/dist/index.html
cp -r frontend/dist/assets source/dist/
cp frontend/dist/index.html source/dist/

# Update base.gohtml with correct asset hashes
INDEX_JS=$(basename source/dist/assets/index-*.js)
INDEX_CSS=$(basename source/dist/assets/index-*.css)
VENDOR_JS=$(basename source/dist/assets/vendor-*.js)

sed -i "s|assets/index-[^\"]*\.js|assets/${INDEX_JS}|g" source/dist/base.gohtml
sed -i "s|assets/index-[^\"]*\.css|assets/${INDEX_CSS}|g" source/dist/base.gohtml
sed -i "s|assets/vendor-[^\"]*\.js|assets/${VENDOR_JS}|g" source/dist/base.gohtml

echo "  JS:  ${INDEX_JS}"
echo "  CSS: ${INDEX_CSS}"

# Build Go binary
echo "Building Go binary..."
go build -o statping.exe ./cmd

echo "=== Done ==="
