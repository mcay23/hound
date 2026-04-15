import { Dialog } from "@mui/material";
import "./VideoModal.css";

function VideoModal(props: any) {
  const { onClose, open, videoKey } = props;
  const handleClose = () => {
    onClose();
  };
  return (
    <Dialog
      onClose={handleClose}
      open={open}
      disableScrollLock={false}
      className="video-dialog"
      maxWidth={false}
    >
      <iframe
        width="960px"
        height="540px"
        src={`https://www.youtube.com/embed/${videoKey}?autoplay=1`}
        title={"title"}
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        allowFullScreen
      />
    </Dialog>
  );
}

export default VideoModal;
