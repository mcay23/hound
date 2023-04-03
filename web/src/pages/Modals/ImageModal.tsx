import { Dialog } from "@mui/material";
import "./ImageModal.css";

function ImageModal(props: any) {
  const { onClose, open, imageURL } = props;
  const handleClose = () => {
    onClose();
  };
  return (
    <Dialog
      onClose={handleClose}
      open={open}
      disableScrollLock={true}
      className="video-dialog"
      maxWidth={false}
    >
      <img src={imageURL} alt={"test"} />
      {/* <iframe
        width="960px"
        height="540px"
        src={`https://www.youtube.com/embed/${videoKey}?autoplay=1`}
        title={"title"}
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        allowFullScreen
      /> */}
    </Dialog>
  );
}

export default ImageModal;
