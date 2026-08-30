import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
} from "@mui/material";
import "./StreamModal.css";
import { ArrowBack, InfoOutlined, Pause } from "@mui/icons-material";
import "video.js/dist/video-js.css";
import { getBaseUrl } from "./../../config/axios_config";
import { useEffect, useState, useMemo, useCallback } from "react";
import axios from "axios";
import toast from "react-hot-toast";
import { useDecodeStream, useSubtitles } from "../../api/hooks/providers";
import MPVElectronPlayer from "../VideoPlayer/MPVElectronPlayer";
import VideoPlayer from "../VideoPlayer/VideoPlayer";
import { isPlatformElectron } from "../../utils/platform";
import { get2LetterLangCode } from "../../helpers/locale";

function StreamModal(props: any) {
  const {
    streamDetails,
    streams,
    setOpen,
    open,
    startTime,
    watchProgress,
    originalAudioLang,
  } = props;
  const [videoURL, setVideoURL] = useState("");
  const [loading, setLoading] = useState(false);
  const [infoModalOpen, setInfoModalOpen] = useState(false);

  const isStreamsMatch = useMemo(
    () =>
      Boolean(
        watchProgress &&
          watchProgress.encoded_data &&
          watchProgress.encoded_data === streamDetails?.encoded_data,
      ),
    [watchProgress, streamDetails?.encoded_data],
  );

  const { data: subtitleData } = useSubtitles(
    streams?.media_type === "tvshow" ? "tv" : "movie",
    streams?.media_source,
    streams?.source_id,
    streams?.season_number,
    streams?.episode_number,
    open && !!streams,
  );
  const { data: decodedData } = useDecodeStream(streamDetails?.encoded_data);
  const subtitles = useMemo(
    () => subtitleData?.subtitles?.flatMap((p: any) => p.subtitles || []) || [],
    [subtitleData],
  );
  const externalSubtitles = useMemo(() => {
    return subtitles.map((sub: any) => ({
      title: sub.title,
      lang: get2LetterLangCode(sub.lang),
      url: `${getBaseUrl()}/api/v1/subtitle/${sub.encoded_data}`,
    }));
  }, [subtitles]);
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
              getBaseUrl() + "/api/v1/stream/" + streamDetails.encoded_data,
            );
            setLoading(false);
          })
          .catch((err) => {
            toast.error("Failed to add torrent " + err, { id: fetchToast });
          });
      } else {
        setVideoURL(
          getBaseUrl() + "/api/v1/stream/" + streamDetails.encoded_data,
        );
        setLoading(false);
      }
    }
  }, [streamDetails, streams, open, startTime]);

  const videoJsOptions = useMemo(
    () => ({
      autoplay: true,
      muted: false,
      startTime: startTime,
      sources: [
        {
          src: videoURL,
          type: "video/mp4",
        },
      ],
    }),
    [videoURL, startTime],
  );

  const handleVideoProgress = useCallback(
    (current: number, total: number, playerSettings?: any) => {
      if (current < 120) return; // don't log before 2 minutes
      const payload: any = {
        stream_protocol: streamDetails?.stream_protocol,
        source_uri: streamDetails?.uri,
        encoded_data: streamDetails?.encoded_data,
        current_progress_seconds: Math.floor(current),
        total_duration_seconds: Math.floor(total),
        ...(streams?.media_type === "tvshow"
          ? {
              season_number: streams?.season_number || 0,
              episode_number: streams?.episode_number || 0,
            }
          : {}),
      };
      if (isPlatformElectron && playerSettings) {
        payload.player_settings = playerSettings;
      }
      axios
        .post(
          `/api/v1/${streams?.media_type === "tvshow" ? "tv" : "movie"}/${
            streams?.media_source
          }-${streams?.source_id}/playback`,
          payload,
        )
        .then((res) => {
          // console.log(res.data);
        })
        .catch((err) => {
          console.log(err);
        });
    },
    [streamDetails, streams],
  );
  return (
    <Dialog
      onClose={handleClose}
      open={open && !loading}
      disableScrollLock={false}
      fullScreen
      disableEscapeKeyDown
      PaperProps={{
        sx: {
          margin: 0,
          backgroundColor: "black",
          maxHeight: "100vh",
          width: "100vw",
        },
      }}
    >
      {isPlatformElectron ? (
        <MPVElectronPlayer
          options={videoJsOptions}
          onVideoProgress={handleVideoProgress}
          setLoading={setLoading}
          handleClose={handleClose}
          setInfoModalOpen={setInfoModalOpen}
          externalSubtitles={externalSubtitles}
          playerSettings={watchProgress?.player_settings}
          isStreamsMatch={isStreamsMatch}
          originalAudioLang={originalAudioLang}
        />
      ) : (
        <VideoPlayer
          options={videoJsOptions}
          onVideoProgress={handleVideoProgress}
          setLoading={setLoading}
          subtitles={subtitles}
        />
      )}
      <InfoModal
        open={infoModalOpen}
        setOpen={setInfoModalOpen}
        decodedData={decodedData}
      />
    </Dialog>
  );
}

function InfoModal({
  open,
  setOpen,
  decodedData,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
  decodedData: any;
}) {
  const handleClose = () => {
    setOpen(false);
  };
  return (
    <Dialog onClose={handleClose} open={open}>
      <DialogTitle>Stream Info</DialogTitle>
      <DialogContent>
        <hr />
        <DialogContentText>
          <h4> {decodedData?.title}</h4>
          {decodedData?.description}
          <br />
          <hr />
          {decodedData?.provider_profile_name &&
            "Provider Profile: " + decodedData?.provider_profile_name}
          <br />
          Protocol: {decodedData?.stream_protocol}
        </DialogContentText>
      </DialogContent>
    </Dialog>
  );
}

export default StreamModal;
