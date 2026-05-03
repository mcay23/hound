import React, { useRef, useEffect } from "react";
import videojs from "video.js";
import Player from "video.js/dist/types/player";
import "video.js/dist/video-js.css";
import { SERVER_URL } from "./../../config/axios_config";

// 1. Define the props interface for type safety
interface IVideoPlayerProps {
  options: any;
  onVideoProgress?: (current: number, total: number) => void;
  setLoading?: (loading: boolean) => void;
  subtitles?: any[];
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

const VideoPlayer = React.memo(
  ({ options, onVideoProgress, setLoading, subtitles }: IVideoPlayerProps) => {
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

        player.on("loadedmetadata", () => {
          if (combinedOptions.startTime) {
            player.currentTime(combinedOptions.startTime);
            player.play();
          }
        });

        let lastReportTime = 0;
        const handleTimeUpdate = () => {
          const currentTime = player.currentTime();
          const duration = player.duration();

          // 5 seconds interval
          if (
            currentTime &&
            duration &&
            Math.abs(currentTime - lastReportTime) >= 5
          ) {
            if (onVideoProgress) {
              onVideoProgress(currentTime, duration);
            }
            lastReportTime = currentTime;
          }
        };
        player.on("timeupdate", handleTimeUpdate);
      }
      return () => {
        const player = playerRef.current;
        if (player && !player.isDisposed()) {
          player.dispose();
          playerRef.current = null;
        }
      };
    }, [options, onVideoProgress]);

    useEffect(() => {
      const player = playerRef.current;
      if (player && subtitles) {
        player.ready(function () {
          // Remove existing remote text tracks to avoid duplicates
          const remoteTracks = player.remoteTextTracks();
          if (remoteTracks) {
            const tracksArray = Array.prototype.slice.call(remoteTracks);
            tracksArray.forEach((track) => {
              player.removeRemoteTextTrack(track);
            });
          }
          subtitles.forEach((sub) => {
            const url =
              SERVER_URL +
              "/api/v1/subtitle/" +
              sub.encoded_data +
              "?convert=vtt";
            player.addRemoteTextTrack(
              {
                src: url,
                kind: "subtitles",
                srclang: sub.lang,
                label: sub.title,
              },
              false,
            );
          });
        });
      }
    }, [subtitles]);

    return <div ref={videoRef} style={{ width: "100%", height: "100%" }} />;
  },
);

export default VideoPlayer;
