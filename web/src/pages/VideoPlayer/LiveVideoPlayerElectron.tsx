import React, { useEffect, useRef, useCallback, useState } from "react";
import ElectronVideoControls from "./ElectronVideoControls";
import toast from "react-hot-toast";

const MPVElectronPlayer = React.memo(({ src }: { src: string }) => {
  const videoRef = useRef<any>(null);
  const lastReportTimeRef = useRef(0);

  const [paused, setPaused] = useState(true);
  const [tracks, setTracks] = useState<any[]>([]);
  const [selectedAudioIdx, setSelectedAudioIdx] = useState<number | undefined>(
    undefined,
  );
  const [selectedSubIdx, setSelectedSubIdx] = useState<number | undefined>(
    undefined,
  );
  const selectedAudioIdxRef = useRef<number | undefined>(undefined);
  const selectedSubIdxRef = useRef<number | undefined>(undefined);
  const originalFullscreen = useRef<boolean | undefined>(undefined);

  const [volume, setVolumeState] = useState(100);
  const [muted, setMuted] = useState(false);
  const [prevVolume, setPrevVolume] = useState(100);
  const tracksInitializedRef = useRef(false);

  const divRef = useRef<HTMLDivElement>(null);
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
      selectedAudioIdxRef.current = id;
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
      selectedSubIdxRef.current = id;
    } catch (error) {
      console.error("MPV set subtitle track error:", error);
    }
  }, []);

  const audioTracks = (tracks || []).filter(
    (t) => t.type === "audio" || t.type === "a",
  );
  const subTracks = (tracks || []).filter(
    (t) => t.type === "sub" || t.type === "s",
  );

  const initializeTracks = useCallback(async (trackList: any[]) => {
    const video = videoRef.current;
    if (!video || tracksInitializedRef.current) return;
    tracksInitializedRef.current = true;
  }, []);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) {
      return;
    }
    let destroyed = false;
    if (!src) {
      console.warn("MPV: no source specified");
      return;
    }
    const load = async () => {
      try {
        await video.open(src);
        if (destroyed) {
          return;
        }
        await video.play();
      } catch (error) {
        console.error("MPV playback error:", error);
        toast.error(`MPV playback error: ${error}`);
      }
    };

    load();

    const handleState = (event: Event) => {
      if (destroyed) return;
      const detail = (event as CustomEvent).detail;
      if (!detail) return;
      if (detail.status === "Paused") {
        setPaused(true);
      } else {
        setPaused(false);
      }
      // second Playing event seems more safe for initial seek
      // if (detail.status === "Playing") {
      //   playingCountRef.current++;
      //   if (playingCountRef.current >= 2 && !seekDoneRef.current) {
      //     seekDoneRef.current = true;

      //   }
      // }
      if (Array.isArray(detail.trackList)) {
        setTracks(detail.trackList);
        if (!tracksInitializedRef.current && detail.trackList.length > 0) {
          initializeTracks(detail.trackList);
        }
      }
      if (
        selectedAudioIdxRef.current === undefined &&
        detail.audioTrack !== undefined
      ) {
        const aIdx = Number(detail.audioTrack);
        setSelectedAudioIdx(aIdx);
        selectedAudioIdxRef.current = aIdx;
      }
      if (
        selectedSubIdxRef.current === undefined &&
        detail.subTrack !== undefined
      ) {
        const sIdx = Number(detail.subTrack);
        setSelectedSubIdx(sIdx);
        selectedSubIdxRef.current = sIdx;
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
  }, [src, initializeTracks]);

  const handleFullscreen = () => {
    if (divRef.current) {
      if (!document.fullscreenElement) {
        divRef.current.requestFullscreen().catch((err) => {
          console.error(
            `Error attempting to enable fullscreen: ${err.message}`,
          );
        });
      } else {
        document.exitFullscreen();
      }
    }
  };

  return (
    <div className="video-container" ref={divRef}>
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
        handleSetAudioTrack={handleSetAudioTrack}
        handleSetSubtitleTrack={handleSetSubtitleTrack}
        handleSetVolume={handleSetVolume}
        handleToggleMute={handleToggleMute}
        handleFullscreen={handleFullscreen}
        paused={paused}
        volume={volume}
        muted={muted}
        audioTracks={audioTracks}
        subTracks={subTracks}
        selectedAudioIdx={selectedAudioIdx}
        selectedSubIdx={selectedSubIdx}
        streamType="live"
      />
    </div>
  );
});

export default MPVElectronPlayer;
