import { Chip, Dialog, Fade, useMediaQuery, useTheme } from "@mui/material";
import "./StreamSelectModal.css";
import "video.js/dist/video-js.css";
import { slotPropsGlass, paperPropsGlass } from "./modalStyles";
import { Button } from "react-bootstrap";
import axios from "axios";
import toast from "react-hot-toast";

function SelectStreamModal(props: any) {
  const { setOpen, open, setMainStream, setIsStreamModalOpen } = props;
  const handleClose = () => {
    setOpen(false);
  };
  // Only aiostreams for now
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm")); // sm = 600px by default
  return (
    <>
      {props.streamData !== null ? (
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
            {props.streamData?.providers?.map((provider: any) =>
              provider?.streams?.map((stream: any) => {
                return (
                  <div className="stream-info-card" key={stream.infohash}>
                    <div
                      className="stream-info-card-title"
                      onClick={() => {
                        if (stream) {
                          setMainStream(stream);
                          setIsStreamModalOpen(true);
                        }
                      }}
                    >
                      {stream.title}
                    </div>
                    <div className="stream-info-card-subtitle">
                      {stream.description}
                    </div>
                    <Chip label={provider.provider} size="small" />
                    {provider.provider !== "Hound" ? (
                      <div className="stream-info-card-footer mt-2">
                        <Button
                          className="stream-info-card-footer-buttons"
                          variant="light"
                          size="sm"
                          onClick={() => {
                            axios
                              .post(
                                "/api/v1/stream/" +
                                  stream.encoded_data +
                                  "/download",
                              )
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
                      </div>
                    ) : (
                      <></>
                    )}
                  </div>
                );
              }),
            )}
          </div>
        </Dialog>
      ) : (
        ""
      )}
    </>
  );
}

export default SelectStreamModal;
