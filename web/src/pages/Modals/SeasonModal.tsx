import {
  CircularProgress,
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
    },
  });
  const [isSeasonDataLoaded, setIsSeasonDataLoaded] = useState(false);
  useEffect(() => {
    setIsSeasonDataLoaded(false);
    // check data is loaded
    if (seasonNumber >= 0) {
      axios
        .get(`/api/v1/tv/tmdb-${sourceID}/season/${seasonNumber}`)
        .then((res) => {
          setSeasonData(res.data);
          setIsSeasonDataLoaded(true);
          console.log("seasondata", res.data);
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
      className="season-dialog"
      maxWidth={false}
    >
      {isSeasonDataLoaded ? (
        <>
          {seasonData.season.name}
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
