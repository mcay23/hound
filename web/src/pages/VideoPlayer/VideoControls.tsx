import { IconButton, Slider } from "@mui/material";
import "./VideoControls.css";
import { Pause, PlayArrow } from "@mui/icons-material";

export default function VideoControls(props: any) {
  return (
    <div className="controls-overlay">
      <div className="controls-bottom">
        <IconButton
          onClick={() => props?.handlePlayPause()}
          className="play-button"
          sx={{ color: "#ffffff" }}
          size="large"
        >
          {props.paused ? (
            <PlayArrow sx={{ fontSize: 32 }} />
          ) : (
            <Pause sx={{ fontSize: 32 }} />
          )}
        </IconButton>
        <Slider
          className="controls-seekbar"
          size="small"
          aria-label="Seekbar"
          valueLabelDisplay="auto"
        />
      </div>
    </div>
  );
}
