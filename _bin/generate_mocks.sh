#!/bin/bash

set -e

for source_file in $(find vendor/github.com/aws/aws-sdk-go/service -name "interface.go" -print); do
  interface_name=$(echo ${source_file} | rev | cut -d/ -f3 | rev) # HACK(el): cut from the end
  echo "Generating mocks for ${interface_name}"
  mockgen -destination=aws/mocks/${interface_name}.go -package mocks -source ${source_file}
done
