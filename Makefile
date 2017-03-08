SRCS := $(shell find . -name '*.go')

all: test

deps:
	go get -d -v ./...

updatedeps:
	go get -d -v -u -f ./...

testdeps:
	go get -d -v -t ./...

updatetestdeps:
	go get -d -v -t -u -f ./...

install: deps
	go install ./...

license: install
	update-license $(SRCS)

golint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	go vet ./...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck ./...

staticcheck: testdeps
	go get -v honnef.co/go/tools/cmd/staticcheck
	staticcheck ./...

unused: testdeps
	go get -v honnef.co/go/tools/cmd/unused
	unused ./...

lint: golint vet errcheck staticcheck unused

test: testdeps lint
	go test -race ./...

clean:
	go clean -i ./...

.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	install \
	dogfood \
	golint \
	vet \
	errcheck \
	staticcheck \
	unused \
	lint \
	test \
	clean
