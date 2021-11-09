FROM golang:1.17

WORKDIR /workdir
ADD . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/app cmd/*.go

EXPOSE 8080

ENTRYPOINT ["/workdir/bin/app"]
