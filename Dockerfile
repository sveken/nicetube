FROM golang:latest AS base

WORKDIR /app
# Copy the source code
COPY . .
RUN go build -o /home/Nicetube/nicetube-linux-docker ./app

FROM base AS build-arm64
WORKDIR /home/Nicetube
ADD https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64 ./yt-dlp
ADD https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz ./ffmpeg-master-latest-linux.tar.xz

FROM base AS build-amd64
WORKDIR /home/Nicetube
ADD https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux ./yt-dlp
ADD https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz ./ffmpeg-master-latest-linux.tar.xz

FROM build-${TARGETARCH} AS build
WORKDIR /home/Nicetube
RUN apt-get update && apt-get install -y --no-install-recommends \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /home/Nicetube/extract \
    && tar -xf /home/Nicetube/ffmpeg-master-latest-linux.tar.xz -C /home/Nicetube/extract \
    && mv $(find /home/Nicetube/extract -type f -name ffprobe) /home/Nicetube \
    && mv $(find /home/Nicetube/extract -type f -name ffmpeg) /home/Nicetube \
    && chmod +x /home/Nicetube/yt-dlp \
    && chmod +x /home/Nicetube/nicetube-linux-docker \
    && chmod +x /home/Nicetube/ffprobe /home/Nicetube/ffmpeg \
    && rm -rf /home/Nicetube/ffmpeg-master-latest-linux.tar.xz \
    && rm -r ./extract

RUN mkdir -p /home/Nicetube/Videos /home/Nicetube/Cookies \
    && useradd -m -d /home/Nicetube -s /bin/bash container \
    && chown -R container:container /home/Nicetube

#Stage 2
FROM debian:bookworm-slim
LABEL org.opencontainers.image.authors="Sveken"
LABEL org.opencontainers.image.title="Nicetube"
LABEL org.opencontainers.image.source=https://github.com/sveken/nicetube
LABEL org.opencontainers.image.description="Official Docker image for Nicetube bundled with required dependencies"
LABEL org.opencontainers.image.licenses=GPL-3.0-or-later
RUN mkdir -p /home/Nicetube \
    && useradd -m -d /home/Nicetube -s /bin/bash container \
    && chown -R container:container /home/Nicetube
COPY --from=build /home/Nicetube /home/Nicetube 
USER container
WORKDIR /home/Nicetube
ENV HOME=/home/Nicetube

STOPSIGNAL SIGINT
HEALTHCHECK --interval=60s --timeout=10s --start-period=5s --retries=3 \
  CMD ["./nicetube-linux-docker", "-checkhealth"]
CMD ["sh", "-c", "./nicetube-linux-docker -maxDuration ${maxDuration:-120} -max-video-age ${max_video_age:-24} -addr ${addr:-:8085} -cookie ${cookies:-n} -web-panel=${web_panel:-false} -disable-ytdlp-update=${disable_ytdlp_update:-false}"]
