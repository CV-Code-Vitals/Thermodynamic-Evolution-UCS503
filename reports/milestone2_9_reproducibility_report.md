# Milestone 2.9: Benchmark Reproducibility Report

## 1. Deterministic Execution Parameters

- **Target Benchmark Suite:** `E:\SWE\benchmarks\self_healing_suite`
- **Ground Truth Manifest:** `E:\SWE\benchmarks\ground_truth_manifest.json`
- **Manifest SHA256:** `5f8841cc57998e9a5a0b6257cb3dde1ed4e42979cf62316c2afb114727565b5e`
- **Integration Branch:** `benchmark/milestone-2.9-integration`
- **Integration Commit:** `LOCAL_INTEGRATION_VERIFIED`

## 2. Reproduction Steps

```pwsh
# 1. Compile benchmark orchestrator
cd go-backend
go build -o bin/benchmark.exe ./cmd/benchmark

# 2. Run full 5-stage benchmark suite
./bin/benchmark.exe --repo ../benchmarks/self_healing_suite --manifest ../benchmarks/ground_truth_manifest.json --output reports
```
