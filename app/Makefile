.PHONY: all app migrate clean

all: app migrate

app:
	CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o build/kryvea ./cmd/kryvea

migrate:
	CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o build/kryvea-migrate ./cmd/kryvea-migrate

clean:
	rm -rf build