FROM debian:bookworm-slim

LABEL org.opencontainers.image.authors="Sveken"
LABEL org.opencontainers.image.title="Nicetube"
LABEL org.opencontainers.image.source=https://github.com/sveken/nicetube
LABEL org.opencontainers.image.description="Official Docker image for Nicetube bundled with required dependencies"
LABEL org.opencontainers.image.licenses=GPL-3.0-or-later

USER container
ENV USER=container
ENV HOME=/home/Nicetube
ENV	DEBIAN_FRONTEND=noninteractive
STOPSIGNAL SIGINT
WORKDIR /home/Nicetube

ADD https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux /home/Nicetube
ADD https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz /home/Nicetube
ADD https://github.com/sveken/nicetube/releases/latest/download/nicetube-linux-amd64 /home/Nicetube/

RUN RUN mkdir /home/Nicetube/Videos \
&& chown -R container:container /home/Nicetube/Videos

CMD nicetube-linux-amd64 