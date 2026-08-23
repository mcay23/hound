import React, { useState } from "react";
import { IconButton, Slider } from "@mui/material";
import "./VideoControls.css";
import { Pause, PlayArrow } from "@mui/icons-material";

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
  paused: boolean;
  currentTime?: number;
  duration?: number;
}

export default function VideoControls({
  handlePlayPause,
  handleSeek,
  paused,
  currentTime = 0,
  duration = 0,
}: IVideoControlsProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [seekValue, setSeekValue] = useState(0);

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
    <div className="controls-overlay">
      <div className="controls-bottom">
        <IconButton
          onClick={handlePlayPause}
          className="play-button"
          sx={{ color: "#ffffff" }}
          size="large"
        >
          {paused ? (
            <PlayArrow sx={{ fontSize: 32 }} />
          ) : (
            <Pause sx={{ fontSize: 32 }} />
          )}
        </IconButton>
        <span className="controls-time">
          {formatTime(displayTime)} / {formatTime(duration)}
        </span>
        <Slider
          className="controls-seekbar"
          size="small"
          aria-label="Seekbar"
          min={0}
          max={duration || 100}
          step={1}
          value={isDragging ? seekValue : currentTime || 0}
          onChange={handleSliderChange}
          onChangeCommitted={handleSliderChangeCommitted}
          valueLabelDisplay="auto"
          valueLabelFormat={formatTime}
        />
      </div>
    </div>
  );
}
