#!/bin/bash

# Define the NGINX container name
CONTAINER_NAME="nginx"

# Path to access logs inside the container
ACCESS_LOG="/var/log/nginx/access.log"

# Copy logs to the host (optional, for performance optimization)
cp ./nginx/logs/access.log ./nginx_access.log

# Calculate total requests
total_requests=$(wc -l < nginx_access.log)

# Calculate successful requests (status 200)
successful_requests=$(grep ' 200 ' nginx_access.log | wc -l)

# Calculate average response time
average_response_time=$(awk '{sum+=$2} END {if (NR > 0) print sum/NR; else print 0}' nginx_access.log)

# Calculate availability
availability=$(echo "scale=2; ($successful_requests / $total_requests) * 100" | bc)

# Output results
echo "Total Requests: $total_requests"
echo "Successful Requests: $successful_requests"
echo "Availability: $availability%"
echo "Average Response Time: $average_response_time seconds"

# Clean up copied logs
rm -f nginx_access.log
