import "./Library.css";
import Topnav from "../Topnav";
import axios from "axios";
import { useEffect, useState } from "react";
import CollectionCard from "./CollectionCard";

function Library(props: any) {
  const [collections, setCollections] = useState([]);
  const [isCollectionsLoaded, setIsCollectionsLoaded] = useState(false);

  useEffect(() => {
    if (!isCollectionsLoaded) {
      axios
        .get(`/api/v1/collection/all`)
        .then((res) => {
          setCollections(res.data);
          setIsCollectionsLoaded(true);
          console.log(res.data);
        })
        .catch((err) => {
          if (err.response.status === 500) {
            alert("500");
          }
        });
    }
  });
  return (
    <>
      <Topnav />
      <div className="library-container">
        {isCollectionsLoaded ? (
          <>
            {collections.map((item) => (
              <CollectionCard data={item} />
            ))}
          </>
        ) : (
          ""
        )}
      </div>
      <br />
    </>
  );
}

export default Library;
