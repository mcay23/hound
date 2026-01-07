import { ScrollMenu } from "react-horizontal-scrolling-menu";
import "react-horizontal-scrolling-menu/dist/styles.css";
import "./HorizontalSection.css";
import { LeftArrow, RightArrow } from "./arrows";
import ItemCard from "./ItemCard";

function HorizontalSection(props: {
  items: any;
  header: string;
  itemType:
    | "poster"
    | "cast"
    | "video"
    | "seasons"
    | "search"
    | "image"
    | "comment"
    | "watch_tile";
  itemOnClick: any | undefined;
}) {
  if (!props.items || props.items.length === 0) {
    return <></>;
  }
  return (
    <>
      <div className="horizontal-section horizontal-section-menu">
        <div className="horizontal-section-header">
          {props.header}
          <span className="horizontal-section-header-separator">|</span>
        </div>
        <div className="horizontal-scroll-container">
          <ScrollMenu LeftArrow={LeftArrow} RightArrow={RightArrow}>
            {props.items.map((item: any) => (
              <div
                className="horizontal-section-item-container"
                key={item.id ? item.id : item.source_id}
              >
                <ItemCard
                  item={item}
                  showTitle={null}
                  itemType={props.itemType}
                  itemOnClick={props.itemOnClick}
                />
              </div>
            ))}
          </ScrollMenu>
        </div>
      </div>
    </>
  );
}
export default HorizontalSection;
