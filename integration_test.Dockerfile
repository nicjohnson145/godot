FROM ubuntu:20.04

RUN apt update && apt install -y ca-certificates sudo

RUN useradd -ms /bin/bash newuser
RUN groupadd passwordless
RUN usermod -a -G passwordless newuser
RUN echo "newuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER newuser
WORKDIR /home/newuser
RUN mkdir -p /home/newuser/.config/godot
COPY integration_test_resources/config.yaml /home/newuser/.config/godot/config.yaml
COPY integration_test_resources/test.sh /scripts/test.sh
COPY godot /bin/godot

CMD ["/scripts/test.sh"]
