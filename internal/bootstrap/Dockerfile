FROM golang:1.15.6 as intermediate
ARG TYPE="apt_integration"
WORKDIR /app
COPY . .
RUN go test -c -o test_binary -tags="$TYPE" -i

FROM debian:buster-slim
RUN apt update
WORKDIR /test
COPY --from=intermediate /app/test_binary .
CMD ["./test_binary"]
