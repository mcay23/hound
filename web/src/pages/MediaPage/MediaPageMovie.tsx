import "./MediaPage.css";
import PlaylistAddIcon from "@mui/icons-material/PlaylistAdd";
import VisibilityIcon from "@mui/icons-material/Visibility";
import HistoryIcon from "@mui/icons-material/History";
import {
  IconButton,
  Skeleton,
  styled,
  Tooltip,
  tooltipClasses,
  TooltipProps,
} from "@mui/material";
import { useState } from "react";
import AddToCollectionModal from "../Modals/AddToCollectionModal";
import HorizontalSection from "../Home/HorizontalSection";
import VideoModal from "../Modals/VideoModal";
import convertDateToReadable from "../../helpers/helpers";
import Reviews from "../Comments/Reviews";
import CreateHistoryModal from "../Modals/CreateHistoryModal";
import HistoryModal from "../Modals/HistoryModal";
import StreamModal from "../Modals/StreamModal";
import axios from "axios";
import toast from "react-hot-toast";
import { Dropdown, Spinner, SplitButton } from "react-bootstrap";
import SelectStreamModal from "../Modals/StreamSelectModal";

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

function MediaPageMovie(props: any) {
  const [isCollectionModalOpen, setIsCollectionModalOpen] = useState(false);
  const [isStreamModalOpen, setIsStreamModalOpen] = useState(false);
  const [isSelectStreamModalOpen, setIsSelectStreamModalOpen] = useState(false);
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false);
  const [isHistoryModalOpen, setIsHistoryModalOpen] = useState(false);
  const [isCreateHistoryModalOpen, setisCreateHistoryModalOpen] =
    useState(false);
  const [videoKey, setVideoKey] = useState("");
  const [streams, setStreams] = useState<any>(null);
  const [mainStream, setMainStream] = useState<any>(null);
  const [isStreamButtonLoading, setIsStreamButtonLoading] = useState(false);
  const [isStreamSelectButtonLoading, setIsStreamSelectButtonLoading] =
    useState(false);
  const [isPosterLoaded, setIsPosterLoaded] = useState(false);

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
  var releaseYear = "";
  try {
    releaseYear = props.data.release_date.slice(0, 4);
  } catch {}
  var genres = props.data.genres
    .map((item: any) => {
      return item.name;
    })
    .join(", ");
  var runtime = "";
  var creators = "";
  var isComingSoon = false;
  if (props.data.release_date) {
    isComingSoon = new Date(props.data.release_date) > new Date();
  }
  // get creators (directors)
  try {
    const lf = new Intl.ListFormat("en");
    creators = lf.format(
      props.data.credits.crew
        .filter((item: any) => item.job === "Director")
        .map((item: any) => {
          return item.name;
        })
    );
  } catch {}
  if (props.data.runtime > 0) {
    if (props.data.runtime >= 60) {
      runtime =
        Math.floor(props.data.runtime / 60) +
        "h " +
        (props.data.runtime % 60) +
        "m";
    } else {
      runtime = props.data.runtime + "m";
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
  // mode is either "direct" or "select"
  // direct plays the stream directly, select opens the stream selection modal
  const handleStreamButtonClick = (mode: string) => {
    if (mode === "direct") {
      setIsStreamButtonLoading(true);
    } else if (mode === "select") {
      setIsStreamSelectButtonLoading(true);
    }
    if (!streams) {
      axios
        .get(
          `/api/v1/movie/${props.data.media_source}-${props.data.source_id}/providers`
        )
        .then((res) => {
          setStreams(res.data);
          let numStreams = res.data.data.providers[0].streams.length;
          if (numStreams > 0) {
            setMainStream(res.data.data.providers[0].streams[0]);
          } else {
            toast.error("No streams found");
          }
          // open stream select modal if only few streams
          if (numStreams > 5 && mode === "direct") {
            setIsStreamModalOpen(true);
          } else {
            setIsSelectStreamModalOpen(true);
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
    } else if (streams.data.providers[0].streams.length > 0) {
      if (mode === "direct") {
        setIsStreamModalOpen(true);
        setIsStreamButtonLoading(false);
      } else if (mode === "select") {
        setIsSelectStreamModalOpen(true);
        setIsStreamSelectButtonLoading(false);
      }
    } else {
      toast.error("No Streams found");
    }
  };
  const handleVideoButtonClick = (key: string) => {
    setIsVideoModalOpen(true);
    setVideoKey(key);
  };
  if (props.data.media_title) {
    var yearString = props.data.release_date
      ? `(${props.data.release_date.slice(0, 4)})`
      : "";
    document.title = props.data.media_title + " " + yearString + " - Hound";
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
                {props.data.status === "Released" ? (
                  "Released"
                ) : (
                  <>
                    {props.data.release_date
                      ? "Releases " +
                        convertDateToReadable(props.data.release_date)
                      : ""}
                  </>
                )}
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
                      "▶ Play Movie"
                    )
                  }
                  autoClose="outside"
                  className="stream-play-button"
                  onClick={() => {
                    handleStreamButtonClick("direct");
                  }}
                >
                  <Dropdown.Item
                    eventKey="1"
                    onClick={() => {
                      handleStreamButtonClick("select");
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
                      Mark as Watched
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
                {/* <BootstrapTooltip
                  title={
                    <span className="media-page-tv-header-button-tooltip-title">
                      Track Show
                    </span>
                  }
                  PopperProps={offsetFix}
                >
                  <IconButton>
                    <BookmarkIcon id="media-page-tv-header-track-button" />
                  </IconButton>
                </BootstrapTooltip> */}
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
      <VideoModal
        onClose={() => {
          setIsVideoModalOpen(false);
        }}
        open={isVideoModalOpen}
        videoKey={videoKey}
      />
      <HistoryModal
        onClose={() => {
          setIsHistoryModalOpen(false);
        }}
        open={isHistoryModalOpen}
        data={props.data}
      />
      <CreateHistoryModal
        onClose={() => {
          setisCreateHistoryModalOpen(false);
        }}
        open={isCreateHistoryModalOpen}
        type={"movie"}
      />
    </>
  );
}

export default MediaPageMovie;
