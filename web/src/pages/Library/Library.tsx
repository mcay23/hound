import "./Library.css";
import Topnav from "../Topnav";
import axios from "axios";
import { useEffect, useState } from "react";
import CollectionCard from "./CollectionCover";
import HorizontalSection from "../Home/HorizontalSection";
import { LinearProgress } from "@mui/material";

function Library(props: any) {
  const [collections, setCollections] = useState([]);
  const [primaryCollection, setPrimaryCollection] = useState<any[]>([]);
  const [isCollectionsLoaded, setIsCollectionsLoaded] = useState(false);

  useEffect(() => {
    if (!isCollectionsLoaded) {
      const fetchData = async () => {
        var primaryCollectionID;
        await axios
          .get(`/api/v1/collection/all`)
          .then((res) => {
            setCollections(res.data);
            primaryCollectionID = res.data.find((item: any) => {
              return item.is_primary;
            }).collection_id;
          })
          .catch((err) => {
            if (err.response.status === 500) {
              alert("500");
            }
          });
        axios
          .get(`/api/v1/collection/${primaryCollectionID}?limit=20&offset=0`)
          .then((res) => {
            setPrimaryCollection(res.data.results);
            setIsCollectionsLoaded(true);
          })
          .catch((err) => {
            if (err.response.status === 500) {
              alert("500");
            }
          });
      };
      fetchData();
    }
  });
  return (
    <>
      <Topnav />
      {isCollectionsLoaded ? (
        <div className="library-main-container">
          <div className="library-top-section-container">
            <HorizontalSection
              items={primaryCollection}
              header="From Your Library"
              itemType="poster"
              itemOnClick={undefined}
            />
          </div>

          {
            <div className="library-collections-section">
              <div className="library-collections-header">Your Collections</div>
              <div className="library-collections-container">
                {collections.map((item) => (
                  <CollectionCard
                    data={item}
                    key={item["collection_id"]}
                    showCaption={true}
                  />
                ))}
              </div>
            </div>
          }
        </div>
      ) : (
        <LinearProgress />
      )}
    </>
  );
}

export default Library;
