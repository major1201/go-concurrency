## test: Test does all pre-release works.
test:
	go mod tidy
	go fmt ./...
	go vet
	golint ./...
	go test -v ./...

.PHONY: help
all: help
## help: List all supported make commands.
help: Makefile
	@echo
	@echo "Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
