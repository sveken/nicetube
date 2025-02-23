FROM debian:bookworm-slim AS stage1



#Uhh i need to install stuff just to extract ffmpeg...

RUN apt-get update && apt-get install -y --no-install-recommends \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /home/Nicetube

ADD https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux ./yt-dlp
ADD https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz ./ffmpeg-master-latest-linux64-gpl.tar.xz
ADD https://github.com/sveken/nicetube/releases/latest/download/nicetube-linux-amd64 ./nicetube-linux-amd64

RUN tar -xf /home/Nicetube/ffmpeg-master-latest-linux64-gpl.tar.xz -C /home/Nicetube \
    && mv $(find /home/Nicetube -type f -name ffprobe) /home/Nicetube \
    && mv $(find /home/Nicetube -type f -name ffmpeg) /home/Nicetube \
    && chmod +x /home/Nicetube/yt-dlp \
    && chmod +x /home/Nicetube/nicetube-linux-amd64 \
    && chmod +x /home/Nicetube/ffprobe /home/Nicetube/ffmpeg \
    && rm -rf /home/Nicetube/ffmpeg-master-latest-linux64-gpl.tar.xz \
    && rm -r ./ffmpeg-master-latest-linux64-gpl

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
COPY --from=stage1 /home/Nicetube /home/Nicetube 
USER container
WORKDIR /home/Nicetube
ENV HOME=/home/Nicetube

STOPSIGNAL SIGINT
HEALTHCHECK --interval=60s --timeout=10s --start-period=5s --retries=3 \
  CMD ["./nicetube-linux-amd64", "-checkhealth"]
CMD ["sh", "-c", "./nicetube-linux-amd64 -maxDuration ${maxDuration:-120} -max-video-age ${max_video_age:-24} -addr ${addr:-:8085} -cookie ${cookies:-n}"]
