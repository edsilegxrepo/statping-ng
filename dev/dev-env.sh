#!/usr/bin/env sh

echo "Starting!"

echo "Serving Vue frontend first..."
cd frontend && npm run dev &

cd /go/src/github.com/statping-ng/statping-ng || exit 1
cd source || exit 1
mkdir -p dist
cp -R ../frontend/dist ./dist

echo "Now serving Vue, lets build the golang backend now..."
cd /go/src/github.com/statping-ng/statping-ng || exit 1
modd -f dev/modd.conf
