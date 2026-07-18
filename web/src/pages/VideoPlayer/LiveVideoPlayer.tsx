import React, { useEffect, useRef, useState } from "react";
import videojs from "video.js";
import Player from "video.js/dist/types/player";
import "video.js/dist/video-js.css";

export default function LiveVideoPlayer({ src }: { src: string }) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const playerRef = useRef<Player | null>(null);
  useEffect(() => {
    if (!containerRef.current) return;
    if (!playerRef.current) {
      const videoElement = document.createElement("video-js");
      containerRef.current.appendChild(videoElement);
      playerRef.current = videojs(videoElement, {
        controls: true,
        responsive: true,
        fluid: true,
        preload: "auto",
        autoplay: true,
        liveui: true,
        sources: [
          {
            src,
            type: "application/x-mpegURL",
          },
        ],
      });
    } else {
      playerRef.current.src({
        src,
        type: "application/x-mpegURL",
      });
    }
  }, [src]);

  useEffect(() => {
    return () => {
      if (playerRef.current) {
        playerRef.current.dispose();
        playerRef.current = null;
      }
    };
  }, []);

  return <div ref={containerRef} />;
}
