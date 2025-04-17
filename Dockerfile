FROM ubuntu:24.04


RUN apt-get update && apt-get install -y \
    wget \
    unzip \
    libfontconfig1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*


RUN wget https://github.com/godotengine/godot/releases/download/4.4-stable/Godot_v4.4-stable_linux.x86_64.zip -O /tmp/godot.zip \
    && unzip /tmp/godot.zip -d /tmp/ \
    && mv /tmp/Godot_v4.4-stable_linux.x86_64 /usr/local/bin/godot \
    && chmod +x /usr/local/bin/godot \
    && rm /tmp/godot.zip

WORKDIR /app
CMD ["godot", "--headless"]
