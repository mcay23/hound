import {
  Dialog,
  Button,
  Chip,
  Divider,
  TextField,
  FormControl,
  InputLabel,
  Select,
  SelectChangeEvent,
  OutlinedInput,
  MenuItem,
  ListItemText,
  Checkbox,
  FormControlLabel,
} from "@mui/material";

import CheckBoxOutlineBlankIcon from "@mui/icons-material/CheckBoxOutlineBlank";
import CheckBoxIcon from "@mui/icons-material/CheckBox";
import "./DownloadSeasonModal.css";
import toast from "react-hot-toast";
import { useDownloadSeason } from "../../api/hooks/media";
import {
  DownloadPreference,
  MatchTypeInfoHash,
  MatchTypeString,
  SeasonDownloadPreferences,
} from "../../api/services/media";
import { useEffect, useMemo, useState } from "react";
import { useProvidersMutation } from "../../api/hooks/providers";
import { Spinner } from "react-bootstrap";
import SelectStreamModal from "./StreamSelectModal";

function DownloadSeasonModal(props: any) {
  const { onClose, open, mediaSource, sourceID, seasonData } = props;
  const searchProviders = useProvidersMutation();
  const [streams, setStreams] = useState<any[]>([]);
  const [mainStream, setMainStream] = useState<any>(undefined);
  // form states
  const [strictMatch, setStrictMatch] = useState<boolean>(false);
  const [skipDownloaded, setSkipDownloaded] = useState<boolean>(true);
  const [preferredStringMatch, setPreferredStringMatch] = useState<
    string | undefined
  >(undefined);
  const [caseSensitive, setCaseSensitive] = useState<boolean>(false);
  const [episodesToDownload, setEpisodesToDownload] = useState<number[]>([]);

  const [isSelectStreamModalOpen, setIsSelectStreamModalOpen] = useState(false);
  const downloadSeasonMutation = useDownloadSeason();

  const preferenceList: DownloadPreference[] = useMemo(() => {
    const prefs: DownloadPreference[] = [];
    if (mainStream && mainStream.info_hash) {
      prefs.push({
        match_type: MatchTypeInfoHash,
        info_hash_preference: {
          info_hash: mainStream.info_hash,
        },
      });
    }
    if (preferredStringMatch) {
      prefs.push({
        match_type: MatchTypeString,
        string_match_preference: {
          match_string: preferredStringMatch,
          case_sensitive: caseSensitive,
        },
      });
    }
    return prefs;
  }, [mainStream, preferredStringMatch, caseSensitive]);

  const preferences: SeasonDownloadPreferences = useMemo(
    () => ({
      strict_match: strictMatch,
      skip_downloaded_episodes: skipDownloaded,
      preference_list: preferenceList,
      episodes_to_download: episodesToDownload,
    }),
    [strictMatch, skipDownloaded, preferenceList],
  );

  // get the first providers, to help users select season packs
  useEffect(() => {
    if (!open) {
      setIsSelectStreamModalOpen(false);
      return;
    }
    // reset default states
    setMainStream(undefined);
    setPreferredStringMatch(undefined);
    setCaseSensitive(false);
    setSkipDownloaded(true);
    setStrictMatch(false);
    if (seasonData.season_number >= 0) {
      if (seasonData.episodes.length <= 0) {
        toast.error("No episodes found for this season!");
        return;
      }
      const episodesList = seasonData.episodes.map(
        (ep: any) => ep.episode_number,
      );
      setEpisodesToDownload(episodesList);
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
    console.log(preferences);
    return;
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
        <p className="pe-5">
          This wizard is experimental, downloading episodes manually is still
          the official method. If you don't set a season pack or string match,
          Hound will download the top result.
        </p>
        <Divider className="mt-3 mb-3" sx={{ borderColor: "black" }} />
        <div className="download-season-main-content">
          {searchProviders.isPending ? (
            <div className="d-flex justify-content-center">
              <Spinner />
            </div>
          ) : searchProviders.isError ? (
            <p>Error fetching streams.</p>
          ) : (
            <div>
              {seasonData ? (
                <>
                  <div className="mb-1">1. Select Episodes</div>
                  <EpisodeSelector
                    episodesData={seasonData?.episodes}
                    episodesToDownload={episodesToDownload}
                    setEpisodesToDownload={setEpisodesToDownload}
                  />
                </>
              ) : (
                ""
              )}
              <Divider className="mt-3 mb-3" sx={{ borderColor: "black" }} />
              {mainStream ? (
                <>
                  <SelectedSeasonPack
                    mainStream={mainStream}
                    setMainStream={setMainStream}
                    setIsSelectStreamModalOpen={setIsSelectStreamModalOpen}
                  />
                </>
              ) : (
                <div>
                  <div className="mb-2">
                    2. (Optional) Select a season pack to match against:{" "}
                  </div>
                  <Button
                    onClick={() => setIsSelectStreamModalOpen(true)}
                    variant="outlined"
                  >
                    Select Season Pack
                  </Button>
                </div>
              )}
              <div className="px-1 mt-2 mb-2 text-muted">
                Episodes in this season pack will be prioritized. Make sure this
                is a season pack and not an individual episode!
              </div>
              <Divider className="mt-3 mb-3" sx={{ borderColor: "black" }} />
              <div>
                <div className="mb-1">
                  3. (Optional) Add a string to match against:
                </div>
                <div>
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={caseSensitive}
                        onChange={(
                          event: React.ChangeEvent<HTMLInputElement>,
                        ) => {
                          setCaseSensitive(event.target.checked);
                        }}
                      />
                    }
                    label="Case Sensitive"
                  />
                </div>
                <TextField
                  id="standard-helperText"
                  label="String Match"
                  value={preferredStringMatch}
                  onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                    setPreferredStringMatch(event.target.value);
                  }}
                  fullWidth
                />
                <div className="text-muted pe-5 mt-2">
                  Streams that include this string will be prioritized.
                  Evaluated after season packs (if selected). Useful to target
                  specific release groups.
                </div>
              </div>
              <Divider className="mt-3 mb-3" sx={{ borderColor: "black" }} />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={strictMatch}
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                      setStrictMatch(event.target.checked);
                    }}
                  />
                }
                label="Strict Matching"
              />
              <div className="text-muted mb-2 pe-5">
                If selected, only streams matching the above criteria are
                downloaded.
              </div>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={skipDownloaded}
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                      setSkipDownloaded(event.target.checked);
                    }}
                  />
                }
                label="Download Missing Episodes Only"
              />
              <div className="text-muted pe-5">
                If selected, episodes that are already downloaded to hound will
                be skipped even if it's selected above
              </div>
              <Divider className="mt-3 mb-3" sx={{ borderColor: "black" }} />
            </div>
          )}
        </div>
        <div className="d-flex justify-content-end mt-1">
          <Button onClick={onClose}>Cancel</Button>
          <Button onClick={handleConfirm}>Confirm</Button>
        </div>
      </div>
      <SelectStreamModal
        modalType="download-season"
        open={open && isSelectStreamModalOpen}
        setOpen={setIsSelectStreamModalOpen}
        setMainStream={setMainStream}
        streamData={streams}
      />
    </Dialog>
  );
}

