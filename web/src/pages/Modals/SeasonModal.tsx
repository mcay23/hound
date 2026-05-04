import {
  Dialog,
  styled,
  Tooltip,
  tooltipClasses,
  TooltipProps,
  useMediaQuery,
  IconButton,
  useTheme,
  Fade,
  Button,
} from "@mui/material";
import axios from "axios";
import { useEffect, useState } from "react";
import "./SeasonModal.css";
import convertDateToReadable from "../../helpers/helpers";
import VisibilityIcon from "@mui/icons-material/Visibility";
import DownloadIcon from "@mui/icons-material/Download";
import DoneAllIcon from "@mui/icons-material/DoneAll";
import CreateHistoryModal from "./CreateHistoryModal";
import { paperPropsGlass, slotPropsGlass } from "./modalStyles";
import Dropdown from "react-bootstrap/Dropdown";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import { Spinner } from "react-bootstrap";
import { PlayArrowRounded } from "@mui/icons-material";
import toast from "react-hot-toast";
import DownloadSeasonModal from "./DownloadSeasonModal";

const offsetFix = {
  modifiers: [
    {
      name: "offset",
      options: {
        offset: [0, -10],
      },
    },
  ],
};

const BootstrapTooltip = styled(({ className, ...props }: TooltipProps) => (
  <Tooltip {...props} arrow classes={{ popper: className }} />
))(({ theme }) => ({
  [`& .${tooltipClasses.arrow}`]: {
    color: theme.palette.common.black,
  },
  [`& .${tooltipClasses.tooltip}`]: {
    backgroundColor: theme.palette.common.black,
  },
}));

type WatchProgressItem = {
  current_progress_seconds: number;
  total_duration_seconds: number;
  encoded_data: string;
};

import {
  useAddTVWatchHistoryMutation,
  useTVSeasonHistory,
} from "../../api/hooks/watchHistory";

