import { Dialog, Button, Chip } from "@mui/material";
import "./DownloadSeasonModal.css";
import toast from "react-hot-toast";
import { useDownloadSeason } from "../../api/hooks/media";
import { SeasonDownloadPreferences } from "../../api/services/media";
import { useEffect, useState } from "react";
import {
  useProvidersMutation,
  useUnifiedStreams,
} from "../../api/hooks/providers";
import { Spinner } from "react-bootstrap";
import SelectStreamModal from "./StreamSelectModal";

function DownloadSeasonModal(props: any) {
  const { onClose, open, mediaSource, sourceID, seasonData } = props;
  const searchProviders = useProvidersMutation();
  const [streams, setStreams] = useState<any[]>([]);
  const [mainStream, setMainStream] = useState<any>(null);
  const [preferences, setPreferences] = useState<
    SeasonDownloadPreferences | undefined
  >(undefined);
  const [isSelectStreamModalOpen, setIsSelectStreamModalOpen] = useState(false);
  const downloadSeasonMutation = useDownloadSeason();

  // get the first providers, to help users select season packs
  useEffect(() => {
    if (!open) {
      setIsSelectStreamModalOpen(false);
      return;
    }
    if (seasonData.season_number >= 0) {
      if (seasonData.episodes.length <= 0) {
        toast.error("No episodes found for this season!");
        return;
      }
      searchProviders.mutate(
        {
          mediaType: "tv",
          mediaSource: mediaSource,
          sourceId: sourceID,
          season: seasonData.season_number,
          episode: seasonData.episodes[0].episode_number,
        },
        {
          onSuccess: (data) => {
            if (data?.providers?.length > 0) {
              const allStreams = data.providers.flatMap(
                (p: any) => p.streams || [],
              );
              setStreams(allStreams || []);
            }
          },
        },
      );
    }
  }, [open, mediaSource, sourceID, seasonData.season_number]);

  const handleConfirm = () => {
    toast.promise(
      downloadSeasonMutation.mutateAsync({
        mediaType: "tv",
        mediaSource: mediaSource,
        sourceID: sourceID,
        seasonNum: seasonData.season_number,
        preferences: preferences,
      }),
      {
        loading: "Queueing download...",
        success: "Downloads queued",
        error: (err) => `Error downloading season: ${err}`,
      },
    );
    onClose();
  };

  return (
    <Dialog onClose={onClose} open={open} disableScrollLock={false}>
      <div className="download-season-modal-container">
        <h4>Download Season Wizard (EXPERIMENTAL)</h4>
        <p>
          This wizard is experimental, downloading episodes manually is still
          the official method.
        </p>
        <div className="download-season-main-content">
          {searchProviders.isPending ? (
            <div className="d-flex justify-content-center">
              <Spinner />
            </div>
          ) : searchProviders.isError ? (
            <p>Error fetching streams.</p>
          ) : (
            <div>
              <Button onClick={() => setIsSelectStreamModalOpen(true)} />
            </div>
          )}
        </div>
        <div className="d-flex justify-content-end">
          <Button onClick={onClose}>Cancel</Button>
          <Button onClick={handleConfirm}>Confirm</Button>
        </div>
      </div>
      <SelectStreamModal
        modalType="select-stream"
        open={open && isSelectStreamModalOpen}
        setOpen={setIsSelectStreamModalOpen}
        setMainStream={setMainStream}
        streamData={streams}
      />
    </Dialog>
  );
}

function SelectedSeasonPack(mainStream: any) {
  return (
    <div className="stream-info-card" key={mainStream.infohash}>
      <div className="stream-info-card-title">{mainStream.title}</div>
      <div className="stream-info-card-subtitle">{mainStream.description}</div>
      <div className="stream-info-card-subtitle mb-2">
        info hash: {mainStream.info_hash}
      </div>
      <Chip label={mainStream.provider} size="small" />
    </div>
  );
}

export default DownloadSeasonModal;
