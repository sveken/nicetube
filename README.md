# nicetube
#### Note this is under testing and devolopment still. 

nicetubes is designed as a simple backend for the [Resonite Youtube Proxy Mod](https://github.com/LeCloutPanda/VideoProxy) and provides remuxed youtube videos at the users desired quality settings, and then providing a simple URL direct to the video file that can be used in world and other games/applications. Depending on the location (or vpn) of the server this can be used to also resolve country restrictions imposed on users in the world you are trying to share too.

## Current Features
- Quality selection from 360,480,720,1080,1440,2160P+ preferencing 60 fps when available.
- Self cleaning video cache purges video files older then 24 hours.
- Force H264 option for weaker computers or large lobbies to help with cpu usage.
- Docker Container or various platforms support

Future goals is to build a web interface and provide a UI for easy copy/pasting for other applications.

## When combined with the Resonite Mod
When used with the linked Resonite mod on import of a link an option is presented for ```Proxy Youtube Video```. When this is selected the request is sent to the server specified in the settings for the mod, and returned back in the format specified with a direct link to the .mp4.

This direct link is pasted into the world instead of the original link, allowing everyone in the world to view the video at the quality you intended.

By Default (For reasons see below) the codec VP9 is used when available, with a fallback to AV1

An option to force H264 as the codec can be set which will result in a H264 video at or closest to your set Quality value. 

### Known issues in resonite
For some reason the H264 version of videos refuses to load when ```Stream``` is ticked in game. To resolve this the mod will auto untick stream when forcing the H264 codec.

This will mean videos with H264 forced will need to fully download for each user before it can be played.
 For this reason VP9 is the default setting.

# Hosting your own download  server.
The mod provides some preset servers however you can host your own as long as its accessible via http/https if you want **none local** users to stream the videos.

## None docker.
Latest binaries for multiple platforms can be found [on the release page](https://github.com/sveken/nicetube/releases/latest)

nicetube also depends on ffmpeg and [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) being in the same folder as itself.

1. Download the latest release for your platform into a folder you will use for nicetube
2. Download the standalone build of [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) for your platform.
3. Download the patched ffmpeg from [yt-dlps repo here](https://github.com/yt-dlp/FFmpeg-Builds) for your platform.
4. Extract the binarys for ffpmpeg and ffprobe into your nicetube folder.
5. Run nicetube

## Docker. 
Latest docker compose file can be found in the root of this repo. 

Example compose file 

```
services:
  Nicetube:
    image: ghcr.io/sveken/nicetube:main
    ports:
      - "8085:8085"
    #volumes:
    #Uncomment if you want to store the video files elsewhere. Otherwise it will store them inside the container, which gets wiped every recreation/update.
     # - /Mydatapth:/home/Nicetube/Videos 
    restart: always
```

Example docker run
```
docker run -p 8085:8085 --restart always ghcr.io/sveken/nicetube:main
```

## Known issues 

- I want the docker image to be smaller.
- Logging read out is not final