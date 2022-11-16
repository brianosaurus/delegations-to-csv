all: build

build: main.go validators/validators.go
	go build -o getValidators
