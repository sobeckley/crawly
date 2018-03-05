

SOURCES:=main.go crawly.go diffy.go

BINARY:=crawl

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	go build $(SOURCES)

all:
	go build $(SOURCES)

.PHONY: clean
clean:
	@rm -f $(BINARY)
