.PHONY: test lint fmt

test:
	@go test -v -race ./...

lint:
	@golangci-lint run ./...

fmt:
	@goimports -w .
	@gofumpt -l -w .
	@golines --ignore-generated --chain-split-dots --max-len 120 --shorten-comments -w . 
	@gci write --skip-generated -s standard -s default -s "prefix(github.com/nikoksr/konfetty)" .
