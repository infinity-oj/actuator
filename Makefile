.PHONY: clean
clean:
	- rm builder
	- rm executor

builder: clean
	go build -o builder ./cmd/builder

executor: clean
	go build -o executor ./cmd/executor
