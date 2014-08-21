install:
	go install ./...

test:
	go test -v

test-all:
	go get gopkg.in/check.v1
	go test -v -tags 'riak' -check.v

CI: install test-all
	