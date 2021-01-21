.PHONY: build test

test:
	go test ./...
	docker build -t godot_integration internal/bootstrap && docker run --rm godot_integration

build:
	rm -rf build/*
	./build.sh
