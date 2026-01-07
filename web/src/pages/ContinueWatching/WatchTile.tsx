import { Skeleton } from "@mui/material";
import "./WatchTile.css";

export default function WatchTile(props: any) {
  // check if it's a resume or next_episode object
  let thumbnailURL = "";
  let primaryCaption = "";
  let secondaryCaption = "";
  let path = props.item?.media_type === "movie" ? "movie" : "tv";
  let href = `/${path}/${props.item?.media_source}-${props.item?.source_id}`;
  if (props.item?.watch_action_type === "resume") {
    let watch_progress = props.item?.watch_progress;
    thumbnailURL = watch_progress.thumbnail_url;
    primaryCaption = watch_progress.media_title;
    if (props.item?.media_type === "tvshow") {
      primaryCaption += ` - S${watch_progress.season_number}E${watch_progress.episode_number}`;
      secondaryCaption = watch_progress.episode_title;
    }
  } else {
    let next_episode = props.item?.next_episode;
    thumbnailURL = next_episode.thumbnail_url;
    primaryCaption = next_episode.media_title;
    if (props.item?.media_type === "tvshow") {
      primaryCaption += ` - S${next_episode.season_number}E${next_episode.episode_number}`;
      secondaryCaption = next_episode.episode_title;
    }
  }
  return (
    <figure>
      <a className="itemcard-watch-tile-container" href={href}>
        {!props.loaded && (
          <Skeleton
            variant="rounded"
            className="rounded w-100 h-100"
            animation="wave"
          />
        )}
        <img
          className="rounded itemcard-watch-tile"
          src={thumbnailURL}
          alt={props.item.media_title}
          loading="lazy"
          onLoad={() => props.setLoaded(true)}
          style={{
            opacity: props.loaded ? 1 : 0,
            transition: "opacity 0.5s ease",
          }}
        />
      </a>
      <figcaption className="watch-tile-caption">
        <div className="itemcard-item-caption-primary">{primaryCaption}</div>
        <div className="itemcard-item-caption-secondary">
          {secondaryCaption}
        </div>
      </figcaption>
    </figure>
  );
}
