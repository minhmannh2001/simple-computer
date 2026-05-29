all:
	go build -o bin/simulator simple-computer/cmd/simulator
	go build -o bin/assembler simple-computer/cmd/assembler
	go build -o bin/generator simple-computer/cmd/generator

programs: all
	$(MAKE) -C _programs

test:
	go test ./...

clean:
	rm -rf bin/
