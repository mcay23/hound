import React, { useEffect, useRef, useCallback, useState } from "react";
import ElectronVideoControls from "./ElectronVideoControls";

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
    const [selectedAudioIdx, setSelectedAudioIdx] = useState<
      number | undefined
    >(0);
    const [selectedSubIdx, setSelectedSubIdx] = useState<number | undefined>(0);
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

    const handleSetAudioTrack = useCallback(async (id: number) => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.setAudioTrack(id);
        setSelectedAudioIdx(id);
      } catch (error) {
        console.error("MPV set audio track error:", error);
      }
    }, []);

    const handleSetSubtitleTrack = useCallback(async (id: number) => {
      const video = videoRef.current;
      if (!video) return;
      try {
        await video.setSubtitleTrack(id);
        setSelectedSubIdx(id);
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
        } catch (error) {
          console.error("MPV playback error:", error);
        }
      };
      console.log(tracks);

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
        if (!selectedAudioIdx && detail.audioTrack !== undefined) {
          setSelectedAudioIdx(Number(detail.audioTrack));
        }
        if (!selectedSubIdx && detail.subTrack !== undefined) {
          setSelectedSubIdx(Number(detail.subTrack));
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

    const audioTracks = (tracks || []).filter(
      (t) => t.type === "audio" || t.type === "a",
    );
    const subTracks = (tracks || []).filter(
      (t) => t.type === "sub" || t.type === "s",
    );
    const activeAudioTrack = audioTracks.find((t) => t.id === selectedAudioIdx);
    const activeSubTrack = subTracks.find((t) => t.id === selectedSubIdx);

    const playerSettings = {
      player: "desktop",
      resize_mode: "contain",
      audio_idx: selectedAudioIdx,
      audio_lang: activeAudioTrack?.lang?.toLowerCase() || "",
      subtitle_idx: selectedSubIdx,
      subtitle_lang: activeSubTrack?.lang?.toLowerCase() || "",
    };

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
        <ElectronVideoControls
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
          audioTracks={audioTracks}
          subTracks={subTracks}
          selectedAudioIdx={selectedAudioIdx}
          selectedSubIdx={selectedSubIdx}
        />
      </div>
    );
  },
);

export default MPVElectronPlayer;
