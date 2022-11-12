build:
	echo "Building GluBNG CLI..."
	go build -o bin/glubng cmd/cli/main.go
	echo "Building GluBNG Server..."
	go build -o bin/glubngd cmd/server/main.go

all: build

clean: 
	rm -f bin/glubngd
	rm -f bin/glubng