.PHONY: all build clean sheep shepherd sheepctl meadow web-build dashboard

BINDIR := bin
GOFLAGS := -trimpath

# Dashboard SPA assets. web/dist is produced by the frontend worker's Vite
# build (`web/`); its contents are copied over the placeholder in
# internal/dashboard/static before shepherd is compiled with embedded assets.
DASHBOARD_STATIC := internal/dashboard/static
WEB_DIST := web/dist

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

# web-build builds the dashboard SPA. It skips gracefully (without failing the
# overall build) when the web/ directory or npm is unavailable, so the Go tree
# still builds on machines that only have the placeholder assets.
web-build:
	@if [ ! -d web ]; then \
		echo "web-build: web/ directory not present, skipping SPA build"; \
	elif ! command -v npm >/dev/null 2>&1; then \
		echo "web-build: npm not found, skipping SPA build"; \
	else \
		echo "web-build: building SPA in web/"; \
		cd web && npm install && npm run build; \
	fi

# dashboard builds the SPA, copies its output over the embedded placeholder,
# and compiles the shepherd binary with the real assets embedded.
dashboard: web-build
	@if [ -d "$(WEB_DIST)" ]; then \
		echo "dashboard: copying $(WEB_DIST)/ into $(DASHBOARD_STATIC)/"; \
		cp -R $(WEB_DIST)/. $(DASHBOARD_STATIC)/; \
	else \
		echo "dashboard: $(WEB_DIST) not found, embedding placeholder assets"; \
	fi
	go build $(GOFLAGS) -o $(BINDIR)/shepherd ./cmd/shepherd

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
