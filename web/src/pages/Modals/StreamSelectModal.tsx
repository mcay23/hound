import { Chip, Dialog, useMediaQuery, useTheme } from "@mui/material";
import "./StreamSelectModal.css";
import "video.js/dist/video-js.css";

function SelectStreamModal(props: any) {
  const { setOpen, open } = props;
  const handleClose = () => {
    setOpen(false);
  };
  // Only aiostreams for now
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm")); // sm = 600px by default
  console.log(props.streamData?.data?.providers?.[0]?.streams[1]);
  return (
    <Dialog
      onClose={handleClose}
      open={open}
      disableScrollLock={false}
      maxWidth="md"
      fullScreen={fullScreen}
      className="stream-select-modal-dialog"
    >
      {props.streamData !== null ? (
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
                        label={
                          stream.resolution === "2160p"
                            ? "4K"
                            : stream.resolution
                        }
                      />
                    ) : (
                      ""
                    )}

                    <Chip
                      size="small"
                      className="stream-info-chip"
                      id="stream-info-size"
                      label={formatBytes(stream.file_size)}
                    />
                  </div>
                  <div className="stream-info-card-title">
                    {stream.file_name}
                  </div>
                  <div className="stream-info-card-subtitle">
                    {stream.addon}
                    {stream.addon && stream.folder_name ? " ⸱ " : ""}
                    {stream.folder_name}
                  </div>
                </div>
              );
            }
          )}
        </div>
      ) : (
        <></>
      )}
    </Dialog>
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
