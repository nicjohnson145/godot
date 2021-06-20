FROM golang:1.16 as intermediate
ARG TYPE="apt_integration"
WORKDIR /app
COPY . .
RUN go build
RUN go test -c -i -o test_binary -tags="apt_integration" ./internal/bootstrap

FROM debian:buster
RUN apt update
RUN apt install -y sudo
RUN useradd -ms /bin/bash newuser
RUN groupadd passwordless
RUN usermod -a -G passwordless newuser
RUN touch /var/lib/sudo/lectured/newuser
RUN echo "%passwordless ALL=(ALL:ALL) NOPASSWD: ALL" >> /etc/sudoers
USER newuser
WORKDIR /home/newuser
COPY --from=intermediate /app/test_binary .
CMD ["./test_binary"]

