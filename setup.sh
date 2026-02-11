#!/bin/bash

echo "Initializing Monorepo Structure..."

# ==========================================
# 1. CONTRACTS & INFRASTRUCTURE
# ==========================================
echo "Creating API & Deploy folders..."
mkdir -p api/openapi
mkdir -p api/proto
mkdir -p deploy

# ==========================================
# 2. SHARED LIBRARY (The "Commons")
# ==========================================
echo "📂 Creating Shared Packages..."
mkdir -p pkg/database
mkdir -p pkg/logger
mkdir -p pkg/utils

# ==========================================
# 3. GO MICROSERVICES (The "Logic")
# ==========================================
services=("auth" "buyer" "seller" "search" "logistics" "payment")

for service in "${services[@]}"; do
    echo "Setting up Service: $service..."
    
    # Create Standard Go Layout
    mkdir -p "services/$service/cmd/server"
    mkdir -p "services/$service/internal/handler"
    mkdir -p "services/$service/internal/service"
    mkdir -p "services/$service/internal/repository"
    mkdir -p "services/$service/internal/domain"
    mkdir -p "services/$service/migrations"

    echo "package main; func main() {}" > "services/$service/cmd/server/main.go"
    echo "FROM golang:1.21-alpine" > "services/$service/Dockerfile"
done

# ==========================================
# 4. PYTHON ML SERVICE (The "Brain")
# ==========================================
echo " Setting up ML Service..."
mkdir -p services/ml-core/app
touch services/ml-core/requirements.txt
touch services/ml-core/Dockerfile

# ==========================================
# 5. FRONTEND (The "Face")
# ==========================================
echo "💻 Setting up Web Folders..."
mkdir -p web/buyer-app
mkdir -p web/seller-dashboard
mkdir -p web/delivery-app

# ==========================================
# 6. ROOT CONFIG FILES
# ==========================================
echo "⚙️ Creating Config Files..."
touch Makefile
touch go.work
touch deploy/docker-compose.yml
touch deploy/init.sql
echo "# E-Commerce Project" > README.md

echo "Structure Created Successfully!"
