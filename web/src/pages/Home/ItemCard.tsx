import { ImageListItem, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";
import "./ItemCard.css";
import PlayCircleFilledIcon from "@mui/icons-material/PlayCircleFilled";

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
    | "image";
  showTitle: any;
  itemOnClick: any;
}) {
  function itemTypePoster() {
    let mediaPagePath = `/${mediaType}/${props.item.media_source}-${props.item.source_id}`;
    return (
      <a href={mediaPagePath}>
        <img
          className={"rounded itemcard-img-poster"}
          src={props.item.thumbnail_url}
          alt={props.item.media_title}
          loading="lazy"
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
          <img
            className={"rounded itemcard-img-cast"}
            src={props.item.thumbnail_url}
            alt={props.item.media_title}
            loading="lazy"
          />
        ) : (
          <div className={"rounded itemcard-img-cast"} />
        )}
        <figcaption className="itemcard-item-caption">
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
        {props.item.poster_url ? (
          <img
            className={"rounded itemcard-img-poster"}
            src={props.item.poster_url}
            alt={props.item.media_title}
            onClick={() => {
              props.itemOnClick(props.item.season_number);
            }}
            loading="lazy"
          />
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
              className="rounded itemcard-item itemcard-img-poster"
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
              className="itemcard-item itemcard-img-poster d-flex w-100 h-100 justify-content-center align-items-center text-center text-wrap bg-light rounded border border-dark"
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
          <img
            className={"rounded itemcard-img-poster " + gameAspectRatioClass}
            src={props.item.poster_url}
            alt={props.item.media_title}
            loading="lazy"
          />
        ) : (
          <div
            className={
              "rounded itemcard-img-poster item-card-no-thumbnail " +
              gameAspectRatioClass
            }
          >
            {props.item.media_title + releaseYearText}
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
      ></div>
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
  } else if (mediaType === "movie") {
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
  }
}

export default ItemCard;
