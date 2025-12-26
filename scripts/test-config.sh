#!/bin/bash
# Test script for VoidAbyss configuration

echo "=== Voidabyss Configuration Test ==="
echo ""

# Check if vb executable exists
if [ ! -f "./vb" ]; then
    echo "❌ vb executable not found. Building..."
    go build -o vb ./cmd/vb
    if [ $? -ne 0 ]; then
        echo "❌ Build failed"
        exit 1
    fi
    echo "✅ Build successful"
fi

echo ""
echo "=== Config Directory ==="
CONFIG_DIR="$HOME/.config/voidabyss"
echo "Expected location: $CONFIG_DIR"

if [ -d "$CONFIG_DIR" ]; then
    echo "✅ Config directory exists"
    
    if [ -f "$CONFIG_DIR/init.lua" ]; then
        echo "✅ init.lua exists"
        echo ""
        echo "=== Current init.lua ==="
        cat "$CONFIG_DIR/init.lua"
    else
        echo "⚠️  init.lua not found"
    fi
    
    if [ -d "$CONFIG_DIR/plugins" ]; then
        echo "✅ plugins directory exists"
    else
        echo "⚠️  plugins directory not found"
    fi
else
    echo "⚠️  Config directory does not exist yet"
    echo "   (Will be created on first editor run)"
fi

echo ""
echo "=== Running Tests ==="
go test ./internal/config -v | grep -E "(PASS|FAIL|RUN)"

echo ""
echo "=== Quick Editor Test ==="
echo "Creating test file..."
echo -e "line 1\nline 2\nline 3" > /tmp/vb_test.txt

echo "Testing config loading..."
# This will create the config directory
timeout 1 ./vb /tmp/vb_test.txt 2>&1 || true

echo ""
if [ -d "$CONFIG_DIR" ]; then
    echo "✅ Config directory created successfully"
    if [ -f "$CONFIG_DIR/init.lua" ]; then
        echo "✅ Default init.lua created"
    fi
else
    echo "❌ Config directory was not created"
fi

echo ""
echo "=== Test Complete ==="
