all:
	CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o build/kryvea cmd/kryvea/main.go

clean:
	rm -rf build