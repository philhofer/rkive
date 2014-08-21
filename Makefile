install:
	go install ./...

test:
	go test -v

test-all:
	go test -v -tags 'riak' -check.v

CI: install test-all