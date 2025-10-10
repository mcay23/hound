import { Dialog, IconButton } from "@mui/material";
import "./StreamModal.css";
import { CloseOutlined } from "@mui/icons-material";
import "video.js/dist/video-js.css";
import VideoPlayer from "../VideoPlayer/VideoPlayer";
import { Col, Container, Row } from "react-bootstrap";

function StreamModal(props: any) {
  const { setOpen, open } = props;
  const handleClose = () => {
    setOpen(false);
  };
  if (props.streamDetails === null) {
    return <>Loading...</>;
  }
  props.streamDetails.url =
    "https://filesamples.com/samples/video/mkv/sample_1280x720_surfing_with_audio.mkv";
  const videoJsOptions = {
    sources: [
      {
        src: props.streamDetails.url,
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
    >
      <Container>
        <Row>
          <Col sm={2}>
            <IconButton
              onClick={handleClose}
              sx={{
                position: "absolute",
                top: 16,
                left: 16,
                color: "black",
                zIndex: 10,
              }}
            >
              <CloseOutlined />
            </IconButton>
            <div className="videojs-container">
              <VideoPlayer options={videoJsOptions} />
            </div>
          </Col>
          <Col>test</Col>
        </Row>
        <Row>abc</Row>
      </Container>
    </Dialog>
  );
}

export default StreamModal;
