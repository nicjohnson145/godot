FROM ubuntu:20.04
RUN apt update && apt install -y ca-certificates
RUN mkdir -p /root/.config/godot
COPY integration_test_resources/config.yaml /root/.config/godot/config.yaml
COPY integration_test_resources/test.sh /scripts/test.sh
COPY godot /bin/godot

CMD ["/scripts/test.sh"]
