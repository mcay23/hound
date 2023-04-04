import React, { useEffect, useState } from "react";
import "./Collection.css";
import Topnav from "../Topnav";
import axios from "axios";
import { useParams } from "react-router-dom";
import { LinearProgress, Pagination } from "@mui/material";
import MediaItem from "./MediaItem";
import CollectionCover from "../Library/CollectionCover";

function Collection(props: any) {
  const [collectionData, setCollectionData] = useState({
    results: [],
    collection: {
      collection_title: "",
      description: "",
      is_public: false,
      owner_user_id: "",
    },
    total_records: 0,
  });
  const [isCollectionDataLoaded, setIsCollectionDataLoaded] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const itemsPerPage = 10;
  const handleReload = () => {
    window.location.reload();
  };
  const handlePageChange = (
    event: React.ChangeEvent<unknown>,
    value: number
  ) => {
    setPage(value);
  };
  const collectionID = useParams().id;
  useEffect(() => {
    axios
      .get(
        `/api/v1/collection/${collectionID}?limit=${itemsPerPage}&offset=${
          itemsPerPage * (page - 1)
        }`
      )
      .then((res) => {
        setCollectionData(res.data);
        setIsCollectionDataLoaded(true);
        setTotalPages(Math.ceil(res.data.total_records / itemsPerPage));
      })
      .catch((err) => {
        if (err.response.status === 500) {
          alert("500");
        }
      });
  }, [collectionID, page]);
  window.scrollTo(0, 0);
  return (
    <>
      <Topnav />
      {isCollectionDataLoaded ? (
        <>
          <div className="collection-main-section">
            <div className="collection-items-list-container">
              <div className="collection-info-section">
                <div className="collection-cover">
                  <CollectionCover
                    data={collectionData.collection}
                    key={collectionData.collection.collection_title}
                    showCaption={false}
                  />
                  <div className="collection-cover-main">
                    <div className="collection-cover-main-title">
                      {collectionData.collection.collection_title}
                    </div>
                    <div className="collection-cover-date">
                      {`by ${collectionData.collection.owner_user_id}`}
                    </div>
                    <hr />
                    <div className="collection-cover-main-description">
                      {collectionData.collection.description}
                    </div>
                  </div>
                </div>
              </div>
              {collectionData.results ? (
                collectionData.results.map((item) => (
                  <MediaItem
                    item={item}
                    collectionID={collectionID}
                    handleReload={handleReload}
                    key={item["media_title"]}
                  />
                ))
              ) : (
                <span className="collection-empty-message">
                  This collection is empty.
                </span>
              )}
            </div>
          </div>
          {collectionData.results ? (
            <div className="d-flex justify-content-center mb-4 mt-2">
              <div className="paginator-container shadow-lg">
                <Pagination
                  id="paginator-component"
                  defaultPage={1}
                  page={page}
                  onChange={handlePageChange}
                  count={totalPages}
                  size="large"
                />
              </div>
            </div>
          ) : (
            ""
          )}
        </>
      ) : (
        <LinearProgress />
      )}
    </>
  );
}

export default Collection;
