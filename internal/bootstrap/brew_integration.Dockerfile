FROM golang:1.15.6 as intermediate
ARG TYPE="brew_integration"
WORKDIR /app
COPY . .
RUN go test -c -o test_binary -tags="$TYPE" -i

FROM homebrew/ubuntu20.04
COPY --from=intermediate /app/test_binary .
CMD ["./test_binary"]
