ROLE := slot-manager
.PHONY: build test test-standalone-layout
build:
	go build -o bin/cofiswarm-slot-manager ./cmd/cofiswarm-slot-manager
test: build test-standalone-layout
test-standalone-layout:
	./test/scripts/assert-layout.sh $(ROLE)
