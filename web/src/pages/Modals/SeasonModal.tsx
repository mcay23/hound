import {
  Dialog,
  styled,
  Tooltip,
  tooltipClasses,
  TooltipProps,
} from "@mui/material";
import axios from "axios";
import { useEffect, useState } from "react";
import "./SeasonModal.css";
import convertDateToReadable from "../../helpers/helpers";
import ChatIcon from "@mui/icons-material/Chat";
import VisibilityIcon from "@mui/icons-material/Visibility";
import {
  IconButton,
  // styled,
  // Tooltip,
  // tooltipClasses,
  // TooltipProps,
} from "@mui/material";

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
  const { onClose, open, sourceID, seasonNumber } = props;
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
  });
  const [isSeasonDataLoaded, setIsSeasonDataLoaded] = useState(false);
  var seasonOverviewPlaceholder = "No description available.";
  if (isSeasonDataLoaded) {
    seasonOverviewPlaceholder = `Season ${seasonData.season.season_number} of ${props.mediaTitle}`;
    if (seasonData.season.season_number === 0) {
      seasonOverviewPlaceholder = "Special Episodes";
    }
  }
  useEffect(() => {
    setIsSeasonDataLoaded(false);
    // check data is loaded
    if (seasonNumber >= 0) {
      axios
        .get(`/api/v1/tv/tmdb-${sourceID}/season/${seasonNumber}`)
        .then((res) => {
          setSeasonData(res.data);
          setIsSeasonDataLoaded(true);
        })
        .catch((err) => {
          if (err.response.status === 500) {
            alert("500");
          }
        });
    }
  }, [seasonNumber, sourceID]);
  // data is already loaded, useEffect not triggered (open and close same season modal)
  if (
    !isSeasonDataLoaded &&
    seasonData &&
    seasonData.season.season_number === seasonNumber
  ) {
    setIsSeasonDataLoaded(true);
  }
  return (
    <Dialog
      onClose={handleClose}
      open={open}
      className="season-modal-dialog"
      maxWidth={false}
    >
      {isSeasonDataLoaded ? (
        <>
          <div className="season-modal-info-container">
            {seasonData.season.poster_path ? (
              <img
                className="rounded season-modal-poster"
                src={seasonData.season.poster_path}
                alt={seasonData.season.name}
              />
            ) : (
              <div
                className={"rounded season-modal-poster item-card-no-thumbnail"}
              >
                {seasonData.season.name}
              </div>
            )}
            <div className="season-modal-info-inner">
              <div className="season-modal-info-title">
                {seasonData.season.name}
                {seasonData.season.air_date ? (
                  <>
                    <span className="media-item-separator">|</span>
                    <span className="season-modal-info-date">
                      {seasonData.season.air_date.slice(0, 4)}
                    </span>
                  </>
                ) : (
                  ""
                )}
              </div>
              <hr />
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
                    <IconButton>
                      <VisibilityIcon />
                    </IconButton>
                  </BootstrapTooltip>
                </span>
                <span className="season-modal-info-button">
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
                </span>
              </div>
            </div>
          </div>
          {seasonData.season.episodes.map((episode) => {
            return EpisodeCard(episode);
          })}
        </>
      ) : (
        ""
      )}
    </Dialog>
  );
}

function EpisodeCard(episode: any) {
  var episodeNumber =
    episode.season_number.toString() &&
    episode.episode_number.toString() &&
    `S${episode.season_number}E${episode.episode_number}`.replace(
      "S0E",
      "Special #"
    );
  return (
    <div className="episode-card-container" key={episode.id}>
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
        <div>8.5/10</div>
        <BootstrapTooltip
          title={
            <span className="media-page-tv-header-button-tooltip-title">
              Write Review
            </span>
          }
          PopperProps={offsetFix}
        >
          <IconButton onClick={undefined}>
            <ChatIcon />
          </IconButton>
        </BootstrapTooltip>
      </div>
    </div>
  );
}

export default SeasonModal;
