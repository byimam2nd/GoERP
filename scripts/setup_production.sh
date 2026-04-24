#!/bin/bash

echo "🚀 Starting GoERP Production Setup..."

# 1. Build and Start Containers
docker-compose up -d --build

# 2. Wait for Database
echo "⏳ Waiting for database to be ready..."
sleep 10

# 3. Provision Initial Tenant
echo "🏢 Provisioning Initial Tenant (GoERP Inc)..."
curl -X POST http://localhost:8080/api/v1/provision \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "GoERP Inc",
    "admin_email": "admin@goerp.com",
    "admin_password": "admin_secure_password"
  }'

echo "✅ Setup Complete. Access GoERP at http://localhost:8080"
