import React, { useEffect, useState } from "react";
import "./Collection.css";
import Topnav from "../Topnav";
import axios from "axios";
import { useNavigate, useParams } from "react-router-dom";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  LinearProgress,
  Pagination,
} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import MediaItem from "./MediaItem";
import CollectionCover from "../Library/CollectionCover";
import toast, { Toaster } from "react-hot-toast";

function Collection(props: any) {
  const [collectionData, setCollectionData] = useState({
    results: [],
    collection: {
      collection_title: "",
      description: "",
      is_public: false,
      is_primary: true,
      owner_user_id: "",
    },
    total_records: 0,
  });
  const [isCollectionDataLoaded, setIsCollectionDataLoaded] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const itemsPerPage = 10;
  const navigate = useNavigate();
  const handlePageChange = (
    event: React.ChangeEvent<unknown>,
    value: number
  ) => {
    setPage(value);
  };
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const handleDeleteClickOpen = () => {
    setIsDeleteDialogOpen(true);
  };
  const handleDeleteDialogClose = () => {
    setIsDeleteDialogOpen(false);
  };
  const collectionID = useParams().id;
  var showDeleteButton = false;
  if (
    isCollectionDataLoaded &&
    collectionData.collection.owner_user_id === localStorage.getItem("username")
  ) {
    showDeleteButton = true;
  }
  const handleDeleteCollection = () => {
    axios
      .delete(`/api/v1/collection/delete/${collectionID}`)
      .then((res) => {
        setIsDeleteDialogOpen(false);
        navigate("/library");
      })
      .catch((err) => {
        console.log(err);
        toast.error("Failed to delete collection");
        setIsDeleteDialogOpen(false);
      });
  };
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
  if (isCollectionDataLoaded) {
    document.title = collectionData.collection.collection_title + " - Hound";
  }
  return (
    <>
      <Topnav />
      {isCollectionDataLoaded ? (
        <>
          <div className="collection-main-section">
            <div className="collection-items-list-container">
              <div className="collection-top-section">
                <div className="collection-cover-container">
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
                    <div className="collection-top-section-actions">
                      {showDeleteButton &&
                        !collectionData.collection.is_primary && (
                          <IconButton onClick={handleDeleteClickOpen}>
                            <DeleteIcon />
                          </IconButton>
                        )}
                    </div>
                  </div>
                </div>
              </div>
              {collectionData.results ? (
                collectionData.results.map((item) => (
                  <MediaItem
                    item={item}
                    collectionID={collectionID}
                    key={item["media_title"]}
                    showDeleteButton={showDeleteButton}
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
          <Dialog
            open={isDeleteDialogOpen}
            onClose={handleDeleteDialogClose}
            aria-labelledby="alert-dialog-title"
            aria-describedby="alert-dialog-description"
          >
            <DialogTitle id="alert-dialog-title">
              {"Delete this Collection?"}
            </DialogTitle>
            <DialogContent>
              <DialogContentText id="alert-dialog-description">
                This action cannot be reversed.
              </DialogContentText>
            </DialogContent>
            <DialogActions>
              <Button onClick={handleDeleteDialogClose}>Cancel</Button>
              <Button onClick={handleDeleteCollection}>Delete</Button>
            </DialogActions>
          </Dialog>
          <Toaster
            toastOptions={{
              duration: 5000,
            }}
          />
        </>
      ) : (
        <LinearProgress />
      )}
    </>
  );
}

export default Collection;
