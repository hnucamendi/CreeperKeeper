#!/bin/bash

# Directory containing the Lambda functions
LAMBDAS_DIR="./BE-CreeperKeeper/functions"

# Initialize an array to hold the Lambda info
lambdas_info=()

# Iterate over each Lambda function directory
for dir in "$LAMBDAS_DIR"/*/; do
  if [ -f "$dir/go.mod" ]; then
    lambda_name=$(basename "$dir")
    go_version=$(grep '^go ' "$dir/go.mod" | awk '{print $2}')
    lambdas_info+=("{\"name\":\"$lambda_name\",\"go_version\":\"$go_version\"}")
  fi
done

# Convert the array to JSON
lambdas_json=$(printf ",%s" "${lambdas_info[@]}")
lambdas_json="[${lambdas_json:1}]"

# Print the JSON
echo "$lambdas_json"
