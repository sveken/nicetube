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
When used with the linked Resonite mod on import of a link an option is presented for ```Proxy Youtube Video```. When this is selected the request is sent to the server specified in the settings for the mod, and returned back in the format specified with a direct link to the .webm.

This direct link is pasted into the world instead of the original link, allowing everyone in the world to view the video at the quality you intended.

By Default the mod is set to use 720P using VP9. This offers a good balance of file size, quality uplift and minimal CPU usage. this can be increased at your own risk with the mod settings. see issues below. 1080P with VP9 should also be fine it can just take a few moments to load the video.

An option to force H264 as the codec can be set which will result in a H264 video at or closest to your set Quality value. These files seem to have lower visual quality to the VP9 versions and larger sizes but can be more reliable with high fps videos.

### Known issues in resonite
if max quality is set, it is possible to download 4K 60fps AV1 files. These do not play well in resonite.

The larger the video the longer it can initially take to download and combine the files for you. For very long videos just give it some more time.

Long 60fps videos in VP9 sometimes do not play with sound or other weird issues. Forcing h264 fixes these but H264 does reduce the visual quality and increase size.

# Hosting your own download  server.
The mod provides some preset servers in North America & Australia. However you can host your own as long as its accessible via http/https if you want **none local** users to stream the videos.

## None docker.
Latest binaries for multiple platforms can be found [on the release page](https://github.com/sveken/nicetube/releases/latest)

nicetube also depends on ffmpeg and [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) being in the same folder as itself.

1. Download the latest release for your platform into a folder you will use for nicetube
2. Download the standalone build of [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) for your platform.
3. Download the patched ffmpeg from [yt-dlps repo here](https://github.com/yt-dlp/FFmpeg-Builds) for your platform.
4. Extract the binarys for ffpmpeg, ffprobe and yt-dlp into your nicetube folder.
5. Run nicetube

Configurable Flags can found with -help

## Docker. 
Latest docker compose file can be found in the root of this repo. 

Example compose file 

```
services:
  Nicetube:
    image: ghcr.io/sveken/nicetube:main
    environment:
      maxDuration: 120 #Max Video length in minutes, default is 2 hours.
      max_video_age: 24 #Max length of time in hours videos will be cached on disk before being cleaned. Set to 0 to disable
      #cookies: "Cookies/cookies.txt" ## Uncomment this to enable cookie support, Please Read cookie info on github
    ports:
      - "8085:8085"
    #volumes:
    #Uncomment if you want to store the video files elsewhere. Otherwise it will store them inside the container, which gets wiped every recreation/update.
     # - /Mydatapth:/home/Nicetube/Videos 
     # - /PathToCookieFolder:/home/Nicetube/Cookies ##Uncomment to specify a folder containing your cookie text file
    restart: always
```

Example docker run with default settings.
```
docker run -p 8085:8085 --restart always ghcr.io/sveken/nicetube:main
```

## Cookies 
To extract a cookies file to get around age restriction or bot check you will need to sign into youtube on a computer. Then export those cookies for example using yt-dlp ``.\yt-dlp.exe --cookies-from-browser firefox --cookies cookies.txt``. 

Then place the cookies.txt file inside the folder you are mapping with docker or next to nicetube if running stand alone. Standalone flag is ``-cookie thefilehere.txt``
## Known issues / TODO

- I want the docker image to be smaller.
- Logging read out is not final.
- webUI for other uses
