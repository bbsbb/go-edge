.PHONY: docker-build
docker-build:
	docker build -f package/Dockerfile --build-arg APP=$(APP) -t go-edge/$(APP) .
