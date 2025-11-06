import { Dialog, IconButton } from "@mui/material";
import "./StreamModal.css";
import { ArrowBack } from "@mui/icons-material";
import "video.js/dist/video-js.css";
import VideoPlayer from "../VideoPlayer/VideoPlayer";
import houndConfig from "./../../config.json";

function StreamModal(props: any) {
  const { setOpen, open } = props;
  const handleClose = () => {
    setOpen(false);
  };
  let videoURL = "";
  if (props.streamDetails != null) {
    videoURL =
      houndConfig.server_host +
      "/api/v1/stream/" +
      props.streamDetails.encoded_data;
  }
  // videoURL =
  //   "https://filesamples.com/samples/video/mkv/sample_1280x720_surfing_with_audio.mkv";
  const videoJsOptions = {
    sources: [
      {
        src: videoURL,
        type: "video/mp4",
      },
    ],
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
      <VideoPlayer options={videoJsOptions} />
    </Dialog>
  );
}

export default StreamModal;
