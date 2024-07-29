set GOOS=linux
set GOARCH=amd64
go build -o myapp

set GOOS=
set GOARCH=
go build .