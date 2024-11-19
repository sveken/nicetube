FROM debian:bookworm-slim

LABEL org.opencontainers.image.authors="Sveken"
LABEL org.opencontainers.image.title="Nicetube"
LABEL org.opencontainers.image.source=https://github.com/sveken/nicetube
LABEL org.opencontainers.image.description="Official Docker image for Nicetube bundled with required dependencies"
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

RUN tar -xf ./ffmpeg-master-latest-linux64-gpl.tar.xz --strip-components=1 -C /home/Nicetube \
&& chmod +x /home/Nicetube/yt-dlp_linux \
&& chmod +x /home/Nicetube/nicetube-linux-amd64 \
&& rm /home/Nicetube/ffmpeg-master-latest-linux64-gpl.tar.xz

RUN mkdir /home/Nicetube/Videos \
&& chown -R container:container /home/Nicetube/Videos

USER container
CMD ["./nicetube-linux-amd64"]