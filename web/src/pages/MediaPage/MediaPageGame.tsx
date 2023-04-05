import {
  IconButton,
  Tooltip,
  TooltipProps,
  styled,
  tooltipClasses,
} from "@mui/material";
import PlaylistAddIcon from "@mui/icons-material/PlaylistAdd";
import BookmarkIcon from "@mui/icons-material/Bookmark";
import React, { useState } from "react";
import convertDateToReadable from "../../helpers/helpers";
import HorizontalSection from "../Home/HorizontalSection";
import AddToCollectionModal from "../Modals/AddToCollectionModal";
import VideoModal from "../Modals/VideoModal";
import ImageModal from "../Modals/ImageModal";
import Reviews from "../Comments/Reviews";

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

function MediaPageGame(props: any) {
  // get backdropURL from artwork or screenshots if available
  var backdropURL = "";
  if (props.data.artworks) {
    backdropURL = props.data.artworks[0].image_url;
  } else if (props.data.screenshots) {
    backdropURL = props.data.screenshots[0].image_url;
  }
  const [isCollectionModalOpen, setIsCollectionModalOpen] = useState(false);
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false);
  const [videoKey, setVideoKey] = useState("");
  const [isImageModalOpen, setIsImageModalOpen] = useState(false);
  const [imageURL, setImageURL] = useState("");
  const handleVideoButtonClick = (key: string) => {
    setIsVideoModalOpen(true);
    setVideoKey(key);
  };
  const handleVideoButtonClose = () => {
    setIsVideoModalOpen(false);
  };
  const handleImageButtonClick = (key: string) => {
    setIsImageModalOpen(true);
    setImageURL(key);
  };
  const handleImageButtonClose = () => {
    setIsImageModalOpen(false);
  };
  // modal functions
  const handleAddToCollectionButtonClick = () => {
    setIsCollectionModalOpen(true);
  };
  const handleAddToCollectionClose = () => {
    setIsCollectionModalOpen(false);
  };
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
        backdropURL +
        ")",
      backgroundAttachment: "fixed",
      backgroundSize: "cover",
      // animation: "backgroundScroll 40s linear infinite",
    },
    opacityBackdrop: {
      // backgroundColor: "blue",
      backgroundImage:
        "linear-gradient(rgba(255, 255, 255, 0.94), rgba(255, 255, 255, 0.94)), url(" +
        backdropURL +
        ")",
      backgroundAttachment: "fixed",
      backgroundSize: "cover",
    },
  };
  var genres = "";
  if (props.data.genres) {
    genres = props.data.genres
      .map((item: any) => {
        return item.name;
      })
      .join(", ");
  }
  var platforms = "";
  if (props.data.platforms) {
    platforms = props.data.platforms
      .map((item: any) => {
        return item.name;
      })
      .join(", ");
  }
  const lf = new Intl.ListFormat("en");
  var credits = "";
  if (props.data.involved_companies) {
    credits = lf.format(
      props.data.involved_companies
        .filter((item: any) => item.developer || item.publisher)
        .map((item: any) => item.company.name)
    );
  }
  let images = [
    ...(props.data.artworks ?? []),
    ...(props.data.screenshots ?? []),
  ];
  return (
    <>
      <div
        className="media-page-tv-header"
        style={backdropURL ? styles.withBackdrop : styles.noBackdrop}
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
                  {props.data.release_date
                    ? "(" + props.data.release_date.slice(0, 4) + ")"
                    : ""}
                </span>
              </div>
              <div className="media-page-tv-header-genres">{platforms}</div>
              <div className="media-page-tv-header-overview">
                {props.data.summary
                  ? props.data.summary
                  : "No description available."}
              </div>
              <div className="media-page-tv-header-credits">
                {credits ? "by " + credits : ""}
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
                  <IconButton onClick={handleAddToCollectionButtonClick}>
                    <PlaylistAddIcon />
                  </IconButton>
                </BootstrapTooltip>
                <BootstrapTooltip
                  title={
                    <span className="media-page-tv-header-button-tooltip-title">
                      Track Game
                    </span>
                  }
                  PopperProps={offsetFix}
                >
                  <IconButton>
                    <BookmarkIcon id="media-page-tv-header-track-button" />
                  </IconButton>
                </BootstrapTooltip>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div className="media-page-tv-main" style={styles.opacityBackdrop}>
        {props.data.storyline ? (
          <div className="media-page-storyline">
            <div className="media-page-section-header">Storyline</div>
            {props.data.storyline}
          </div>
        ) : (
          ""
        )}
        <HorizontalSection
          items={props.data.videos}
          header={"Videos"}
          itemType="video"
          itemOnClick={handleVideoButtonClick}
        />
        <HorizontalSection
          items={images}
          header={"Images"}
          itemType="image"
          itemOnClick={handleImageButtonClick}
        />
        <Reviews data={props.data.comments} />
      </div>
      <AddToCollectionModal
        onClose={handleAddToCollectionClose}
        open={isCollectionModalOpen}
        item={props.data}
      />
      <VideoModal
        onClose={handleVideoButtonClose}
        open={isVideoModalOpen}
        videoKey={videoKey}
      />
      <ImageModal
        onClose={handleImageButtonClose}
        open={isImageModalOpen}
        imageURL={imageURL}
      />
    </>
  );
}

export default MediaPageGame;
