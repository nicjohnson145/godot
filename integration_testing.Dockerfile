FROM golang:1.15.6 as intermediate
WORKDIR /app
COPY . .
RUN go build -o godot

FROM debian:buster
RUN apt update
RUN apt install -y sudo git
RUN useradd -ms /bin/bash newuser
RUN groupadd passwordless
RUN usermod -a -G passwordless newuser
RUN touch /var/lib/sudo/lectured/newuser
RUN echo "%passwordless ALL=(ALL:ALL) NOPASSWD: ALL" >> /etc/sudoers
USER newuser
WORKDIR /home/newuser
RUN mkdir -p dotfiles/templates
RUN mkdir -p .config/godot
COPY integration_config.json /home/newuser/dotfiles/config.json
RUN echo -ne "contents\nof\nsome_conf" > /home/newuser/dotfiles/templates/some_conf
RUN echo -ne "contents\nof\ndot_zshrc" > /home/newuser/dotfiles/templates/dot_zshrc
RUN echo '{"target": "test_host", "package_managers": ["apt", "git"]}' > /home/newuser/.config/godot/config.json
COPY --from=intermediate /app/godot .
CMD /bin/bash

