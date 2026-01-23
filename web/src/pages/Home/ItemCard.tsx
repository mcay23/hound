import { ImageListItem, Skeleton, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";
import "./ItemCard.css";
import PlayCircleFilledIcon from "@mui/icons-material/PlayCircleFilled";
import CommentCard from "../Comments/CommentCard";
import WatchTile from "../ContinueWatching/WatchTile";
import { useState } from "react";

const maxTitleLength = 30;

function ItemCard(props: {
  item: any;
  itemType:
    | "collectionImageList"
    | "poster"
    | "cast"
    | "video"
    | "seasons"
    | "search"
    | "image"
    | "comment"
    | "watch_tile";
  showTitle: any;
  itemOnClick: any;
}) {
  const [loaded, setLoaded] = useState(false);
  function itemTypePoster() {
    let mediaPagePath = `/${mediaType}/${props.item.media_source}-${props.item.source_id}`;
    if (!props.item.thumbnail_url) {
      return (
        <a href={mediaPagePath} className="itemcard-img-poster-container">
          <div
            className={
              "rounded w-100 h-100 itemcard-img-poster item-card-no-thumbnail border border-primary"
            }
          >
            {props.item.media_title + releaseYearText}
          </div>
        </a>
      );
    }
    return (
      <a href={mediaPagePath} className="itemcard-img-poster-container">
        {!loaded && (
          <Skeleton
            variant="rounded"
            className="rounded w-100 h-100"
            animation="wave"
          />
        )}
        <img
          className="rounded itemcard-img-poster"
          src={props.item.thumbnail_url}
          alt={props.item.media_title}
          loading="lazy"
          onLoad={() => setLoaded(true)}
          style={{
            opacity: loaded ? 1 : 0,
            transition: "opacity 0.5s ease",
          }}
        />
      </a>
    );
  }
  function itemTypeCast() {
    var primaryCaption = props.item.credits.name;
    var secondaryCaption = props.item.credits.character;
    return (
      <figure>
        {props.item.thumbnail_url ? (
          <>
            <div className="itemcard-img-cast-container">
              {!loaded && (
                <Skeleton
                  variant="rounded"
                  className="rounded w-100 h-100"
                  animation="wave"
                />
              )}
              <img
                className="rounded itemcard-img-cast"
                src={props.item.thumbnail_url}
                alt={props.item.media_title}
                loading="lazy"
                onLoad={() => setLoaded(true)}
                style={{
                  opacity: loaded ? 1 : 0,
                  transition: "opacity 0.5s ease",
                }}
              />
            </div>
          </>
        ) : (
          <div className="itemcard-img-cast-container">
            <div className="rounded itemcard-img-cast" />
          </div>
        )}
        <figcaption className="itemcard-cast-item-caption">
          <div className="itemcard-item-caption-primary">{primaryCaption}</div>
          <div className="itemcard-item-caption-secondary">
            {secondaryCaption}
          </div>
        </figcaption>
      </figure>
    );
  }
  function itemTypeSeason() {
    let primaryCaption = props.item.name;
    let secondaryCaption = props.item.episode_count
      ? props.item.episode_count + " episodes"
      : "";
    return (
      <figure>
        <div className="itemcard-img-poster-container">
          {props.item.poster_url ? (
            <>
              {!loaded && (
                <Skeleton
                  variant="rounded"
                  className="rounded w-100 h-100"
                  animation="wave"
                />
              )}
              <img
                className={"rounded itemcard-img-poster"}
                src={props.item.poster_url}
                alt={props.item.media_title}
                onClick={() => {
                  props.itemOnClick(props.item.season_number);
                }}
                onLoad={() => setLoaded(true)}
                style={{
                  opacity: loaded ? 1 : 0,
                  transition: "opacity 0.5s ease",
                }}
                loading="lazy"
              />
            </>
          ) : (
            <div
              className={"rounded itemcard-img-poster item-card-no-thumbnail"}
              onClick={() => {
                props.itemOnClick(props.item.season_number);
              }}
            >
              {props.item.name}
            </div>
          )}
        </div>
        {primaryCaption || secondaryCaption ? (
          <figcaption className="itemcard-item-caption">
            <div className="itemcard-item-caption-primary text-center">
              {primaryCaption}
            </div>
            <div className="itemcard-item-caption-secondary text-center">
              {secondaryCaption}
            </div>
          </figcaption>
        ) : (
          ""
        )}
      </figure>
    );
  }
  function itemTypeCollectionImageList() {
    let mediaPagePath = `/${mediaType}/${props.item.media_source}-${props.item.source_id}`;
    return (
      <>
        {props.item.thumbnail_url ? (
          <ImageListItem key={props.item.source_id}>
            <img
              className="rounded itemcard-img-poster"
              src={props.item.thumbnail_url}
              alt={props.item.media_title}
              onClick={() => navigate(mediaPagePath)}
              loading="lazy"
            />
          </ImageListItem>
        ) : (
          <ImageListItem key={props.item.media_title}>
            <a
              href={mediaPagePath}
              className="itemcard-img-poster d-flex w-100 h-100 justify-content-center align-items-center text-center text-wrap bg-light rounded border border-dark"
            >
              <h3>
                {props.item.media_title ? (
                  <Typography variant="h5">
                    {props.item.media_title.length > maxTitleLength
                      ? props.item.media_title.substring(0, maxTitleLength) +
                        "..."
                      : props.item.media_title}
                  </Typography>
                ) : (
                  <div className="">"Invalid title"</div>
                )}
              </h3>
            </a>
          </ImageListItem>
        )}
      </>
    );
  }
  function itemTypeSearch() {
    let mediaPagePath = `/${mediaType}/${props.item.media_source}-${props.item.source_id}`;
    let gameAspectRatioClass =
      mediaType === "game" && "itemcard-img-poster-game-cover";
    return (
      <a href={mediaPagePath}>
        {props.item.poster_url ? (
          <div className="itemcard-img-poster-container">
            {!loaded && (
              <Skeleton
                variant="rounded"
                className="rounded w-100 h-100"
                animation="wave"
              />
            )}
            <img
              className={"rounded itemcard-img-poster"}
              src={props.item.poster_url}
              alt={props.item.media_title}
              loading="lazy"
              onLoad={() => setLoaded(true)}
              style={{
                opacity: loaded ? 1 : 0,
                transition: "opacity 0.5s ease",
              }}
            />
          </div>
        ) : (
          <div className="itemcard-img-poster-container">
            <div
              className={
                "rounded itemcard-img-poster item-card-no-thumbnail " +
                gameAspectRatioClass
              }
            >
              {props.item.media_title + releaseYearText}
            </div>
          </div>
        )}
      </a>
    );
  }
  function itemTypeVideo() {
    return (
      <div
        className="video-button-trigger"
        key={props.item.key}
        onClick={() => {
          props.itemOnClick(props.item.key);
        }}
        style={{
          backgroundImage: `url('https://img.youtube.com/vi/${props.item.key}/0.jpg')`,
        }}
        title={props.item.id}
      >
        <PlayCircleFilledIcon fontSize="inherit" />
      </div>
    );
  }
  function itemTypeImage() {
    return (
      <div
        className="video-button-trigger"
        key={props.item.image_url}
        onClick={() => {
          props.itemOnClick(props.item.image_url);
        }}
        style={{
          backgroundImage: `url('${props.item.image_url}')`,
        }}
      />
    );
  }
  function itemTypeComment() {
    return <CommentCard item={props.item} />;
  }
  function itemTypeWatchTile() {
    return (
      <WatchTile item={props.item} loaded={loaded} setLoaded={setLoaded} />
    );
  }
  // get release years for use if thumbnail is not available - eg. Attack on Titan (2013)
  var mediaType = props.item.media_type;
  var releaseYearText = "";
  if (mediaType === "tvshow") {
    mediaType = "tv";
    if (props.item.first_air_date) {
      releaseYearText = ` (${props.item.first_air_date.slice(0, 4)})`;
    }
  } else if (mediaType === "movie" || mediaType === "game") {
    if (props.item.release_date) {
      releaseYearText = ` (${props.item.release_date.slice(0, 4)})`;
    }
  }
  const navigate = useNavigate();
  switch (props.itemType) {
    case "poster":
      return itemTypePoster();
    case "cast":
      return itemTypeCast();
    case "seasons":
      return itemTypeSeason();
    case "collectionImageList":
      return itemTypeCollectionImageList();
    case "search":
      return itemTypeSearch();
    case "video":
      return itemTypeVideo();
    case "image":
      return itemTypeImage();
    case "comment":
      return itemTypeComment();
    case "watch_tile":
      return itemTypeWatchTile();
  }
}

export default ItemCard;
