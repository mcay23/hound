import "./Library.css";
import { ImageList } from "@mui/material";
import ItemCard from "./ItemCard";

function Library(props: any) {
  let libraryCols = 6;
  if (props.breakpoint === "md") {
    libraryCols = 5;
  } else if (props.breakpoint === "sm") {
    libraryCols = 3;
  } else if (props.breakpoint === "xs") {
    libraryCols = 2;
  }
  return (
    <>
      <div className="library-container">
        <ImageList cols={libraryCols} gap={10}>
          {props.library.map(
            (item: {
              thumbnail_url: string | undefined;
              media_title: string | undefined;
            }) => (
              <ItemCard
                item={item}
                showTitle={props.showTitle}
                itemType="collectionImageList"
                itemOnClick={undefined}
              />
            )
          )}
        </ImageList>
      </div>
      <br />
    </>
  );
}

export default Library;
