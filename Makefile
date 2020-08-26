.PHONY: clean
clean:
	- rm actuator

actuator: clean
	go build -o actuator ./cmd/actuator
