.PHONY: clean executor

builder:
	- rm builder
	go build -o builder ./cmd/builder


evaluator:
	- rm evaluator.exe
	go env -w GOOS=windows
	go env -w GOARCH=amd64
	go build -o evaluator.exe ./cmd/evaluator/cs303

executor:
	- rm executor
	go env -w GOOS=linux
	go env -w GOARCH=amd64
	go build -o executor ./cmd/executor/cs303
	scp executor ai:~/proj3-executor
