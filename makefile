

SOURCES:=main.go fetch.go

BINARY:=crawl

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	go build $(SOURCES)

all:
	go build $(SOURCES)

.PHONY: clean
clean:
	@rm -f $(BINARY)
