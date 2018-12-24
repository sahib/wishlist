build:
	go install wedlist.go

run:
	go run wedlist.go

test:
	# New go test ignores vendor/
	go test -v ./...

lint:
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec gofmt -s -w {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec go fix {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec sh -c 'golint {} | grep -v unexported' \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec misspell {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec gocyclo -over 20 {} \; | sort -n
	gosec -exclude=G104 -quiet -fmt json ./... | jq '.Issues[] | select(.file | contains("capnp.go") | not)'
