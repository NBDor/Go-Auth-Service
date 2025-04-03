#!/bin/bash

echo "Testing health endpoint..."
curl http://localhost:8080/health

echo -e "\n\nTesting authentication with admin user..."
curl -X POST http://localhost:8080/auth/login \
  -d "username=admin&password=admin123"

echo -e "\n"
