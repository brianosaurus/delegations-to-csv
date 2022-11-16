all: build

build: main.go validators/validators.go delegations/delegations.go
	go build -o getData
