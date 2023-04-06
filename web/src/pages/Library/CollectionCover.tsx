import React from "react";
import "./CollectionCover.css";

function CollectionCover(props: any) {
  var backdropURL =
    "https://image.tmdb.org/t/p/w300/sHim6U0ANsbzxcmNRYuIubBVQaz.jpg";
  var styles = {
    withBackdrop: {
      backgroundImage:
        "linear-gradient(rgba(0, 0, 0, 0.6) 40%, rgba(0, 0, 0, 0.4) 70%), url(" +
        backdropURL +
        ")",
      backgroundSize: "cover",
      color: "yellow",
    },
  };
  return (
    <a className="a-no-style" href={`/collection/${props.data.collection_id}`}>
      <figure className="itemcard-item-figure-container">
        <div
          className={"rounded collection-card-cover"}
          // style={styles.withBackdrop}
        >
          <div className={"collection-card-cover-inner"}>
            {props.data.collection_title}
          </div>
        </div>
        {props.showCaption ? (
          <figcaption className="itemcard-item-caption itemcard-item-collection-caption ms-2">
            <div className="itemcard-item-caption-primary">
              {props.data.collection_title}
            </div>
            <div className="itemcard-item-caption-secondary">
              {"by " + props.data.owner_user_id}
            </div>
          </figcaption>
        ) : (
          ""
        )}
      </figure>
    </a>
  );
}

export default CollectionCover;
