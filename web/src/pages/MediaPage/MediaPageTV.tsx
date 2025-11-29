import "./MediaPage.css";
import PlaylistAddIcon from "@mui/icons-material/PlaylistAdd";
import HistoryIcon from "@mui/icons-material/History";
import {
  IconButton,
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
import axios from "axios";

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
  const [videoKey, setVideoKey] = useState("");
  const [seasonModal, setSeasonModal] = useState(-1);
  const [isSeasonModalOpen, setIsSeasonModalOpen] = useState(false);
  const [isHistoryModalOpen, setIsHistoryModalOpen] = useState(false);

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
  if (props.data.media_title) {
    var yearString = props.data.first_air_date
      ? `(${props.data.first_air_date.slice(0, 4)})`
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
              {props.data.poster_url ? (
                <img
                  className="media-page-tv-poster"
                  src={props.data.poster_url}
                  alt={props.data.media_title}
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
      />
      <HistoryModal
        onClose={() => {
          setIsHistoryModalOpen(false);
        }}
        open={isHistoryModalOpen}
        data={props.data}
      />
    </>
  );
}

export default MediaPageTV;
