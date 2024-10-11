#!/bin/bash

init=${INIT:-true}
plan=${PLAN:-true}
echo "Cloning repository from: $MODULEPATH"
git clone $MODULEPATH
REPO_NAME=$(basename $MODULEPATH .git)

echo "Changing into directory: $REPO_NAME"
cd "$REPO_NAME" || { echo "Failed to enter directory: $REPO_NAME"; exit 1; }

if [ "$init" = true ]; then
  echo "Running terraform init"
  terraform init
fi
if [ "$plan" = true ]; then
  echo "Running terraform plan"
  terraform plan
fi

echo "Running terraform apply"
terraform apply -auto-approve