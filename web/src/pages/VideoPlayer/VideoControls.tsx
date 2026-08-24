import React, { useState, useRef, useEffect, useCallback } from "react";
import {
  IconButton,
  Slider,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
} from "@mui/material";
import "./VideoControls.css";
import {
  Pause,
  PlayArrow,
  Audiotrack,
  Subtitles,
  Check,
  VolumeUp,
  VolumeDown,
  VolumeOff,
  ArrowBack,
  InfoOutlined,
  Fullscreen,
} from "@mui/icons-material";

function formatTime(seconds: number): string {
  if (isNaN(seconds) || seconds < 0) return "00:00";
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  const mm = m.toString().padStart(2, "0");
  const ss = s.toString().padStart(2, "0");
  if (h > 0) {
    return `${h}:${mm}:${ss}`;
  }
  return `${mm}:${ss}`;
}

interface IVideoControlsProps {
  handlePlayPause: () => void;
  handlePause?: () => void;
  handlePlay?: () => void;
  handleSeek?: (time: number) => void;
  handleSetAudioTrack?: (id: string | number) => void;
  handleSetSubtitleTrack?: (id: string | number) => void;
  handleSetVolume?: (val: number) => void;
  handleToggleMute?: () => void;
  handleClose?: () => void;
  setInfoModalOpen?: (open: boolean) => void;
  paused: boolean;
  currentTime?: number;
  duration?: number;
  volume?: number;
  muted?: boolean;
  tracks?: any[];
  selectedAudio?: string | number;
  selectedSub?: string | number;
}

