import { Chip, Dialog, Fade, useMediaQuery, useTheme } from "@mui/material";
import "./StreamSelectModal.css";
import "video.js/dist/video-js.css";
import { slotPropsGlass, paperPropsGlass } from "./modalStyles";
import { Button } from "react-bootstrap";
import axios from "axios";
import toast from "react-hot-toast";

function SelectStreamModal(props: {
  modalType: "select-stream" | "download-season";
  streamData: any;
  setOpen: (open: boolean) => void;
  open: boolean;
  setMainStream: (stream: any) => void;
  setIsStreamModalOpen?: (open: boolean) => void;
}) {
  const {
    modalType,
    streamData,
    setOpen,
    open,
    setMainStream,
    setIsStreamModalOpen,
  } = props;
  const handleClose = () => {
    setOpen(false);
  };
  // Only aiostreams for now
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm")); // sm = 600px by default
  return (
    <>
      {streamData !== null ? (
        <Dialog
          onClose={handleClose}
          open={open}
          disableScrollLock={false}
          fullScreen={fullScreen}
          className="stream-select-modal-dialog"
          slotProps={slotPropsGlass}
          PaperProps={paperPropsGlass}
        >
          <div className="stream-info-card-container">
            {streamData?.streams?.map((stream: any) => {
              return (
                <div className="stream-info-card" key={stream.infohash}>
                  <div
                    className="stream-info-card-title"
                    onClick={() => {
                      if (stream) {
                        // for season pack downloader, sets this stream as the one
                        // to reference the infohash
                        setMainStream(stream);
                        if (modalType === "select-stream") {
                          setIsStreamModalOpen?.(true);
                        }
                      }
                    }}
                  >
                    {stream.title}
                  </div>
                  <div className="stream-info-card-subtitle">
                    {stream.description}
                  </div>
                  <div className="stream-info-card-subtitle mb-2">
                    info hash: {stream.info_hash}
                  </div>
                  <Chip label={stream.provider} size="small" />
                  {modalType === "select-stream" ? (
                    <div className="stream-info-card-footer mt-2">
                      {stream.provider !== "Hound" && (
                        <Button
                          className="stream-info-card-footer-buttons me-2"
                          variant="light"
                          size="sm"
                          onClick={() => {
                            axios
                              .post("/api/v1/download/" + stream.encoded_data)
                              .then((res) => {
                                toast.success("Download added to queue");
                              })
                              .catch((err) => {
                                toast.error("Download Failed! " + err);
                              });
                          }}
                        >
                          Download to Hound
                        </Button>
                      )}
                      <Button
                        className="stream-info-card-footer-buttons"
                        variant="light"
                        size="sm"
                        onClick={() => {
                          const handleCopy = async () => {
                            try {
                              await navigator.clipboard.writeText(
                                window.location.origin +
                                  "/api/v1/stream/" +
                                  stream.encoded_data,
                              );
                              toast.success("Link copied to clipboard");
                            } catch (err) {
                              console.error("Failed to copy text: ", err);
                              toast.error("Copy to clipboard failed! " + err);
                            }
                          };
                          handleCopy();
                        }}
                      >
                        Copy Link
                      </Button>
                    </div>
                  ) : (
                    ""
                  )}
                </div>
              );
            })}
          </div>
        </Dialog>
      ) : (
        ""
      )}
    </>
  );
}

export default SelectStreamModal;
