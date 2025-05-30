#!/bin/bash

# Number of worker instances to launch
num_workers=$1

for ((i=1; i<=num_workers; i++))
do
    echo "Starting translate_worker instance $i"
    go run translate_worker.go > "logs/worker_translate_$i.log" 2>&1 &
done


# Wait for all background jobs to complete
wait