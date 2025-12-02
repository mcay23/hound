import { Dialog, IconButton } from "@mui/material";
import "./StreamModal.css";
import { ArrowBack } from "@mui/icons-material";
import "video.js/dist/video-js.css";
import VideoPlayer from "../VideoPlayer/VideoPlayer";
import houndConfig from "./../../config.json";
import { useEffect, useState } from "react";
import axios from "axios";

function StreamModal(props: any) {
  const { streamDetails, streams, setOpen, open } = props;
  const [videoURL, setVideoURL] = useState("");
  const handleClose = () => {
    setOpen(false);
  };

  useEffect(() => {
    if (streamDetails != null) {
      setVideoURL(
        houndConfig.server_host + "/api/v1/stream/" + streamDetails.encoded_data
      );
      setVideoURL(
        "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4"
      );
    }
  }, [streamDetails, streams, open]);

  const videoJsOptions = {
    sources: [
      {
        src: videoURL,
        type: "video/mp4",
      },
    ],
  };
  const handleSetWatched = () => {
    const payload = {
      action_type: "scrobble",
      ...(streams.media_type === "tvshow"
        ? { episode_ids: [streams.source_episode_id] }
        : {}),
    };
    axios
      .post(
        `/api/v1/${streams.media_type === "tvshow" ? "tv" : "movie"}/${
          streams.media_source
        }-${streams.source_id}/history`,
        payload
      )
      .then((res) => {
        // console.log(res.data);
      })
      .catch((err) => {
        console.log(err);
      });
  };
  return (
    <Dialog
      onClose={handleClose}
      open={open}
      disableScrollLock={false}
      fullScreen
      PaperProps={{
        sx: {
          margin: 0,
          backgroundColor: "black",
          maxHeight: "100vh",
          width: "100vw",
        },
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
      <VideoPlayer options={videoJsOptions} onVideoEnding={handleSetWatched} />
    </Dialog>
  );
}

export default StreamModal;
