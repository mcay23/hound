import React, { useRef, useEffect } from "react";
import videojs from "video.js";
import Player from "video.js/dist/types/player";
import "video.js/dist/video-js.css";

// 1. Define the props interface for type safety
interface IVideoPlayerProps {
  options: any;
  onVideoEnding?: () => void;
}

const initialOptions: any = {
  controls: true,
  controlBar: {
    volumePanel: {
      inline: false,
    },
  },
  loop: false,
  sources: [
    {
      src: "http://vjs.zencdn.net/v/oceans.mp4",
      type: "video/mp4",
    },
  ],
};

function VideoPlayer({ options, onVideoEnding }: IVideoPlayerProps) {
  const videoRef = useRef<HTMLDivElement>(null);
  const playerRef = useRef<Player | null>(null);

  useEffect(() => {
    const combinedOptions = { ...initialOptions, ...options };

    if (!playerRef.current && videoRef.current) {
      const videoElement = document.createElement("video");
      videoElement.classList.add("video-js", "vjs-big-play-centered");
      videoRef.current.appendChild(videoElement);

      const player = videojs(videoElement, combinedOptions);
      playerRef.current = player;
      player.fill(true);

      const checkDuration = () => {
        const duration = player.duration();
        const currentTime = player.currentTime();
        if (!duration || !currentTime) return;
        // call if there are less than 10% or 5 minutes left on the video (and video is at least 10 mins)
        // some failed streams have short durations, so we set 60 second threshold
        if (
          duration > 60 &&
          (currentTime >= duration * 0.9 ||
            (duration > 900 && duration - currentTime <= 300))
        ) {
          if (onVideoEnding) {
            onVideoEnding();
          }
          player.off("timeupdate", checkDuration);
        }
      };
      player.on("timeupdate", checkDuration);
    }
    return () => {
      const player = playerRef.current;
      if (player && !player.isDisposed()) {
        player.dispose();
        playerRef.current = null;
      }
    };
  }, [options, onVideoEnding]);

  return <div ref={videoRef} style={{ width: "100%", height: "100%" }} />;
}

export default VideoPlayer;
