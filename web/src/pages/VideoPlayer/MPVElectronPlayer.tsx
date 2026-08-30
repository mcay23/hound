import React, { useEffect, useRef, useCallback, useState } from "react";
import ElectronVideoControls from "./ElectronVideoControls";
import { get2LetterLangCode } from "../../helpers/locale";

interface IVideoPlayerProps {
  options: any;
  onVideoProgress?: (
    current: number,
    total: number,
    playerSettings?: any,
  ) => void;
  setLoading?: (loading: boolean) => void;
  externalSubtitles?: any[];
  handleClose?: () => void;
  setInfoModalOpen?: (open: boolean) => void;
  playerSettings?: any;
  isStreamsMatch?: boolean;
  originalAudioLang?: string;
}

const MPVElectronPlayer = React.memo(
  ({
    options,
    onVideoProgress,
    externalSubtitles,
    handleClose,
    setInfoModalOpen,
    playerSettings,
    isStreamsMatch,
    originalAudioLang,
  }: IVideoPlayerProps) => {
    const videoRef = useRef<any>(null);
    const lastReportTimeRef = useRef(0);

    const [paused, setPaused] = useState(true);
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);
    const [tracks, setTracks] = useState<any[]>([]);
    const [selectedAudioIdx, setSelectedAudioIdx] = useState<
      number | undefined
    >(undefined);
    const [selectedSubIdx, setSelectedSubIdx] = useState<number | undefined>(
      undefined,
    );
    const selectedAudioIdxRef = useRef<number | undefined>(undefined);
    const selectedSubIdxRef = useRef<number | undefined>(undefined);
    const originalFullscreen = useRef<boolean | undefined>(undefined);

    const playerSettingsRef = useRef(playerSettings);
    playerSettingsRef.current = playerSettings;
    const isStreamsMatchRef = useRef(isStreamsMatch);
    isStreamsMatchRef.current = isStreamsMatch;
    const originalAudioLangRef = useRef(originalAudioLang);
    originalAudioLangRef.current = originalAudioLang;
    const onVideoProgressRef = useRef(onVideoProgress);
    onVideoProgressRef.current = onVideoProgress;

    const [volume, setVolumeState] = useState(100);
    const [muted, setMuted] = useState(false);
    const [prevVolume, setPrevVolume] = useState(100);
    const playingCountRef = useRef(0);
    const seekDoneRef = useRef(false);
    const tracksInitializedRef = useRef(false);

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

      const availAudio = trackList.filter(
        (t) => t.type === "audio" || t.type === "a",
      );
      const availSub = trackList.filter(
        (t) => t.type === "sub" || t.type === "s",
      );

      const curSettings = playerSettingsRef.current;
      const isMatch = isStreamsMatchRef.current;
      const origAudio = originalAudioLangRef.current;

      // Audio Track Restoration
      let targetAudio: number | undefined = undefined;
      if (
        isMatch &&
        curSettings?.audio_idx !== undefined &&
        curSettings?.audio_idx !== null
      ) {
        const exists = availAudio.find(
          (t) => Number(t.id) === Number(curSettings.audio_idx),
        );
        if (exists) {
          targetAudio = Number(curSettings.audio_idx);
        }
      }
      if (targetAudio === undefined) {
        const targetLang =
          get2LetterLangCode(curSettings?.audio_lang) ||
          get2LetterLangCode(origAudio);
        if (targetLang) {
          const match = availAudio.find(
            (t) => get2LetterLangCode(t.lang) === targetLang,
          );
          if (match) {
            targetAudio = Number(match.id);
          }
        }
      }
      if (targetAudio !== undefined) {
        try {
          await video.setAudioTrack(targetAudio);
          setSelectedAudioIdx(targetAudio);
          selectedAudioIdxRef.current = targetAudio;
        } catch (e) {
          console.error("Failed to restore audio track:", e);
        }
      }

      // Subtitle Track Restoration
      let targetSub: number | undefined = undefined;
      if (
        isMatch &&
        curSettings?.subtitle_idx !== undefined &&
        curSettings?.subtitle_idx !== null
      ) {
        const subIdxNum = Number(curSettings.subtitle_idx);
        if (subIdxNum === 0) {
          targetSub = 0;
        } else {
          const exists = availSub.find((t) => Number(t.id) === subIdxNum);
          if (exists) {
            targetSub = subIdxNum;
          }
        }
      }
      if (targetSub === undefined) {
        const targetLang =
          get2LetterLangCode(curSettings?.subtitle_lang) || "en";
        const match = availSub.find(
          (t) => get2LetterLangCode(t.lang) === targetLang,
        );
        if (match) {
          targetSub = Number(match.id);
        }
      }
      if (targetSub !== undefined) {
        try {
          await video.setSubtitleTrack(targetSub);
          setSelectedSubIdx(targetSub);
          selectedSubIdxRef.current = targetSub;
        } catch (e) {
          console.error("Failed to restore subtitle track:", e);
        }
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
        if (
          typeof current === "number" &&
          typeof dur === "number" &&
          Math.abs(current - lastReportTimeRef.current) >= 5
        ) {
          lastReportTimeRef.current = current;
          const currentAudio =
            detail.audioTrack ?? selectedAudioIdxRef.current ?? 0;
          const currentSub = detail.subTrack ?? selectedSubIdxRef.current ?? 0;

          const currentAudioTrack = (detail.trackList || []).find(
            (t: any) =>
              (t.type === "audio" || t.type === "a") &&
              Number(t.id) === Number(currentAudio),
          );
          const currentSubTrack = (detail.trackList || []).find(
            (t: any) =>
              (t.type === "sub" || t.type === "s") &&
              Number(t.id) === Number(currentSub),
          );

          const playerSettingsPayload = {
            player: "desktop",
            resize_mode: "contain",
            audio_idx: Number(currentAudio),
            audio_lang: get2LetterLangCode(currentAudioTrack?.lang),
            subtitle_idx: Number(currentSub),
            subtitle_lang: get2LetterLangCode(currentSubTrack?.lang),
          };

          onVideoProgressRef.current?.(current, dur, playerSettingsPayload);
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
    }, [options?.sources?.[0]?.src, options?.startTime, initializeTracks]);

    // if fullscreen state was changed by the player, revert back to original
    useEffect(() => {
      async function updateFullscreen() {
        const isFullscreen = await window?.electron?.isFullscreen();
        originalFullscreen.current = isFullscreen;
      }
      if (originalFullscreen.current === undefined) {
        updateFullscreen();
      }
      async function handleExit() {
        const isFullscreen = await window?.electron?.isFullscreen();
        if (originalFullscreen.current !== isFullscreen) {
          window?.electron?.toggleFullscreen();
        }
      }
      return () => {
        handleExit();
      };
    }, []);

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
