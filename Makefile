.PHONY: dev run build test race vet fmt check generate i18n snapshots icon integration

dev:
	go tool mygo dev

run:
	go run .

build:
	go tool mygo build

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

generate:
	go tool mygo generate

i18n:
	go tool mygo i18n check

check: test vet i18n

snapshots:
	RIVU_SNAPSHOTS="$(CURDIR)/snapshots" go test ./internal/app -run 'Test(View|AuxiliaryView)Snapshots|TestLocalizedSurfaces' -count=1

icon:
	go run ./tools/genicon

integration:
	go test -tags integration ./internal/torrent -count=1 -timeout=60s
