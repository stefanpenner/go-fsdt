#!/usr/bin/env bash
set -euo pipefail

# Runs selected benchmarks and prints a simple summary table

echo "Running benchmarks..." >&2
RAW=$(go test -bench='Benchmark_(Traversal|Diff_Basic|Hash_NoXAttr|Hash_WithSidecar)$' -run=^$ -benchmem ./... | sed -n 's/^Benchmark_/Benchmark_/p')

printf "\nSummary (ns/op, B/op, allocs/op)\n"
printf "%-28s %12s %10s %11s\n" "Benchmark" "ns/op" "B/op" "allocs/op"

echo "$RAW" | while read -r line; do
  name=$(echo "$line" | awk '{print $1}')
  ns=$(echo "$line" | awk '{print $3}')
  b=$(echo "$line" | awk '{print $5}')
  allocs=$(echo "$line" | awk '{print $7}')
  printf "%-28s %12s %10s %11s\n" "$name" "$ns" "$b" "$allocs"
done