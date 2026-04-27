import {
  Chip,
  Dialog,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  SelectChangeEvent,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import "./StreamSelectModal.css";
import "video.js/dist/video-js.css";
import { slotPropsGlass, paperPropsGlass } from "./modalStyles";
import { Button, Spinner } from "react-bootstrap";
import axios from "axios";
import toast from "react-hot-toast";
import { useEffect, useState } from "react";
import {
  useProvidersMutation,
  useUnifiedStreamsMutation,
} from "../../api/hooks/providers";
import { useProviderProfiles } from "../../api/hooks/providerProfiles";
import { copyToClipboard } from "../../helpers/helpers";

type FetchParams = {
  mediaType: string;
  mediaSource: string;
  sourceId: string;
  season?: number;
  episode?: number;
};

function SelectStreamModal(props: {
  modalType: "select-stream" | "download-season";
  fetchParams?: FetchParams;
  setOpen: (open: boolean) => void;
  open: boolean;
  setMainStream: (stream: any) => void;
  setProviderID?: (providerID: number) => void;
  setIsStreamModalOpen?: (open: boolean) => void;
}) {
  const {
    modalType,
    fetchParams,
    setOpen,
    open,
    setMainStream,
    setProviderID: setProviderIDSeasonDownloader,
    setIsStreamModalOpen,
  } = props;

  const [streamData, setStreamData] = useState<any[] | null>(null);
  const [providerID, setProviderID] = useState<number | undefined>(undefined);
  const { data: providerProfiles } = useProviderProfiles();
  const { mutateAsync: fetchUnifiedStreams } = useUnifiedStreamsMutation();
  const { mutateAsync: fetchProviders } = useProvidersMutation();

  useEffect(() => {
    if (
      providerProfiles &&
      providerProfiles.length > 0 &&
      providerID === undefined
    ) {
      setProviderID(providerProfiles[0].provider_profile_id);
    }
  }, [providerProfiles]);

  const providerProfileId = providerID;

  useEffect(() => {
    if (!open) return;
    if (
      providerProfiles &&
      providerProfiles.length > 0 &&
      providerID === undefined
    ) {
      return;
    }
    setStreamData(null);
    if (fetchParams && modalType === "select-stream") {
      fetchUnifiedStreams({
        ...fetchParams,
        providerProfileId: providerProfileId,
      })
        .then((data) => {
          setStreamData(data?.streams ?? []);
        })
        .catch((err) => {
          console.error("Failed to fetch streams", err);
          toast.error("Failed to fetch streams");
        });
    } else if (fetchParams && modalType === "download-season") {
      fetchProviders({
        ...fetchParams,
        providerProfileId: providerProfileId,
      })
        .then((data) => {
          const allStreams =
            data?.providers?.flatMap((p: any) => p.streams || []) ?? [];
          setStreamData(allStreams);
        })
        .catch((err) => {
          console.error("Failed to fetch providers", err);
          toast.error("Failed to fetch providers");
        });
    }
  }, [open, providerID]);

  const handleClose = () => {
    setOpen(false);
  };

  const handleProviderChange = (event: SelectChangeEvent<number>) => {
    setProviderID(Number(event.target.value));
    setProviderIDSeasonDownloader?.(Number(event.target.value));
  };

  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm")); // sm = 600px by default

  return (
    <>
      {open && (
        <Dialog
          onClose={handleClose}
          open={open}
          disableScrollLock={false}
          fullScreen={fullScreen}
          fullWidth
          maxWidth="md"
          className="stream-select-modal-dialog"
          slotProps={slotPropsGlass}
          PaperProps={paperPropsGlass}
        >
          <div className="stream-info-card-container">
            {modalType === "download-season" ? (
              <div className="px-4 mb-3 h5">
                Choose a season pack to prioritize...
              </div>
            ) : (
              ""
            )}
            {providerProfiles && providerProfiles.length > 0 && (
              <div className="mb-2 px-4">
                <FormControl
                  sx={{
                    mt: 1,
                    mb: 1,
                    minWidth: 120,
                  }}
                  size="small"
                >
                  <InputLabel id="provider-select-label">Provider</InputLabel>
                  <Select
                    labelId="provider-select-label"
                    value={providerID ?? ""}
                    label="Provider"
                    onChange={handleProviderChange}
                  >
                    {providerProfiles.map((provider: any) => (
                      <MenuItem
                        key={provider.provider_profile_id}
                        value={provider.provider_profile_id}
                      >
                        {provider.name}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </div>
            )}
            {streamData === null ? (
              <div className="d-flex justify-content-center mt-5 mb-5">
                <Spinner
                  animation="border"
                  size="sm"
                  role="status"
                  id="stream-select-button-loading"
                >
                  <span className="visually-hidden">Loading...</span>
                </Spinner>
              </div>
            ) : streamData.length === 0 ? (
              <div className="d-flex justify-content-center px-4 w-full mt-5 mb-5">
                No streams found.
              </div>
            ) : (
              streamData.map((stream: any) => {
                return (
                  <div className="stream-info-card" key={stream.infohash}>
                    <div
                      className="stream-info-card-title"
                      onClick={() => {
                        if (stream) {
                          // for season pack downloader, sets this stream as the one
                          // to reference the infohash
                          if (modalType === "select-stream") {
                            setMainStream(stream);
                            setIsStreamModalOpen?.(true);
                          } else if (modalType === "download-season") {
                            if (!stream.info_hash || stream.info_hash === "") {
                              toast.error(
                                "This season pack doesn't have a valid info hash, please select another pack",
                              );
                              return;
                            }
                            setMainStream(stream);
                            setOpen(false);
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
                              toast.error("Can't do this in the demo");
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
                                await copyToClipboard(
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
              })
            )}
          </div>
        </Dialog>
      )}
    </>
  );
}

export default SelectStreamModal;