export default function VideoControls({
  handlePlayPause,
  handleSeek,
  handleSetAudioTrack,
  handleSetSubtitleTrack,
  handleSetVolume,
  handleToggleMute,
  handleClose,
  setInfoModalOpen,
  paused,
  currentTime = 0,
  duration = 0,
  volume = 100,
  muted = false,
  tracks = [],
  selectedAudio = "auto",
  selectedSub = "auto",
}: IVideoControlsProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [seekValue, setSeekValue] = useState(0);
  const [mouseActive, setMouseActive] = useState(true);
  const inactivityTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const controlsVisible = paused || mouseActive;

  const resetInactivityTimer = useCallback(() => {
    setMouseActive(true);
    if (inactivityTimer.current) clearTimeout(inactivityTimer.current);
    inactivityTimer.current = setTimeout(() => {
      setMouseActive(false);
    }, 3000);
  }, []);

  useEffect(() => {
    resetInactivityTimer();
    return () => {
      if (inactivityTimer.current) clearTimeout(inactivityTimer.current);
    };
  }, [resetInactivityTimer]);

  const [audioMenuAnchor, setAudioMenuAnchor] = useState<null | HTMLElement>(
    null,
  );
  const [subMuiMenuAnchor, setSubMuiMenuAnchor] = useState<null | HTMLElement>(
    null,
  );

  const audioTracks = (tracks || []).filter(
    (t) => t.type === "audio" || t.type === "a",
  );
  const subTracks = (tracks || []).filter(
    (t) => t.type === "sub" || t.type === "s",
  );

  const activeAudioTrack =
    audioTracks.find(
      (t) =>
        t.selected ||
        (selectedAudio !== "auto" &&
          t.id.toString() === selectedAudio?.toString()),
    ) ||
    audioTracks.find((t) => t.selected) ||
    audioTracks[0];

  const audioLang =
    activeAudioTrack?.lang?.toUpperCase() ||
    (activeAudioTrack?.title
      ? activeAudioTrack.title.substring(0, 3).toUpperCase()
      : activeAudioTrack
        ? `A${activeAudioTrack.id}`
        : "");

  const activeSubTrack =
    selectedSub !== "no"
      ? subTracks.find(
          (t) =>
            t.selected ||
            (selectedSub !== "auto" &&
              t.id.toString() === selectedSub?.toString()),
        ) || subTracks.find((t) => t.selected)
      : null;

  const subLang =
    selectedSub === "no"
      ? "OFF"
      : activeSubTrack?.lang?.toUpperCase() ||
        (activeSubTrack?.title
          ? activeSubTrack.title.substring(0, 3).toUpperCase()
          : activeSubTrack
            ? `S${activeSubTrack.id}`
            : "AUTO");

  const displayTime = isDragging ? seekValue : currentTime;

  const handleSliderChange = (
    _event: Event | React.SyntheticEvent,
    newValue: number | number[],
  ) => {
    setIsDragging(true);
    setSeekValue(newValue as number);
  };

  const handleSliderChangeCommitted = (
    _event: Event | React.SyntheticEvent,
    newValue: number | number[],
  ) => {
    setIsDragging(false);
    handleSeek?.(newValue as number);
  };

  return (
    <>
      <div
        className="play-pause-overlay"
        onClick={handlePlayPause}
        onMouseMove={resetInactivityTimer}
        onMouseEnter={resetInactivityTimer}
      />
      <div
        className="controls-overlay"
        onMouseMove={resetInactivityTimer}
        onMouseEnter={resetInactivityTimer}
        style={{ cursor: controlsVisible ? "default" : "none" }}
      >
        <div
          className="controls-bottom"
          style={{
            opacity: controlsVisible ? 1 : 0,
            transition: "opacity 0.6s ease",
            pointerEvents: controlsVisible ? "auto" : "none",
          }}
        >
          <IconButton
            onClick={handleClose}
            sx={{
              position: "absolute",
              top: 16,
              left: 16,
              color: "white",
              zIndex: 10,
            }}
          >
            <ArrowBack />
          </IconButton>
          <IconButton
            onClick={() => setInfoModalOpen?.(true)}
            sx={{
              position: "absolute",
              top: 16,
              right: 16,
              color: "white",
              zIndex: 10,
            }}
          >
            <InfoOutlined />
          </IconButton>
          {/* Top Row: Seekbar & Time */}
          <div className="controls-row controls-row-top">
            <span className="controls-time">{formatTime(displayTime)}</span>
            <Slider
              className="controls-seekbar"
              aria-label="Seekbar"
              min={0}
              max={duration || 100}
              step={1}
              value={isDragging ? seekValue : currentTime || 0}
              onChange={handleSliderChange}
              onChangeCommitted={handleSliderChangeCommitted}
              valueLabelDisplay="auto"
              valueLabelFormat={formatTime}
              sx={{
                color: "red",
                "& .MuiSlider-rail, & .MuiSlider-track": {
                  transition: "height 0.2s ease-in-out",
                  height: 2,
                },
                "& .MuiSlider-thumb": {
                  transition:
                    "width 0.2s ease-in-out, height 0.2s ease-in-out, box-shadow 0.2s ease-in-out",
                  width: 12,
                  height: 12,
                },
                "&:hover": {
                  "& .MuiSlider-rail, & .MuiSlider-track": {
                    height: 5,
                  },
                  "& .MuiSlider-thumb": {
                    width: 16,
                    height: 16,
                  },
                },
              }}
            />
            <span className="controls-time">{formatTime(duration)}</span>
          </div>

          {/* Bottom Row: Buttons */}
          <div className="controls-row controls-row-bottom">
            <div className="controls-left">
              <IconButton
                onClick={handlePlayPause}
                className="play-button"
                sx={{ color: "#ffffff" }}
                size="medium"
              >
                {paused ? (
                  <PlayArrow sx={{ fontSize: 28 }} />
                ) : (
                  <Pause sx={{ fontSize: 28 }} />
                )}
              </IconButton>

              <div className="controls-volume-container">
                <IconButton
                  onClick={handleToggleMute}
                  sx={{ color: "#ffffff" }}
                  size="medium"
                  title={muted || volume === 0 ? "Unmute" : "Mute"}
                >
                  {muted || volume === 0 ? (
                    <VolumeOff sx={{ fontSize: 22 }} />
                  ) : volume < 50 ? (
                    <VolumeDown sx={{ fontSize: 22 }} />
                  ) : (
                    <VolumeUp sx={{ fontSize: 22 }} />
                  )}
                </IconButton>
                <Slider
                  className="controls-volume-slider"
                  size="small"
                  aria-label="Volume"
                  min={0}
                  max={100}
                  value={muted ? 0 : volume}
                  onChange={(_e, val) => handleSetVolume?.(val as number)}
                  valueLabelDisplay="auto"
                  valueLabelFormat={(v) => `${v}%`}
                  sx={{
                    color: "white",
                    "& .MuiSlider-rail, & .MuiSlider-track": {
                      transition: "height 0.2s ease-in-out",
                      height: 3,
                    },
                    "& .MuiSlider-thumb": {
                      transition:
                        "width 0.2s ease-in-out, height 0.2s ease-in-out, box-shadow 0.2s ease-in-out",
                      width: 8,
                      height: 8,
                    },
                    "&:hover": {
                      "& .MuiSlider-rail, & .MuiSlider-track": {
                        height: 5,
                      },
                      "& .MuiSlider-thumb": {
                        width: 12,
                        height: 12,
                      },
                    },
                  }}
                />
              </div>
            </div>

            <div className="controls-right">
              {audioTracks.length > 0 && (
                <div className="controls-track-item">
                  <div
                    className="controls-lang-container"
                    onClick={(e) => setAudioMenuAnchor(e.currentTarget)}
                  >
                    <IconButton
                      sx={{ color: "#ffffff" }}
                      size="medium"
                      title="Audio Tracks"
                    >
                      <Audiotrack />
                    </IconButton>
                    {audioLang && (
                      <span className="controls-lang-badge">{audioLang}</span>
                    )}
                  </div>
                  <Menu
                    anchorEl={audioMenuAnchor}
                    open={Boolean(audioMenuAnchor)}
                    onClose={() => setAudioMenuAnchor(null)}
                    PaperProps={{
                      sx: { backgroundColor: "#1e1e1e", color: "#ffffff" },
                    }}
                  >
                    {audioTracks.map((t) => {
                      const isSelected =
                        t.selected ||
                        t.id.toString() === selectedAudio?.toString();
                      const langStr = t.lang
                        ? ` (${t.lang.toUpperCase()})`
                        : "";
                      const label = `${t.title || `Audio Track ${t.id}`}${langStr}`;
                      return (
                        <MenuItem
                          key={t.id}
                          selected={isSelected}
                          onClick={() => {
                            handleSetAudioTrack?.(t.id);
                            setAudioMenuAnchor(null);
                          }}
                        >
                          {isSelected && (
                            <ListItemIcon
                              sx={{ color: "#ffffff", minWidth: 32 }}
                            >
                              <Check fontSize="small" />
                            </ListItemIcon>
                          )}
                          <ListItemText
                            inset={!isSelected}
                            primary={label}
                            secondary={
                              t.codec ? `Codec: ${t.codec}` : undefined
                            }
                            secondaryTypographyProps={{
                              style: { color: "#aaa" },
                            }}
                          />
                        </MenuItem>
                      );
                    })}
                  </Menu>
                </div>
              )}
              {subTracks.length > 0 && (
                <div className="controls-track-item">
                  <div
                    className="controls-lang-container"
                    onClick={(e) => setSubMuiMenuAnchor(e.currentTarget)}
                  >
                    <IconButton
                      sx={{ color: "#ffffff" }}
                      size="medium"
                      title="Subtitles"
                    >
                      <Subtitles />
                    </IconButton>
                    <span className="controls-lang-badge">{subLang}</span>
                  </div>
                  <Menu
                    anchorEl={subMuiMenuAnchor}
                    open={Boolean(subMuiMenuAnchor)}
                    onClose={() => setSubMuiMenuAnchor(null)}
                    PaperProps={{
                      sx: { backgroundColor: "#1e1e1e", color: "#ffffff" },
                    }}
                  >
                    <MenuItem
                      onClick={() => {
                        handleSetSubtitleTrack?.("no");
                        setSubMuiMenuAnchor(null);
                      }}
                    >
                      {selectedSub === "no" && (
                        <ListItemIcon sx={{ color: "#ffffff", minWidth: 32 }}>
                          <Check fontSize="small" />
                        </ListItemIcon>
                      )}
                      <ListItemText
                        inset={selectedSub !== "no"}
                        primary="Off"
                      />
                    </MenuItem>
                    {subTracks.map((t) => {
                      const isSelected =
                        selectedSub !== "no" &&
                        (t.selected ||
                          t.id.toString() === selectedSub?.toString());
                      const langStr = t.lang
                        ? ` (${t.lang.toUpperCase()})`
                        : "";
                      const label = `${t.title || `Subtitle ${t.id}`}${langStr}`;
                      return (
                        <MenuItem
                          key={t.id}
                          selected={isSelected}
                          onClick={() => {
                            handleSetSubtitleTrack?.(t.id);
                            setSubMuiMenuAnchor(null);
                          }}
                        >
                          {isSelected && (
                            <ListItemIcon
                              sx={{ color: "#ffffff", minWidth: 32 }}
                            >
                              <Check fontSize="small" />
                            </ListItemIcon>
                          )}
                          <ListItemText inset={!isSelected} primary={label} />
                        </MenuItem>
                      );
                    })}
                  </Menu>
                </div>
              )}
              <IconButton
                sx={{ color: "#ffffff" }}
                size="medium"
                title="Toggle Fullscreen"
                onClick={() => {
                  window.electron.toggleFullscreen();
                }}
              >
                <Fullscreen />
              </IconButton>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
