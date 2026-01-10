import "./MediaPage.css";
import PlaylistAddIcon from "@mui/icons-material/PlaylistAdd";
import HistoryIcon from "@mui/icons-material/History";
import CachedIcon from "@mui/icons-material/Cached";
import {
  IconButton,
  Skeleton,
  styled,
  Tooltip,
  tooltipClasses,
  TooltipProps,
} from "@mui/material";
import { useEffect, useState } from "react";
import AddToCollectionModal from "../Modals/AddToCollectionModal";
import HorizontalSection from "../Home/HorizontalSection";
import VideoModal from "../Modals/VideoModal";
import SeasonModal from "../Modals/SeasonModal";
import Reviews from "../Comments/Reviews";
import HistoryModal from "../Modals/HistoryModal";
import ConfirmRewatchModal from "../Modals/ConfirmRewatchModal";
import StreamModal from "../Modals/StreamModal";
import SelectStreamModal from "../Modals/StreamSelectModal";
import axios from "axios";
import toast from "react-hot-toast";
import { Dropdown, Spinner, SplitButton } from "react-bootstrap";

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

function MediaPageTV(props: any) {
  const [isCollectionModalOpen, setIsCollectionModalOpen] = useState(false);
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false);
  const [isConfirmRewatchModalOpen, setIsConfirmRewatchModalOpen] =
    useState(false);
  const [videoKey, setVideoKey] = useState("");
  const [seasonModal, setSeasonModal] = useState(-1);
  const [isSeasonModalOpen, setIsSeasonModalOpen] = useState(false);
  const [isHistoryModalOpen, setIsHistoryModalOpen] = useState(false);
  const [isPosterLoaded, setIsPosterLoaded] = useState(false);

  const [isStreamModalOpen, setIsStreamModalOpen] = useState(false);
  const [isSelectStreamModalOpen, setIsSelectStreamModalOpen] = useState(false);
  const [isStreamButtonLoading, setIsStreamButtonLoading] = useState(false);
  const [isStreamSelectButtonLoading, setIsStreamSelectButtonLoading] =
    useState(false);
  const [streams, setStreams] = useState<any>(null);
  const [mainStream, setMainStream] = useState<any>(null);
  const [streamStartTime, setStreamStartTime] = useState(0);
  const [continueWatchingData, setContinueWatchingData] = useState<any>(null);

  var styles = {
    noBackdrop: {
      background:
        "linear-gradient(rgba(24, 11, 111, 1) 5%, rgba(0, 0, 0, 0.8) 30%, rgba(0, 0, 0, 0.3) 70%)",
      backgroundColor: "black",
    },
    withBackdrop: {
      // backgroundColor: "blue",
      backgroundImage:
        "linear-gradient(rgba(24, 11, 111, 1) 9%, rgba(0, 0, 0, 0.8) 30%, rgba(0, 0, 0, 0.3) 70%), url(" +
        props.data.backdrop_url +
        ")",
      backgroundAttachment: "fixed",
      backgroundSize: "cover",
      // animation: "backgroundScroll 40s linear infinite",
    },
    opacityBackdrop: {
      // backgroundColor: "blue",
      backgroundImage:
        "linear-gradient(rgba(255, 255, 255, 0.94), rgba(255, 255, 255, 0.94)), url(" +
        props.data.backdrop_url +
        ")",
      backgroundAttachment: "fixed",
      backgroundSize: "cover",
    },
  };
  // handle variables for display
  var releaseYear = props.data.first_air_date.slice(0, 4);
  var genres = props.data.genres
    .map((item: any) => {
      return item.name;
    })
    .join(", ");
  var runtime = "";
  const lf = new Intl.ListFormat("en");
  var creators = lf.format(
    props.data.created_by.map((item: any) => {
      return item.name;
    })
  );
  if (props.data.episode_run_time.length > 0) {
    if (props.data.episode_run_time[0] >= 60) {
      runtime =
        Math.floor(props.data.episode_run_time[0] / 60) +
        "h " +
        (props.data.episode_run_time[0] % 60) +
        "m";
    } else {
      runtime = props.data.episode_run_time[0] + "m";
    }
  }
  // handle actor profiles
  var creditsList = props.data.credits.cast.map((item: any) => {
    return {
      thumbnail_url: item.profile_path,
      credits: {
        name: item.name,
        character: item.character,
        id: item.id,
      },
      id: item.credit_id,
    };
  });
  // if specials exist (season number 0), move to end of array for displaying
  // if (props.data.seasons && props.data.seasons[0].season_number === 0) {
  //   props.data.seasons.push(props.data.seasons.shift());
  // }
  // modal functions
  const handleVideoButtonClick = (key: string) => {
    setIsVideoModalOpen(true);
    setVideoKey(key);
  };
  const handleSeasonButtonClick = (key: number) => {
    setSeasonModal(key);
    setIsSeasonModalOpen(true);
  };

  useEffect(() => {
    if (props.data) {
      const mediaSource = props.data.media_source;
      const sourceID = props.data.source_id;
      axios
        .get(`/api/v1/tv/${mediaSource}-${sourceID}/continue_watching`)
        .then((res) => {
          if (res.data.status === "success") {
            setContinueWatchingData(res.data.data);
          }
        })
        .catch((err) => {
          console.error("Failed to fetch continue watching data", err);
        });
    }
  }, [
    props.data,
    isStreamModalOpen,
    isSeasonModalOpen,
    isConfirmRewatchModalOpen,
  ]);

  const handleStreamButtonClick = (
    season: number,
    episode: number,
    mode: string,
    episodeID: number,
    overrideStartTime?: number,
    overrideEncodedData?: string
  ) => {
    if (mode === "direct") {
      setIsStreamButtonLoading(true);
    } else if (mode === "select") {
      setIsStreamSelectButtonLoading(true);
    }
    const mediaSource = props.data.media_source;
    const sourceID = props.data.source_id;
    const searchProvidersToast = toast.loading("Searching providers...");
    // if we have current watch data, use encodedData to match a stream
    const requestProviderStream = (startTime: number, encodedData: string) => {
      setStreamStartTime(startTime);
      return axios
        .get(
          `/api/v1/tv/${mediaSource}-${sourceID}/providers?season=${season}&episode=${episode}`
        )
        .then((res) => {
          toast.dismiss(searchProvidersToast);
          setStreams(res.data);
          let numStreams = res.data?.data?.providers[0]?.streams?.length;
          if (numStreams > 0) {
            let selectedStream = res.data.data.providers[0].streams[0];
            // if we have watch progress, set this as the main stream
            // note this doesn't handle different hosts/protocols, eg. if
            // the urls are different even if they are the same file, it won't match
            if (mode === "direct" && encodedData) {
              const matchingStream = res.data.data.providers[0].streams.find(
                (stream: any) => stream.encoded_data === encodedData
              );
              if (matchingStream) {
                selectedStream = matchingStream;
              }
            }
            setMainStream(selectedStream);
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
          toast.dismiss(searchProvidersToast);
          console.error("Failed to fetch providers", err);
          toast.error("Failed to fetch providers");
        });
    };
    if (overrideStartTime !== undefined) {
      requestProviderStream(
        overrideStartTime,
        overrideEncodedData || ""
      ).finally(() => {
        if (mode === "direct") {
          setIsStreamButtonLoading(false);
        } else if (mode === "select") {
          setIsStreamSelectButtonLoading(false);
        }
      });
      return;
    }
    // no override, get playback progress (season modal case)
    axios
      .get(`/api/v1/tv/${mediaSource}-${sourceID}/season/${season}/playback`)
      .then((progressRes) => {
        let startTime = 0;
        let encodedData = "";
        if (progressRes.data?.data && episodeID !== -1) {
          const episodeProgress = progressRes.data.data.find(
            (item: any) => parseInt(item.episode_source_id, 10) === episodeID
          );
          if (episodeProgress) {
            startTime = episodeProgress.current_progress_seconds || 0;
            encodedData = episodeProgress.encoded_data;
          }
        }
        return requestProviderStream(startTime, encodedData);
      })
      .catch((err) => {
        toast.error("Failed to get playback progress " + err, {
          id: searchProvidersToast,
        });
      })
      .finally(() => {
        if (mode === "direct") {
          setIsStreamButtonLoading(false);
        } else if (mode === "select") {
          setIsStreamSelectButtonLoading(false);
        }
      });
  };

  const handleHeaderPlayClick = (mode: string) => {
    if (!continueWatchingData) {
      // Fallback if data not loaded
      handleStreamButtonClick(1, 1, mode, -1);
      return;
    }
    const { watch_action_type, next_episode, watch_progress } =
      continueWatchingData;

    if (watch_action_type === "resume" && watch_progress) {
      handleStreamButtonClick(
        watch_progress.season_number,
        watch_progress.episode_number,
        mode,
        parseInt(watch_progress.episode_source_id, 10),
        watch_progress.current_progress_seconds,
        watch_progress.encoded_data
      );
    } else if (watch_action_type === "next_episode" && next_episode) {
      handleStreamButtonClick(
        next_episode.season_number,
        next_episode.episode_number,
        mode,
        parseInt(next_episode.episode_source_id, 10)
      );
    } else {
      // Fallback, episodeID is only used to find progress
      // fine to ignore
      handleStreamButtonClick(1, 1, mode, -1);
    }
  };
  if (props.data.media_title) {
    var yearString = props.data.first_air_date
      ? `(${props.data.first_air_date.slice(0, 4)})`
      : "";
    document.title = props.data.media_title + " " + yearString + " - Hound";
  }
  var continueWatchingText = "▶ Play S1E1";
  if (continueWatchingData) {
    const { watch_action_type, next_episode, watch_progress } =
      continueWatchingData;
    if (watch_action_type === "resume" && watch_progress) {
      continueWatchingText = `▶ Resume S${watch_progress.season_number}E${watch_progress.episode_number}`;
    } else if (watch_action_type === "next_episode" && next_episode) {
      continueWatchingText = `▶ Play S${next_episode.season_number}E${next_episode.episode_number}`;
    }
  }
  return (
    <>
      <div
        className="media-page-tv-header"
        style={
          props.data.backdrop_url ? styles.withBackdrop : styles.noBackdrop
        }
      >
        <div className="media-page-tv-header-container">
          <div className="media-page-tv-inline-container">
            <div className="media-page-tv-poster-container">
              {!isPosterLoaded && props.data.poster_url && (
                <Skeleton
                  variant="rounded"
                  className="rounded media-page-tv-poster-skeleton"
                  animation="wave"
                />
              )}
              {props.data.poster_url ? (
                <img
                  className={
                    "media-page-tv-poster " + (!isPosterLoaded && "d-none")
                  }
                  src={props.data.poster_url}
                  alt={props.data.media_title}
                  onLoad={() => setIsPosterLoaded(true)}
                />
              ) : (
                <div className="media-page-tv-poster">
                  {props.data.media_title}
                </div>
              )}
            </div>
            <div className="media-page-tv-header-info">
              <div className="media-page-tv-header-title">
                {props.data.media_title}
                <span className="media-page-tv-header-year">
                  {" "}
                  {releaseYear.length !== 4 ? "" : "(" + releaseYear + ")"}
                </span>
              </div>
              <div className="media-page-tv-header-genres">
                {props.data.status === "Ended"
                  ? "Finished Airing"
                  : props.data.status}
                {props.data.status && (runtime || genres) ? "     ⸱     " : ""}
                {runtime}
                {runtime && genres ? "     ⸱     " : ""}
                {genres}
              </div>
              <div className="media-page-tv-header-overview">
                {props.data.overview
                  ? props.data.overview
                  : "No description available."}
              </div>
              <div className="media-page-tv-header-credits">
                {creators ? "by " + creators : ""}
              </div>
              <div className="media-page-tv-header-button-container">
                <SplitButton
                  title={
                    isStreamButtonLoading ? (
                      <Spinner
                        animation="grow"
                        size="sm"
                        role="status"
                        className="stream-play-button-spinner"
                      />
                    ) : (
                      continueWatchingText
                    )
                  }
                  autoClose="outside"
                  className="stream-play-button"
                  onClick={() => {
                    handleHeaderPlayClick("direct");
                  }}
                >
                  <Dropdown.Item
                    eventKey="1"
                    onClick={() => {
                      handleHeaderPlayClick("select");
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
                </SplitButton>
                <BootstrapTooltip
                  title={
                    <span className="media-page-tv-header-button-tooltip-title">
                      Add To Collection
                    </span>
                  }
                  PopperProps={offsetFix}
                >
                  <IconButton
                    onClick={() => {
                      setIsCollectionModalOpen(true);
                    }}
                  >
                    <PlaylistAddIcon />
                  </IconButton>
                </BootstrapTooltip>
                <BootstrapTooltip
                  title={
                    <span className="media-page-tv-header-button-tooltip-title">
                      View Watch History
                    </span>
                  }
                  PopperProps={offsetFix}
                >
                  <IconButton
                    onClick={() => {
                      setIsHistoryModalOpen(true);
                    }}
                  >
                    <HistoryIcon id="media-page-tv-header-track-button" />
                  </IconButton>
                </BootstrapTooltip>
                <BootstrapTooltip
                  title={
                    <span className="media-page-tv-header-button-tooltip-title">
                      Rewatch Show
                    </span>
                  }
                  PopperProps={offsetFix}
                >
                  <IconButton
                    onClick={() => {
                      setIsConfirmRewatchModalOpen(true);
                    }}
                  >
                    <CachedIcon />
                  </IconButton>
                </BootstrapTooltip>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div className="media-page-tv-main" style={styles.opacityBackdrop}>
        <HorizontalSection
          items={creditsList}
          header={"Cast"}
          itemType="cast"
          itemOnClick={undefined}
        />
        <HorizontalSection
          items={props.data.videos.results}
          header={"Videos"}
          itemType="video"
          itemOnClick={handleVideoButtonClick}
        />
        <HorizontalSection
          items={props.data.seasons}
          header={"Seasons"}
          itemType="seasons"
          itemOnClick={handleSeasonButtonClick}
        />
        <Reviews data={props.data.comments} />
      </div>
      <div className="media-page-tv-footer" style={styles.withBackdrop} />
      <AddToCollectionModal
        onClose={() => {
          setIsCollectionModalOpen(false);
        }}
        open={isCollectionModalOpen}
        item={props.data}
      />
      <VideoModal
        onClose={() => {
          setIsVideoModalOpen(false);
        }}
        open={isVideoModalOpen}
        videoKey={videoKey}
      />
      <SeasonModal
        onClose={() => {
          setIsSeasonModalOpen(false);
        }}
        open={isSeasonModalOpen}
        mediaSource={props.data ? props.data.media_source : undefined}
        sourceID={props.data ? props.data.source_id : undefined}
        seasonNumber={seasonModal}
        mediaTitle={props.data.media_title}
        handleStreamButtonClick={handleStreamButtonClick}
        isStreamButtonLoading={isStreamButtonLoading}
        isStreamSelectButtonLoading={isStreamSelectButtonLoading}
        isStreamModalOpen={isStreamModalOpen}
      />
      <HistoryModal
        onClose={() => {
          setIsHistoryModalOpen(false);
        }}
        open={isHistoryModalOpen}
        data={props.data}
      />
      <ConfirmRewatchModal
        onClose={() => {
          setIsConfirmRewatchModalOpen(false);
        }}
        open={isConfirmRewatchModalOpen}
        mediaSource={props.data ? props.data.media_source : undefined}
        sourceID={props.data ? props.data.source_id : undefined}
      />
      <StreamModal
        setOpen={setIsStreamModalOpen}
        open={isStreamModalOpen}
        streamDetails={mainStream}
        startTime={streamStartTime}
        streams={streams?.data}
      />
      <SelectStreamModal
        setOpen={setIsSelectStreamModalOpen}
        open={isSelectStreamModalOpen}
        streamData={streams}
        setMainStream={setMainStream}
        setIsStreamModalOpen={setIsStreamModalOpen}
      />
    </>
  );
}

export default MediaPageTV;
