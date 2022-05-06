.PHONY: %

go_build:
	CGO_ENABLED=0 GOOS=linux go build

check_pat_set:
	@if [ -z "${GITHUB_PAT}" ]; then echo "GITHUB_PAT not set, integration tests can't run"; exit 1; fi

docker_build:
	docker build . -t godot_integration -f integration_test.Dockerfile

integration_test: go_build check_pat_set docker_build 
	@docker run -e GITHUB_PAT=${GITHUB_PAT} --rm godot_integration 

integration_debug: go_build check_pat_set docker_build 
	@docker run -e GITHUB_PAT=${GITHUB_PAT} -it --rm godot_integration /bin/bash
