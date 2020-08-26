.PHONY: clean
clean:
	- rm actuator
	- rm executor

actuator: clean
	go build -o actuator ./cmd/actuator

executor: clean
	go build -o executor ./cmd/executor
