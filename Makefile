SRCS := $(shell find . -name '*.go')

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
	go get -v github.com/kisielk/errcheck
	go get -v honnef.co/go/tools/cmd/staticcheck
	go get -v honnef.co/go/tools/cmd/unused

.PHONY: updatetestdeps
updatetestdeps:
	go get -d -v -t -u -f ./...
	go get -u -v github.com/kisielk/errcheck
	go get -u -v honnef.co/go/tools/cmd/staticcheck
	go get -u -v honnef.co/go/tools/cmd/unused

.PHONY: install
install: deps
	go install ./...

.PHONY: license
license: install
	update-license $(SRCS)

.PHONY: golint
golint: testdeps
	go get -v github.com/golang/lint/golint
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
