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
import DoneAllIcon from "@mui/icons-material/DoneAll";
import CreateHistoryModal from "./CreateHistoryModal";
import { paperPropsGlass, slotPropsGlass } from "./modalStyles";
import Dropdown from "react-bootstrap/Dropdown";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import { Spinner } from "react-bootstrap";
import { PlayArrowRounded } from "@mui/icons-material";
import toast from "react-hot-toast";

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
    season: {
      air_date: "",
      episodes: [],
      id: -1,
      name: "",
      poster_path: "",
      season_number: -1,
      overview: "",
    },
    watch_info: [],
  });
  const [watchedEpisodes, setWatchedEpisodes] = useState<number[]>([]);
  const [watchProgress, setWatchProgress] = useState<
    Map<number, WatchProgressItem>
  >(() => new Map());
  const [isSeasonDataLoaded, setIsSeasonDataLoaded] = useState(false);
  const [isCreateHistoryModalOpen, setisCreateHistoryModalOpen] =
    useState(false);
  const handleWatchEpisode = (
    season: number,
    episode: number,
    episodeID: number
  ) => {
    // don't send in episode_id array, since this doesn't delete
    // resume progress if mark as watched
    var payload = {
      season_number: season,
      episode_number: episode,
      action_type: "watch",
    };
    axios
      .post(`/api/v1/tv/${mediaSource}-${sourceID}/history`, payload)
      .then((res) => {
        setWatchedEpisodes([...watchedEpisodes, episodeID]);
        if (res.status === 200) {
          toast.success("Episode marked as watched");
        }
      })
      .catch((err) => {
        console.log(err);
        toast.error("Failed to mark episode as watched");
      });
  };
  var seasonOverviewPlaceholder = "No description available.";
  if (isSeasonDataLoaded) {
    seasonOverviewPlaceholder = `Season ${seasonData.season.season_number} of ${props.mediaTitle}`;
    if (seasonData.season.season_number === 0) {
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
      // get watch data
      axios
        .get(
          `/api/v1/tv/${mediaSource}-${sourceID}/season/${seasonNumber}/history`
        )
        .then((historyRes) => {
          if (historyRes.data.data) {
            const latest = historyRes.data.data.reduce((a: any, b: any) =>
              new Date(a.rewatch_started_at) > new Date(b.rewatch_started_at)
                ? a
                : b
            );
            const sourceIDs = (latest.watch_events || [])
              .map((event: any) => parseInt(event.source_id, 10))
              .filter((tmdbID: number) => !isNaN(tmdbID));
            setWatchedEpisodes(sourceIDs);
          }
        })
        .catch((err) => {
          console.log(err);
        });
      // get watch progress
      axios
        .get(
          `/api/v1/tv/${mediaSource}-${sourceID}/season/${seasonNumber}/playback`
        )
        .then((progressRes) => {
          // overwrite state each time
          if (progressRes.data.data) {
            const progressMap = new Map<number, WatchProgressItem>();
            progressRes.data.data.forEach((item: any) => {
              const episodeIDNum = parseInt(item.episode_id, 10);
              progressMap.set(episodeIDNum, {
                current_progress_seconds: item.current_progress_seconds,
                total_duration_seconds: item.total_duration_seconds,
                encoded_data: item.encoded_data,
              });
            });
            setWatchProgress(progressMap);
          } else {
            // null progress is also a valid response
            setWatchProgress(new Map<number, WatchProgressItem>());
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
              {seasonData.season.poster_path ? (
                <img
                  className="season-modal-poster"
                  src={seasonData.season.poster_path}
                  alt={seasonData.season.name}
                />
              ) : (
                <div className={"season-modal-poster item-card-no-thumbnail"}>
                  {seasonData.season.name}
                </div>
              )}
              <div className="season-modal-info-inner">
                <div className="season-modal-info-title">
                  {seasonData.season.name}
                  {seasonData.season.air_date ? (
                    <>
                      <span
                        className="media-item-separator"
                        style={{ color: "gray" }}
                      >
                        |
                      </span>
                      <span className="season-modal-info-date">
                        {seasonData.season.air_date.slice(0, 4)}
                      </span>
                    </>
                  ) : (
                    ""
                  )}
                </div>
                <hr className="" />
                <div className="season-modal-info-description">
                  {seasonData.season.overview
                    ? seasonData.season.overview
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
                          setisCreateHistoryModalOpen(true);
                        }}
                      >
                        <VisibilityIcon />
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
              {seasonData.season.episodes.map((episode) => {
                return EpisodeCard(
                  episode,
                  watchedEpisodes.includes(episode["id"]),
                  watchProgress.get(episode["id"]),
                  handleWatchEpisode,
                  props.handleStreamButtonClick,
                  props.isStreamButtonLoading,
                  props.isStreamSelectButtonLoading
                );
              })}
            </div>
          </div>
          <CreateHistoryModal
            onClose={() => {
              setisCreateHistoryModalOpen(false);
            }}
            open={isCreateHistoryModalOpen}
            type={"season"}
            seasonNumber={seasonData.season.season_number}
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
  isStreamSelectButtonLoading: boolean
) {
  var episodeNumber =
    episode.season_number.toString() &&
    episode.episode_number.toString() &&
    `S${episode.season_number}E${episode.episode_number}`.replace(
      "S0E",
      "Special #"
    );
  return (
    <div className="episode-card-container" key={episode.id}>
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
            episode.id
          );
        }}
      >
        <img
          src={episode.still_path}
          alt={episode.name}
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
                    60
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
        <div className="episode-card-title">{episode.name}</div>
        {episode.air_date && (
          <div className="episode-card-date">
            {episodeNumber}
            {episodeNumber && episode.air_date && "     ⸱     "}
            {convertDateToReadable(episode.air_date)}
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
                  episode.id
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
                  episode.id
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
            <Dropdown.Item>Mark as Watched</Dropdown.Item>
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
                  episode.id
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
