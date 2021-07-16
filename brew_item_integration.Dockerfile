FROM golang:1.16 as intermediate
WORKDIR /app
COPY . .
RUN go build
RUN go test -c -o test_binary -tags="brew_integration" ./internal/bootstrap

FROM homebrew/ubuntu20.04:3.0.2
COPY --from=intermediate /app/test_binary .
CMD ["./test_binary"]