function SelectedSeasonPack(props: any) {
  const { mainStream, setMainStream, setIsSelectStreamModalOpen } = props;
  if (!mainStream) return null;
  return (
    <>
      <div
        className="download-season-pack-card"
        onClick={() => {
          setIsSelectStreamModalOpen(true);
        }}
      >
        <div className="stream-info-card-title">{mainStream.title}</div>
        <div className="stream-info-card-subtitle">
          {mainStream.description}
        </div>
        <div className="stream-info-card-subtitle mb-2">
          info hash: {mainStream.info_hash}
        </div>
        <Chip label={mainStream.provider} size="small" />
      </div>
      <Button
        onClick={() => setMainStream(null)}
        variant="outlined"
        className="mb-2"
      >
        Clear Selection
      </Button>
    </>
  );
}

function EpisodeSelector(props: any) {
  const { episodesData, episodesToDownload, setEpisodesToDownload } = props;
  const ITEM_HEIGHT = 48;
  const ITEM_PADDING_TOP = 8;
  const MenuProps = {
    PaperProps: {
      style: {
        maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
        width: 250,
      },
    },
  };

  const handleChange = (event: SelectChangeEvent<number[]>) => {
    const value = event.target.value;
    const episodeNumbers =
      typeof value === "string"
        ? value.split(",").map(Number)
        : value.map(Number);

    setEpisodesToDownload(episodeNumbers.sort((a, b) => a - b));
  };

  return (
    <FormControl fullWidth className="mt-2">
      <InputLabel id="demo-multiple-checkbox-label">Episodes</InputLabel>
      <Select
        labelId="demo-multiple-checkbox-label"
        id="demo-multiple-checkbox"
        multiple
        value={episodesToDownload}
        onChange={handleChange}
        input={<OutlinedInput label="Episodes" />}
        renderValue={(selected) => selected.join(", ")}
        MenuProps={MenuProps}
      >
        {episodesData.map((ep: any) => {
          const selected = episodesToDownload?.includes(ep.episode_number);
          const SelectionIcon = selected
            ? CheckBoxIcon
            : CheckBoxOutlineBlankIcon;

          return (
            <MenuItem key={ep.episode_number} value={ep.episode_number}>
              <SelectionIcon
                fontSize="small"
                style={{ marginRight: 8, padding: 9, boxSizing: "content-box" }}
              />
              <ListItemText
                primary={`S${ep.season_number}E${ep.episode_number} - ${ep.media_title}`}
              />
            </MenuItem>
          );
        })}
      </Select>
    </FormControl>
  );
}

export default DownloadSeasonModal;
