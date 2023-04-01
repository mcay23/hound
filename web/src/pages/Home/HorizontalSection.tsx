import { ScrollMenu } from "react-horizontal-scrolling-menu";
import PlayCircleFilledIcon from "@mui/icons-material/PlayCircleFilled";
import "react-horizontal-scrolling-menu/dist/styles.css";
import "./HorizontalSection.css";
import { LeftArrow, RightArrow } from "./arrows";
import ItemCard2 from "../Library/ItemCard";

function HorizontalSection(props: {
  items: any;
  header: string;
  itemType: "poster" | "cast" | "videos" | "seasons" | "search";
  itemOnClick: any | undefined;
}) {
  if (!props.items || props.items.length === 0) {
    return <></>;
  }
  return (
    <>
      <div className="horizontal-section horizontal-section-menu">
        <div className="horizontal-section-header">{props.header}</div>
        <div>
          <ScrollMenu LeftArrow={LeftArrow} RightArrow={RightArrow}>
            {props.items.map((item: any) => (
              <ItemCard2
                item={item}
                key={item.id}
                showTitle={null}
                itemType={props.itemType}
                itemOnClick={props.itemOnClick}
              />
            ))}
          </ScrollMenu>
        </div>
      </div>
    </>
  );
}
export default HorizontalSection;
