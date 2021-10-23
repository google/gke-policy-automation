GOCMD=go
TEST?=$$(go list ./... |grep -v 'vendor')

default: build test

build:
	${GOCMD} build

test:
	echo $(TEST) | \
		xargs -t ${GOCMD} test -v

clean:
	${GOCMD} clean

.PHONY: test clean