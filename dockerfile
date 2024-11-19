FROM debian:bookworm-slim

LABEL org.opencontainers.image.authors="Sveken"
LABEL org.opencontainers.image.title="Nicetube"
LABEL org.opencontainers.image.source=https://github.com/sveken/nicetube
LABEL org.opencontainers.image.description="Official Docker image for Nicetube bundled with required dependencies"

#Uhh i need to install stuff just to extract ffmpeg...

RUN apt-get update && apt-get install -y --no-install-recommends \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

LABEL org.opencontainers.image.licenses=GPL-3.0-or-later

RUN useradd -m -d /home/Nicetube -s /bin/bash/ container
ENV USER=container
ENV HOME=/home/Nicetube
ENV	DEBIAN_FRONTEND=noninteractive
STOPSIGNAL SIGINT
WORKDIR /home/Nicetube

ADD https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux ./yt-dlp_linux
ADD https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz ./ffmpeg-master-latest-linux64-gpl.tar.xz
ADD https://github.com/sveken/nicetube/releases/latest/download/nicetube-linux-amd64 ./nicetube-linux-amd64

RUN tar -xf /home/Nicetube/ffmpeg-master-latest-linux64-gpl.tar.xz -C /home/Nicetube \
    && mv /home/Nicetube/bin/ffprobe /home/Nicetube \
    && mv /home/Nicetube/bin/ffplay /home/Nicetube \
    && mv /home/Nicetube/bin/ffmpeg /home/Nicetube \
    && chmod +x /home/Nicetube/yt-dlp_linux \
    && chmod +x /home/Nicetube/nicetube-linux-amd64 \
    && chmod +x /home/Nicetube/ffprobe /home/Nicetube/ffplay /home/Nicetube/ffmpeg \
    && rm -rf /home/Nicetube/ffmpeg-master-latest-linux64-gpl.tar.xz /home/Nicetube/bin /home/Nicetube/*

RUN mkdir /home/Nicetube/Videos \
&& chown -R container:container /home/Nicetube/Videos

USER container
CMD ["./nicetube-linux-amd64"]