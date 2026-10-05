#!/bin/bash
set -e

echo "Starting deployment of Thermodynamic Software Evolution on EC2..."

# 1. Update repository
echo "Pulling latest changes..."
git pull origin main

# 2. Compile Rust AST Engine
echo "Compiling Rust AST engine (release)..."
cd thermodynamic-ast-engine
cargo build --release
cd ..

# Verify Rust binary
if [ ! -f "thermodynamic-ast-engine/target/release/thermodynamic-ast-engine" ]; then
    echo "ERROR: Rust binary failed to compile!"
    exit 1
fi

# 3. Compile Go backend
echo "Compiling Go backend API..."
cd go-backend
go build -o thermo-api main.go evolve.go scanner.go deliverables.go
cd ..

echo "Build complete! Proceed with systemd restart to apply changes."
