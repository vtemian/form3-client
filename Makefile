.PHONY: docs
docs:
	@docker run -v $$PWD/:/docs pandoc/latex -f markdown /docs/README.md -o /docs/build/output/README.pdf

.PHONY: run
run:
	@docker-compose up

.PHONY: clean
clean:
	@echo "Stopping all"
	@docker-compose stop
	@echo "Removing all"
	@docker-compose rm -f

.PHONY: clean-%
clean-%:
	@echo "Stopping $*"
	@docker-compose stop $*
	@echo "Removing $*"
	@docker-compose rm -f $*

.PHONY: purge
purge: clean
	@echo "Removing all images"
	@docker rmi $$(docker images  | grep form3 | awk '{print $$1}')

.PHONY: purge-%
purge-%: clean-%
	@echo "Removing image for $*"
	@docker rmi $$(docker images | grep form3 | grep $* | awk '{print $$1}')

.PHONY: exec-%
exec-%:
	@echo "Welcome to $*"
	@docker-compose exec $* bash

.PHONY: deps
deps:
	mkdir -p ./bin
	test -f ./bin/golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.27.0

.PHONY: lint
lint: deps
	./bin/golangci-lint run ./pkg/...

.PHONY: fmt
fmt:
	gofmt -s -w ./pkg/

.PHONY: check-fmt
check-fmt:
	test -z $$(gofmt -l ./cmd/ ./pkg/)

.PHONY: check
check: check-fmt lint

.PHONY: tests
tests:
	ginkgo pkg/...