function SeasonModal(props: any) {
  const {
    onClose,
    open,
    mediaSource,
    sourceID,
    seasonNumber,
    isStreamModalOpen,
  } = props;
  const handleClose = () => {
    setIsSeasonDataLoaded(false);
    onClose();
  };
  const [seasonData, setSeasonData] = useState({
    media_source: "",
    source_id: -1,
    release_date: "",
    episodes: [],
    id: -1,
    media_title: "",
    thumbnail_uri: "",
    season_number: -1,
    overview: "",
    watch_info: [],
  });

  const { data: historyData } = useTVSeasonHistory(
    mediaSource,
    sourceID,
    seasonNumber,
    open,
  );

  const [watchedEpisodes, setWatchedEpisodes] = useState<string[]>([]);
  const [watchProgress, setWatchProgress] = useState<
    Map<string, WatchProgressItem>
  >(() => new Map());
  const [isSeasonDataLoaded, setIsSeasonDataLoaded] = useState(false);
  const [isCreateHistoryModalOpen, setIsCreateHistoryModalOpen] =
    useState(false);
  const [isDownloadSeasonModalOpen, setIsDownloadSeasonModalOpen] =
    useState(false);

  const addTVWatchActivityMutation = useAddTVWatchHistoryMutation();

  useEffect(() => {
    if (historyData) {
      const latest = historyData.reduce(
        (a: any, b: any) =>
          new Date(a.rewatch_started_at) > new Date(b.rewatch_started_at)
            ? a
            : b,
        historyData[0],
      );
      if (latest) {
        const sourceIDs = (latest.watch_events || []).map(
          (event: any) => event.source_id,
        );
        setWatchedEpisodes(sourceIDs);
      }
    }
  }, [historyData]);

  const handleWatchEpisode = (
    season: number,
    episode: number,
    episodeID: string,
  ) => {
    addTVWatchActivityMutation.mutate(
      {
        mediaSource,
        sourceID,
        episodeIDs: [],
        seasonNumber: season,
        episodeNumber: episode,
      },
      {
        onSuccess: () => {
          toast.success("Episode marked as watched");
        },
        onError: (err) => {
          console.error(err);
          toast.error("Failed to mark episode as watched");
        },
      },
    );
  };

  const [historyModalType, setHistoryModalType] = useState<
    "season" | "episode"
  >("season");
  const [historyModalEpisodeIDs, setHistoryModalEpisodeIDs] = useState<
    number[]
  >([]);
  const [historyModalSeasonNumber, setHistoryModalSeasonNumber] = useState<
    number | undefined
  >();
  const [historyModalEpisodeNumber, setHistoryModalEpisodeNumber] = useState<
    number | undefined
  >();

  const handleOpenEpisodeHistoryModal = (
    episodeID: number,
    seasonNumber: number,
    episodeNumber: number,
  ) => {
    setHistoryModalType("episode");
    setHistoryModalEpisodeIDs([episodeID]);
    setHistoryModalSeasonNumber(seasonNumber);
    setHistoryModalEpisodeNumber(episodeNumber);
    setIsCreateHistoryModalOpen(true);
  };

  var seasonOverviewPlaceholder = "No description available.";
  if (isSeasonDataLoaded) {
    seasonOverviewPlaceholder = `Season ${seasonData.season_number} of ${props.mediaTitle}`;
    if (seasonData.season_number === 0) {
      seasonOverviewPlaceholder = "Special Episodes";
    }
  }

  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm")); // sm = 600px by default
  useEffect(() => {
    // no need to call on close
    if (open === false) return;
    if (seasonNumber < 0) return;

    // season 0 is used for extras, specials sometimes
    const loadData = async () => {
      const seasonRes = await axios
        .get(`/api/v1/tv/${mediaSource}-${sourceID}/season/${seasonNumber}`)
        .catch((err) => {
          console.log(err);
        });
      if (!seasonRes) return;
      setSeasonData(seasonRes.data);
      setIsSeasonDataLoaded(true);

      // get watch progress
      axios
        .get(
          `/api/v1/tv/${mediaSource}-${sourceID}/season/${seasonNumber}/playback`,
        )
        .then((progressRes) => {
          // overwrite state each time
          if (progressRes.data) {
            const progressMap = new Map<string, WatchProgressItem>();
            progressRes.data.forEach((item: any) => {
              progressMap.set(item.episode_source_id, {
                current_progress_seconds: item.current_progress_seconds,
                total_duration_seconds: item.total_duration_seconds,
                encoded_data: item.encoded_data,
              });
            });
            setWatchProgress(progressMap);
          } else {
            // null progress is also a valid response
            setWatchProgress(new Map<string, WatchProgressItem>());
          }
        })
        .catch((err) => {
          console.log(err);
        });
    };
    loadData();
  }, [seasonNumber, mediaSource, sourceID, open, isStreamModalOpen]);

  return (
    <>
      {isSeasonDataLoaded ? (
        <Dialog
          onClose={handleClose}
          open={open}
          className="season-modal-dialog"
          maxWidth={false}
          fullScreen={fullScreen}
          TransitionComponent={Fade}
          TransitionProps={{ timeout: 0 }}
          slotProps={slotPropsGlass}
          PaperProps={paperPropsGlass}
        >
          <div className="season-modal-container">
            <div className="season-modal-info-container">
              {seasonData.thumbnail_uri ? (
                <img
                  className="season-modal-poster"
                  src={seasonData.thumbnail_uri}
                  alt={seasonData.media_title}
                />
              ) : (
                <div className={"season-modal-poster item-card-no-thumbnail"}>
                  {seasonData.media_title}
                </div>
              )}
              <div className="season-modal-info-inner">
                <div className="season-modal-info-title">
                  {seasonData.media_title}
                  {seasonData.release_date ? (
                    <>
                      <span
                        className="media-item-separator"
                        style={{ color: "gray" }}
                      >
                        |
                      </span>
                      <span className="season-modal-info-date">
                        {seasonData.release_date?.slice(0, 4)}
                      </span>
                    </>
                  ) : (
                    ""
                  )}
                </div>
                <hr className="" />
                <div className="season-modal-info-description">
                  {seasonData.overview
                    ? seasonData.overview
                    : seasonOverviewPlaceholder}
                </div>
                <div className="season-modal-actions-container">
                  <span className="season-modal-info-button">
                    <BootstrapTooltip
                      title={
                        <span className="media-page-tv-header-button-tooltip-title">
                          Mark Season As Watched
                        </span>
                      }
                      PopperProps={offsetFix}
                    >
                      <IconButton
                        onClick={() => {
                          setHistoryModalType("season");
                          setHistoryModalEpisodeIDs(
                            seasonData.episodes.map((ep: any) => ep.source_id),
                          );
                          setHistoryModalSeasonNumber(undefined);
                          setHistoryModalEpisodeNumber(undefined);
                          setIsCreateHistoryModalOpen(true);
                        }}
                      >
                        <VisibilityIcon />
                      </IconButton>
                    </BootstrapTooltip>
                    <BootstrapTooltip
                      title={
                        <span className="media-page-tv-header-button-tooltip-title">
                          Download Season
                        </span>
                      }
                      PopperProps={offsetFix}
                    >
                      <IconButton
                        onClick={() => {
                          if (isSeasonDataLoaded) {
                            setIsDownloadSeasonModalOpen(true);
                          }
                        }}
                      >
                        <DownloadIcon />
                      </IconButton>
                    </BootstrapTooltip>
                  </span>
                  {/* <span className="season-modal-info-button">
                  <BootstrapTooltip
                    title={
                      <span className="media-page-tv-header-button-tooltip-title">
                        Add Review
                      </span>
                    }
                    PopperProps={offsetFix}
                  >
                    <IconButton>
                      <ChatIcon />
                    </IconButton>
                  </BootstrapTooltip>
                </span> */}
                </div>
              </div>
            </div>
            <div className="season-episode-card-container">
              {seasonData.episodes.map((episode: any) => {
                return EpisodeCard(
                  episode,
                  watchedEpisodes.includes(episode["source_id"]),
                  watchProgress.get(episode["source_id"]),
                  handleWatchEpisode,
                  props.handleStreamButtonClick,
                  props.isStreamButtonLoading,
                  props.isStreamSelectButtonLoading,
                  handleOpenEpisodeHistoryModal,
                );
              })}
            </div>
          </div>
          <CreateHistoryModal
            onClose={() => {
              setIsCreateHistoryModalOpen(false);
            }}
            open={isCreateHistoryModalOpen}
            type={historyModalType}
            mediaSource={mediaSource}
            sourceID={sourceID}
            episodeIDs={historyModalEpisodeIDs}
            seasonNumber={historyModalSeasonNumber}
            episodeNumber={historyModalEpisodeNumber}
          />
          <DownloadSeasonModal
            onClose={() => {
              setIsDownloadSeasonModalOpen(false);
            }}
            open={isDownloadSeasonModalOpen}
            mediaSource={mediaSource}
            sourceID={sourceID}
            seasonNumber={seasonNumber}
            seasonData={seasonData}
          />
        </Dialog>
      ) : (
        ""
      )}
    </>
  );
}

