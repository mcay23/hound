import { ImageListItem, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";
import "./ItemCard.css";

const maxTitleLength = 30;

function ItemCard(props: {
  item: any;
  itemType:
    | "collectionImageList"
    | "poster"
    | "cast"
    | "videos"
    | "seasons"
    | "search";
  showTitle: any;
  itemOnClick: any;
}) {
  var mediaType = "";
  var releaseYearText = "";
  if (props.item.media_type === "tvshow") {
    mediaType = "tv";
    if (props.item.first_air_date) {
      releaseYearText = ` (${props.item.first_air_date.slice(0, 4)})`;
    }
  } else if (props.item.media_type === "movie") {
    mediaType = "movie";
    if (props.item.release_date) {
      releaseYearText = ` (${props.item.release_date.slice(0, 4)})`;
    }
  }
  const navigate = useNavigate();
  var mediaPagePath = "";
  if (props.itemType === "poster" || props.itemType === "collectionImageList") {
    mediaPagePath =
      "/" +
      mediaType +
      "/" +
      props.item.media_source +
      "-" +
      props.item.source_id;
  }
  if (props.itemType === "search") {
    mediaPagePath = `/${mediaType}/tmdb-` + props.item.source_id;
  }
  var thumbnailURL = props.item.thumbnail_url
    ? props.item.thumbnail_url
    : props.item.poster_url;
  // add caption for cast and season views
  var primaryCaption = "";
  var secondaryCaption = "";
  var centerCaptionStyle = {};
  if (props.itemType === "cast") {
    primaryCaption = props.item.credits.name;
    secondaryCaption = props.item.credits.character;
  } else if (props.itemType === "seasons") {
    primaryCaption = props.item.name;
    secondaryCaption = props.item.episode_count
      ? props.item.episode_count + " episodes"
      : "";
    centerCaptionStyle = {
      textAlign: "center",
    };
  }
  if (thumbnailURL) {
    if (props.itemType === "collectionImageList") {
      return (
        <ImageListItem
          key={
            mediaType +
            "-" +
            props.item.media_source +
            "-" +
            props.item.source_id
          }
        >
          <img
            className="rounded itemcard-item itemcard-img-poster"
            src={props.item.thumbnail_url}
            alt={props.item.media_title}
            onClick={() => navigate(mediaPagePath)}
            loading="lazy"
          />
        </ImageListItem>
      );
    }
    return (
      <figure>
        {props.itemType === "cast" ? (
          <img
            className={"rounded itemcard-img-cast"}
            src={thumbnailURL}
            alt={props.item.media_title}
            loading="lazy"
          />
        ) : (
          <>
            {props.itemType === "seasons" ? (
              <img
                className={"rounded itemcard-img-poster"}
                src={thumbnailURL}
                alt={props.item.media_title}
                onClick={() => {
                  props.itemOnClick(props.item.season_number);
                }}
                loading="lazy"
              />
            ) : (
              <a href={mediaPagePath}>
                <img
                  className={"rounded itemcard-img-poster"}
                  src={thumbnailURL}
                  alt={props.item.media_title}
                  loading="lazy"
                />
              </a>
            )}
          </>
        )}
        {primaryCaption || secondaryCaption ? (
          <figcaption className="itemcard-item-caption">
            <div
              className="itemcard-item-caption-primary"
              style={centerCaptionStyle}
            >
              {primaryCaption}
            </div>
            <div
              className="itemcard-item-caption-secondary"
              style={centerCaptionStyle}
            >
              {secondaryCaption}
            </div>
          </figcaption>
        ) : (
          ""
        )}
      </figure>
    );
  } else {
    if (props.itemType === "cast") {
      return (
        <>
          <img
            className={
              "rounded itemcard-img-cast itemcard-img-cast-no-thumbnail"
            }
            src={thumbnailURL}
            alt={props.item.media_title}
            loading="lazy"
          />
          {primaryCaption || secondaryCaption ? (
            <figcaption className="itemcard-item-caption">
              <div
                className="itemcard-item-caption-primary"
                style={centerCaptionStyle}
              >
                {primaryCaption}
              </div>
              <div
                className="itemcard-item-caption-secondary"
                style={centerCaptionStyle}
              >
                {secondaryCaption}
              </div>
            </figcaption>
          ) : (
            ""
          )}
        </>
      );
    }
    return props.itemType === "collectionImageList" ? (
      <ImageListItem key={props.item.media_title}>
        <a
          href={mediaPagePath}
          className="itemcard-item itemcard-img-poster d-flex w-100 h-100 justify-content-center align-items-center text-center text-wrap bg-light rounded border border-dark"
        >
          <h3>
            {props.item.media_title ? (
              <Typography variant="h5">
                {props.item.media_title.length > maxTitleLength
                  ? props.item.media_title.substring(0, maxTitleLength) + "..."
                  : props.item.media_title}
              </Typography>
            ) : (
              <div className="">"Invalid title"</div>
            )}
          </h3>
        </a>
      </ImageListItem>
    ) : (
      <>
        {mediaPagePath ? (
          <a href={mediaPagePath}>
            <div
              className={"rounded itemcard-img-poster item-card-no-thumbnail"}
            >
              {props.item.media_title + releaseYearText}
            </div>
          </a>
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
        {primaryCaption || secondaryCaption ? (
          <figcaption className="itemcard-item-caption">
            <div
              className="itemcard-item-caption-primary"
              style={centerCaptionStyle}
            >
              {primaryCaption}
            </div>
            <div
              className="itemcard-item-caption-secondary"
              style={centerCaptionStyle}
            >
              {secondaryCaption}
            </div>
          </figcaption>
        ) : (
          ""
        )}
      </>
    );
  }
}

export default ItemCard;
