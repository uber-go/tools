SRCS := $(shell find . -name '*.go')
LINTERS := \
	github.com/golang/lint/golint \
	github.com/kisielk/errcheck \
	honnef.co/go/tools/cmd/staticcheck \
	honnef.co/go/tools/cmd/unused


.PHONY: all
all: test

.PHONY: deps
deps:
	go get -d -v ./...

.PHONY: updatedeps
updatedeps:
	go get -d -v -u -f ./...

.PHONY: testdeps
testdeps:
	go get -d -v -t ./...
	go get -v $(LINTERS)

.PHONY: updatetestdeps
updatetestdeps:
	go get -d -v -t -u -f ./...
	go get -u -v $(LINTERS)

.PHONY: install
install: deps
	go install ./cmd/...

.PHONY: license
license: install
	update-license $(SRCS)

.PHONY: golint
golint: testdeps
	for file in $(SRCS); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

.PHONY: vet
vet: testdeps
	go vet ./...

.PHONY: testdeps
errcheck: testdeps
	errcheck ./...

.PHONY: staticcheck
staticcheck: testdeps
	staticcheck ./...

.PHONY: unused
unused: testdeps
	unused ./...

.PHONY: checklicense
checklicense: install
	@echo update-license --dry $(SRCS)
	@if [ -n "$$(update-license --dry $(SRCS))" ]; then \
		echo "These files need to have their license updated by running make license:"; \
		update-license --dry $(SRCS); \
		exit 1; \
	fi

.PHONY: lint
lint: golint vet errcheck staticcheck unused checklicense

.PHONY: test
test: testdeps lint
	go test -race ./...

.PHONY: clean
clean:
	go clean -i ./...
