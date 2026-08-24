import React, { useEffect, useRef, useCallback, useState } from "react";
import VideoControls from "./VideoControls";

interface IVideoPlayerProps {
  options: any;
  onVideoProgress?: (current: number, total: number) => void;
  setLoading?: (loading: boolean) => void;
  subtitles?: any[];
  handleClose?: () => void;
  setInfoModalOpen?: (open: boolean) => void;
}

const MPVElectronPlayer = React.memo(
  ({
    options,
    onVideoProgress,
    handleClose,
    setInfoModalOpen,
  }: IVideoPlayerProps) => {
    const videoRef = useRef<any>(null);
    const lastReportTimeRef = useRef(0);

    const [paused, setPaused] = useState(true);
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);
    const [tracks, setTracks] = useState<any[]>([]);
    const [selectedAudio, setSelectedAudio] = useState<string | number>("auto");
    const [selectedSub, setSelectedSub] = useState<string | number>("auto");
    const [volume, setVolumeState] = useState(100);
    const [muted, setMuted] = useState(false);
    const [prevVolume, setPrevVolume] = useState(100);
    const playingCountRef = useRef(0);
    const seekDoneRef = useRef(false);

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

    const handleSetVolume = useCallback(async (val: number) => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.setVolume(val);
        setVolumeState(val);
        if (val > 0) setMuted(false);
      } catch (error) {
        console.error("MPV set volume error:", error);
      }
    }, []);

    const handleToggleMute = useCallback(async () => {
      const video = videoRef.current;
      if (!video) return;
      try {
        if (muted) {
          const restore = prevVolume > 0 ? prevVolume : 80;
          await video.setVolume(restore);
          setVolumeState(restore);
          setMuted(false);
        } else {
          setPrevVolume(volume);
          await video.setVolume(0);
          setVolumeState(0);
          setMuted(true);
        }
      } catch (error) {
        console.error("MPV mute error:", error);
      }
    }, [muted, volume, prevVolume]);

    const handleSetAudioTrack = useCallback(async (id: string | number) => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.setAudioTrack(id);
        setSelectedAudio(id);
      } catch (error) {
        console.error("MPV set audio track error:", error);
      }
    }, []);

    const handleSetSubtitleTrack = useCallback(async (id: string | number) => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.setSubtitleTrack(id);
        setSelectedSub(id);
      } catch (error) {
        console.error("MPV set subtitle track error:", error);
      }
    }, []);

    const handleAddSubTrack = useCallback(
      async (url: string, select = true) => {
        const video = videoRef.current;
        if (!video) return;
        try {
          await video.addSubTrack(url, select);
        } catch (error) {
          console.error("MPV add subtitle track error:", error);
        }
      },
      [],
    );

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
          try {
            const initialTracks = await video.getTrackList();
            if (initialTracks && initialTracks.length) {
              setTracks(initialTracks);
            }
          } catch (e) {
            // ignore initial track load error
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
        const current = detail.time;
        const dur = detail.duration;
        if (detail.status === "Paused") {
          setPaused(true);
        } else {
          setPaused(false);
        }
        if (detail.status) {
          console.log("status", detail.status);
        }

        // second Playing event seems more safe for initial seek
        if (detail.status === "Playing") {
          playingCountRef.current++;
          if (playingCountRef.current >= 2 && !seekDoneRef.current) {
            seekDoneRef.current = true;
            if (startTime && startTime > 0) {
              setTimeout(async () => {
                if (destroyed) return;
                try {
                  await video.seek(startTime);
                } catch (err) {
                  console.warn("MPV initial seek notice (retrying):", err);
                  setTimeout(async () => {
                    if (!destroyed) {
                      try {
                        await video.seek(startTime);
                      } catch (e) {}
                    }
                  }, 1000);
                }
              }, 100);
            }
          }
        }
        if (typeof current === "number") {
          setCurrentTime(current);
        }
        if (typeof dur === "number" && dur > 0) {
          setDuration(dur);
        }
        if (Array.isArray(detail.trackList)) {
          setTracks(detail.trackList);
        }
        if (detail.audioTrack !== undefined) {
          setSelectedAudio(detail.audioTrack);
        }
        if (detail.subTrack !== undefined) {
          setSelectedSub(detail.subTrack);
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
          handleSetAudioTrack={handleSetAudioTrack}
          handleSetSubtitleTrack={handleSetSubtitleTrack}
          handleSetVolume={handleSetVolume}
          handleToggleMute={handleToggleMute}
          handleClose={handleClose}
          setInfoModalOpen={setInfoModalOpen}
          paused={paused}
          currentTime={currentTime}
          duration={duration}
          volume={volume}
          muted={muted}
          tracks={tracks}
          selectedAudio={selectedAudio}
          selectedSub={selectedSub}
        />
      </div>
    );
  },
);

export default MPVElectronPlayer;
