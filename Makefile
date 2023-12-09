
help: 	## show this help
	@grep -e "^[0-9a-zA-Z_-]*:.*##" $(MAKEFILE_LIST) | \
	sed 's/^\(.*\):.*##\(.*\)/\1\t\2/'

tidy:  ## update deps
	go mod tidy

build: ## build
	go build ./store/...

clean: 	## clean
	@-rm store.test

test: ## test
	go test -v -timeout=30s ./store/...

coverage:	## test with coverage
	#go test --converage ./store/...
	go test -coverprofile=coverage.txt -covermode=atomic -v -count=1 -timeout=30s -parallel=4 -failfast ./store/...


