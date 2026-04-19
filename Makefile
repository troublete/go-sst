.PHONY: bench bench-orchestration
bench:
	cd sst && go test -bench=. -cpu=1,2,4,8 -benchmem -test.bench BenchmarkLookup10
	cd sst && go test -bench=. -cpu=1,2,4,8 -benchmem -test.bench BenchmarkLookup20
bench-orchestration:
	cd sst && go test -bench=. -cpu=1,2,4,8 -benchmem -test.bench BenchmarkOrchestration
