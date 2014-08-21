install:
	go install ./...

test-all:
	go test -v -tags 'riak' -check.v

CI: install test-all
