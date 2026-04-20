.PHONY: bench bench-orchestration
bench:
	cd sst && go test -bench=. -cpu=1,2,4,8 -benchmem -test.bench BenchmarkLookup
	cd sst && go test -bench=. -cpu=1,2,4,8 -benchmem -test.bench BenchmarkIterate
	cd sst && go test -bench=. -cpu=1,2,4,8 -benchmem -test.bench BenchmarkResponseAllocation
bench-orchestration:
	cd sst && go test -bench=. -cpu=1,2,4,8 -benchmem -test.bench BenchmarkOrchestration
