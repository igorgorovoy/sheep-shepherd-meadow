.PHONY: all build clean sheep shepherd sheepctl meadow

BINDIR := bin
GOFLAGS := -trimpath

all: build

build: sheep shepherd sheepctl meadow

sheep:
	go build $(GOFLAGS) -o $(BINDIR)/sheep ./cmd/sheep

shepherd:
	go build $(GOFLAGS) -o $(BINDIR)/shepherd ./cmd/shepherd

sheepctl:
	go build $(GOFLAGS) -o $(BINDIR)/sheepctl ./cmd/sheepctl

meadow:
	go build $(GOFLAGS) -o $(BINDIR)/meadow ./cmd/meadow

clean:
	rm -rf $(BINDIR)

install: build
	cp $(BINDIR)/sheep /usr/local/bin/
	cp $(BINDIR)/shepherd /usr/local/bin/
	cp $(BINDIR)/sheepctl /usr/local/bin/
	cp $(BINDIR)/meadow /usr/local/bin/

test:
	go test ./...

lint:
	golangci-lint run ./...
