install:
	go install ./...

test:
	go test -v

test-all:
	go get gopkg.in/check.v1
	go test -v -tags 'riak' -check.v

ci-deps:
	echo no deps

continuous: install test-all
	