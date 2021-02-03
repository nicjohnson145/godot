.PHONY: build test

test: basic_tests apt_integration brew_integration

basic_tests:
	go test ./...

apt_integration:
	docker build -t apt_integration internal/bootstrap -f internal/bootstrap/apt_integration.Dockerfile
	docker run --rm apt_integration

brew_integration:
	docker build -t brew_integration internal/bootstrap -f internal/bootstrap/brew_integration.Dockerfile
	docker run --rm brew_integration

build:
	rm -rf build/*
	./build.sh
