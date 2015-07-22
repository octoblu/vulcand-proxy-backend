test:
	go test -v ./backend

cover:
	go test -v ./backend  -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out
