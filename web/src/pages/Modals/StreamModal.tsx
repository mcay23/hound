import { Dialog, IconButton } from "@mui/material";
import "./StreamModal.css";
import { ArrowBack } from "@mui/icons-material";
import "video.js/dist/video-js.css";
import VideoPlayer from "../VideoPlayer/VideoPlayer";
import { SERVER_URL } from "./../../config/axios_config";
import { useEffect, useState } from "react";
import axios from "axios";
import toast from "react-hot-toast";

function StreamModal(props: any) {
  const { streamDetails, streams, setOpen, open, startTime } = props;
  const [videoURL, setVideoURL] = useState("");
  const [loading, setLoading] = useState(false);
  const handleClose = () => {
    setLoading(false);
    setOpen(false);
  };

  useEffect(() => {
    if (!open) {
      setVideoURL("");
      return;
    }
    setLoading(true);
    if (streamDetails) {
      if (streamDetails.stream_protocol === "p2p") {
        const fetchToast = toast.loading("Fetching torrent...");
        axios
          .post("/api/v1/torrent/" + streamDetails.encoded_data)
          .then(() => {
            toast.dismiss(fetchToast);
            setVideoURL(
              SERVER_URL + "/api/v1/stream/" + streamDetails.encoded_data
            );
            setLoading(false);
          })
          .catch((err) => {
            toast.error("Failed to add torrent " + err, { id: fetchToast });
          });
      } else {
        setVideoURL(
          SERVER_URL + "/api/v1/stream/" + streamDetails.encoded_data
        );
        setLoading(false);
      }
    }
  }, [streamDetails, streams, open, startTime]);

  const videoJsOptions = {
    autoplay: true,
    muted: false,
    startTime: startTime,
    sources: [
      {
        src: videoURL,
        type: "video/mp4",
      },
    ],
  };

  const handleVideoProgress = (current: number, total: number) => {
    if (current < 300) return; // don't log before 5 minutes
    const payload = {
      stream_protocol: streamDetails.stream_protocol,
      source_uri: streamDetails.uri,
      encoded_data: streamDetails.encoded_data,
      current_progress_seconds: Math.floor(current),
      total_duration_seconds: Math.floor(total),
      ...(streams.media_type === "tvshow"
        ? {
            season_number: streams.season_number || 0,
            episode_number: streams.episode_number || 0,
          }
        : {}),
    };
    axios
      .post(
        `/api/v1/${streams.media_type === "tvshow" ? "tv" : "movie"}/${
          streams.media_source
        }-${streams.source_id}/playback`,
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
      open={open && !loading}
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
      <VideoPlayer
        options={videoJsOptions}
        onVideoProgress={handleVideoProgress}
        setLoading={setLoading}
      />
    </Dialog>
  );
}

export default StreamModal;