function EpisodeCard(
  episode: any,
  watched: boolean,
  watchProgress: WatchProgressItem | undefined,
  handleWatchEpisode: Function,
  handleStreamButtonClick: Function,
  isStreamButtonLoading: boolean,
  isStreamSelectButtonLoading: boolean,
  handleOpenEpisodeHistoryModal: Function,
) {
  var episodeNumber =
    episode.season_number.toString() &&
    episode.episode_number.toString() &&
    `S${episode.season_number}E${episode.episode_number}`.replace(
      "S0E",
      "Special #",
    );
  return (
    <div
      className="episode-card-container"
      key={episode.media_source + "-" + episode.source_id}
    >
      <div
        className="episode-card-img-container"
        onClick={() => {
          if (isStreamButtonLoading || isStreamSelectButtonLoading) {
            return;
          }
          handleStreamButtonClick(
            episode.season_number,
            episode.episode_number,
            "direct",
            episode.source_id,
            watchProgress?.current_progress_seconds,
            watchProgress?.encoded_data,
          );
        }}
      >
        <img
          src={episode.thumbnail_uri}
          alt={episode.media_title}
          className="episode-card-img hide-alt"
          loading="lazy"
          onError={({ currentTarget }) => {
            currentTarget.onerror = null; // prevents looping
            currentTarget.src = "/landscape-placeholder.jpg";
          }}
        />
        <div className="episode-card-img-play-overlay">
          <div className="episode-card-img-play-icon">
            <PlayArrowRounded sx={{ fontSize: "90px" }} />
          </div>
        </div>
        {watchProgress && (
          <>
            <div className="episode-card-progress-pill">
              <div className="episode-card-progress-pill-text">
                {Math.ceil(
                  (watchProgress.total_duration_seconds -
                    watchProgress.current_progress_seconds) /
                    60,
                )}
                {"m left"}
              </div>
            </div>
            <div className="episode-card-progress-bar-container">
              <div
                className="episode-card-progress-bar"
                style={{
                  width: `${
                    (watchProgress.current_progress_seconds /
                      watchProgress.total_duration_seconds) *
                    100
                  }%`,
                }}
              />
            </div>
          </>
        )}
      </div>
      <div className="episode-card-content">
        <div className="episode-card-title">{episode.media_title}</div>
        {episode.release_date && (
          <div className="episode-card-date">
            {episodeNumber}
            {episodeNumber && episode.release_date && "     ⸱     "}
            {convertDateToReadable(episode.release_date)}
          </div>
        )}
        <div className="episode-card-description">
          {episode.overview ? episode.overview : "No description available."}
        </div>
      </div>
      <div className="episode-card-actions">
        <Dropdown
          align="end"
          autoClose="outside"
          id="season-episode-card-dropdown-container"
        >
          <Dropdown.Toggle
            as={Button}
            variant="light"
            id="season-episode-card-dropdown"
            className="border-0 p-0"
            style={{ minWidth: "auto" }}
          >
            <MoreVertIcon />
          </Dropdown.Toggle>
          <Dropdown.Menu>
            <Dropdown.Item
              onClick={() => {
                handleStreamButtonClick(
                  episode.season_number,
                  episode.episode_number,
                  "direct",
                  episode.source_id,
                  watchProgress?.current_progress_seconds,
                  watchProgress?.encoded_data,
                );
              }}
            >
              {isStreamButtonLoading ? (
                <div className="d-flex justify-content-center">
                  <Spinner
                    animation="border"
                    size="sm"
                    role="status"
                    id="stream-select-button-loading"
                  >
                    <span className="visually-hidden">Loading...</span>
                  </Spinner>
                </div>
              ) : (
                "Play Episode"
              )}
            </Dropdown.Item>
            <Dropdown.Item
              onClick={() => {
                handleStreamButtonClick(
                  episode.season_number,
                  episode.episode_number,
                  "select",
                  episode.source_id,
                  watchProgress?.current_progress_seconds,
                  watchProgress?.encoded_data,
                );
              }}
            >
              {isStreamSelectButtonLoading ? (
                <div className="d-flex justify-content-center">
                  <Spinner
                    animation="border"
                    size="sm"
                    role="status"
                    id="stream-select-button-loading"
                  >
                    <span className="visually-hidden">Loading...</span>
                  </Spinner>
                </div>
              ) : (
                "Select Stream..."
              )}
            </Dropdown.Item>
            <Dropdown.Item
              onClick={() => {
                handleOpenEpisodeHistoryModal(
                  episode.source_id,
                  episode.season_number,
                  episode.episode_number,
                );
              }}
            >
              Add Watch History...
            </Dropdown.Item>
          </Dropdown.Menu>
        </Dropdown>
        {watched ? (
          <IconButton disabled>
            <DoneAllIcon />
          </IconButton>
        ) : (
          <BootstrapTooltip
            title={
              <span className="media-page-tv-header-button-tooltip-title">
                Mark as Watched
              </span>
            }
            PopperProps={offsetFix}
          >
            <IconButton
              onClick={() => {
                handleWatchEpisode(
                  episode.season_number,
                  episode.episode_number,
                  episode.source_id,
                );
              }}
            >
              <VisibilityIcon />
            </IconButton>
          </BootstrapTooltip>
        )}
      </div>
    </div>
  );
}

export default SeasonModal;
