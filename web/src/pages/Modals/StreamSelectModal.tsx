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
          maxWidth="md"
          fullScreen={fullScreen}
          className="stream-select-modal-dialog"
          slotProps={slotPropsGlass}
          PaperProps={paperPropsGlass}
        >
          <div className="stream-info-card-container">
            {props.streamData?.data?.providers?.[0]?.streams?.map(
              (stream: any) => {
                return (
                  <div className="stream-info-card" key={stream.infohash}>
                    <div className="stream-info-top">
                      {stream.cached === "true" ? (
                        <Chip
                          size="small"
                          className="stream-info-chip"
                          id="stream-info-cache-good"
                          label={"Instant " + stream.service}
                        />
                      ) : (
                        <Chip
                          size="small"
                          className="stream-info-chip"
                          label={"P2P (Slow)"}
                        />
                      )}
                      {stream.resolution ? (
                        <Chip
                          size="small"
                          className="stream-info-chip"
                          id="stream-info-resolution"
                          label={
                            stream.resolution === "2160p"
                              ? "4K"
                              : stream.resolution
                          }
                        />
                      ) : (
                        ""
                      )}
                      {stream.file_size ? (
                        <Chip
                          size="small"
                          className="stream-info-chip"
                          id="stream-info-size"
                          label={formatBytes(stream.file_size)}
                        />
                      ) : (
                        ""
                      )}
                      {stream.data.codec ? (
                        <Chip
                          size="small"
                          className="stream-info-chip"
                          id={
                            stream.data.codec === "hevc"
                              ? "stream-info-codec-hevc"
                              : stream.data.codec === "avc"
                              ? "stream-info-codec-avc"
                              : "stream-info-codec-generic"
                          }
                          label={stream.data.codec
                            .toUpperCase()
                            .replace("HEVC", "x265")
                            .replace("AVC", "x264")}
                        />
                      ) : (
                        ""
                      )}
                      {stream.data.audio.length > 0
                        ? stream.data.audio.map(
                            (item: string, index: number) => {
                              if (item === "Dolby Digital") {
                                item = "DD";
                              } else if (item === "Dolby Digital Plus") {
                                item = "DD+";
                              }
                              return (
                                <Chip
                                  key={index + "-" + item}
                                  size="small"
                                  className="stream-info-chip"
                                  id="stream-info-audio"
                                  label={item}
                                />
                              );
                            }
                          )
                        : ""}
                    </div>
                    <div
                      className="stream-info-card-title"
                      onClick={() => {
                        if (stream) {
                          setMainStream(stream);
                          setIsStreamModalOpen(true);
                        }
                      }}
                    >
                      {stream.file_name
                        ? stream.file_name
                        : "Unknown File Name"}
                    </div>
                    <div className="stream-info-card-subtitle">
                      {stream.addon}
                      {stream.addon && stream.folder_name ? " ⸱ " : ""}
                      {stream.folder_name}
                    </div>
                    <div>
                      <Button
                        onClick={() => {
                          axios
                            .post(
                              "/api/v1/torrent/" +
                                stream.encoded_data +
                                "/download"
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
                  </div>
                );
              }
            )}
          </div>
        </Dialog>
      ) : (
        ""
      )}
    </>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

export default SelectStreamModal;
