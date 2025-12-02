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
import StreamModal from "../Modals/StreamModal";
import toast from "react-hot-toast";
import { paperPropsGlass, slotPropsGlass } from "./modalStyles";
import Dropdown from "react-bootstrap/Dropdown";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import SelectStreamModal from "./StreamSelectModal";
import { Spinner } from "react-bootstrap";
import { PlayArrow, PlayArrowRounded } from "@mui/icons-material";

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

function SeasonModal(props: any) {
  const { onClose, open, mediaSource, sourceID, seasonNumber } = props;
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
  const [isSeasonDataLoaded, setIsSeasonDataLoaded] = useState(false);
  const [isCreateHistoryModalOpen, setisCreateHistoryModalOpen] =
    useState(false);
  const [isStreamModalOpen, setIsStreamModalOpen] = useState(false);
  const [isSelectStreamModalOpen, setIsSelectStreamModalOpen] = useState(false);
  const [isStreamButtonLoading, setIsStreamButtonLoading] = useState(false);
  const [isStreamSelectButtonLoading, setIsStreamSelectButtonLoading] =
    useState(false);
  const [streams, setStreams] = useState<any>(null);
  const [mainStream, setMainStream] = useState<any>(null);
  const handleWatchEpisode = (tmdbID: number) => {
    var payload = {
      episode_ids: [tmdbID],
      action_type: "watch",
    };
    axios
      .post(`/api/v1/tv/${mediaSource}-${sourceID}/history`, payload)
      .then(() => {
        setWatchedEpisodes([...watchedEpisodes, tmdbID]);
      })
      .catch((err) => {
        console.log(err);
      });
  };
  const handleStreamButtonClick = (
    season: number,
    episode: number,
    mode: string
  ) => {
    if (mode === "direct") {
      setIsStreamButtonLoading(true);
    } else if (mode === "select") {
      setIsStreamSelectButtonLoading(true);
    }
    axios
      .get(
        `/api/v1/tv/${mediaSource}-${sourceID}/providers?season=${season}&episode=${episode}`
      )
      .then((res) => {
        setStreams(res.data);
        let numStreams = res.data.data.providers[0].streams.length;
        if (numStreams > 0) {
          setMainStream(res.data.data.providers[0].streams[0]);
          if (mode === "direct") {
            setIsStreamModalOpen(true);
          } else {
            setIsSelectStreamModalOpen(true);
          }
        } else {
          toast.error("No streams found");
        }
      })
      .catch((err) => {
        if (err.response.status === 500) {
          toast.error("Error getting streams");
        }
      })
      .finally(() => {
        if (mode === "direct") {
          setIsStreamButtonLoading(false);
        } else if (mode === "select") {
          setIsStreamSelectButtonLoading(false);
        }
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

    // TODO try to cache the call, but watch info might change
    // check data is loaded
    // season 0 is used for extras, specials sometimes
    const loadData = async () => {
      try {
        const seasonRes = await axios.get(
          `/api/v1/tv/${mediaSource}-${sourceID}/season/${seasonNumber}`
        );
        setSeasonData(seasonRes.data);
        setIsSeasonDataLoaded(true);
        // get watch data
        const historyRes = await axios.get(
          `/api/v1/tv/${mediaSource}-${sourceID}/season/${seasonNumber}/history`
        );
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
      } catch (err: any) {
        console.log(err.response);
      }
    };
    loadData();
  }, [seasonNumber, mediaSource, sourceID, open]);

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
                  handleWatchEpisode,
                  handleStreamButtonClick,
                  isStreamButtonLoading,
                  isStreamSelectButtonLoading
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
          <StreamModal
            setOpen={setIsStreamModalOpen}
            open={isStreamModalOpen}
            streamDetails={mainStream}
            streams={streams?.data}
          />
          <SelectStreamModal
            setOpen={setIsSelectStreamModalOpen}
            open={isSelectStreamModalOpen}
            streamData={streams}
            setMainStream={setMainStream}
            setIsStreamModalOpen={setIsStreamModalOpen}
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
          handleStreamButtonClick(
            episode.season_number,
            episode.episode_number,
            "select"
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
                  "direct"
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
                  "select"
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
                handleWatchEpisode(episode.id);
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
