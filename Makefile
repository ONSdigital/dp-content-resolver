build:
	go build -o build/dp-content-resolver

debug: build
	HUMAN_LOG=1 ./build/dp-content-resolver

.PHONY: build debug
