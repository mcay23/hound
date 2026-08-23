import React, { useEffect, useRef, useCallback, useState } from "react";
import VideoControls from "./VideoControls";

interface IVideoPlayerProps {
  options: any;
  onVideoProgress?: (current: number, total: number) => void;
  setLoading?: (loading: boolean) => void;
  subtitles?: any[];
}

const MPVElectronPlayer = React.memo(
  ({ options, onVideoProgress }: IVideoPlayerProps) => {
    const videoRef = useRef<any>(null);
    const lastReportTimeRef = useRef(0);

    const [paused, setPaused] = useState(true);
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);

    const handlePlayPause = async () => {
      const video = videoRef.current;
      if (!video) return;
      try {
        if (paused) {
          await video.play();
        } else {
          await video.pause();
        }
      } catch (error) {
        console.error("MPV pause error:", error);
      }
    };

    const handlePause = useCallback(async () => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.pause();
      } catch (error) {
        console.error("MPV pause error:", error);
      }
    }, []);

    const handlePlay = useCallback(async () => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.play();
      } catch (error) {
        console.error("MPV play error:", error);
      }
    }, []);

    const handleSeek = useCallback(async (time: number) => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.seek(time);
        setCurrentTime(time);
      } catch (error) {
        console.error("MPV seek error:", error);
      }
    }, []);

    useEffect(() => {
      const video = videoRef.current;
      if (!video) {
        return;
      }
      let destroyed = false;
      const source = options?.sources?.[0]?.src;
      if (!source) {
        console.warn("MPV: no source specified");
        return;
      }
      const startTime = options?.startTime;
      const load = async () => {
        try {
          await video.open(source);
          if (destroyed) {
            return;
          }
          await video.play();
          if (startTime && startTime > 0) {
            setTimeout(async () => {
              if (destroyed) return;
              try {
                await video.seek(startTime);
              } catch (error) {
                console.error("MPV seek error:", error);
              }
            }, 500);
          }
        } catch (error) {
          console.error("MPV playback error:", error);
        }
      };

      load();

      const handleState = (event: Event) => {
        if (destroyed) return;
        const detail = (event as CustomEvent).detail;
        if (!detail) return;
        const current = detail.currentTime;
        const dur = detail.duration;
        if (detail.status === "Playing") {
          setPaused(false);
        } else {
          setPaused(true);
        }
        if (typeof current === "number") {
          setCurrentTime(current);
        }
        if (typeof dur === "number" && dur > 0) {
          setDuration(dur);
        }
        if (
          typeof current === "number" &&
          typeof dur === "number" &&
          Math.abs(current - lastReportTimeRef.current) >= 5
        ) {
          lastReportTimeRef.current = current;
          onVideoProgress?.(current, dur);
        }
      };

      video.addEventListener("mpv-state", handleState);

      return () => {
        destroyed = true;
        video.removeEventListener("mpv-state", handleState);
        try {
          video.stop?.();
        } catch (error) {
          console.warn("MPV stop failed:", error);
        }
        try {
          video.destroy?.();
        } catch (error) {
          console.warn("MPV destroy failed:", error);
        }
      };
    }, [options?.sources?.[0]?.src, options?.startTime, onVideoProgress]);

    return (
      <div className="video-container">
        <mpv-video
          ref={videoRef}
          render-mode="shared-texture"
          volume="100"
          style={{
            width: "100%",
            height: "100%",
            display: "block",
          }}
        />
        <VideoControls
          handlePause={handlePause}
          handlePlay={handlePlay}
          handlePlayPause={handlePlayPause}
          handleSeek={handleSeek}
          paused={paused}
          currentTime={currentTime}
          duration={duration}
        />
      </div>
    );
  },
);

export default MPVElectronPlayer;
