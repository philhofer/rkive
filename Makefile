install:
	@go install ./...

test:
	@go test -v

test-all:
	@go get gopkg.in/check.v1
	@go test -v -tags 'riak' -check.v

ci-deps:
	go get gopkg.in/check.v1
	go get code.google.com/p/gogoprotobuf/proto
	go get -d ./...

continuous: install test-all
	