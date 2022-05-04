.PHONY: %

build:
	CGO_ENABLED=0 go build

integration_test: build
	@if [ -z "${GITHUB_PAT}" ]; then echo "GITHUB_PAT not set, integration tests can't run"; exit 1; fi
	docker build . -t godot_integration -f integration_test.Dockerfile
	@docker run -e GITHUB_PAT=${GITHUB_PAT} --rm godot_integration 
