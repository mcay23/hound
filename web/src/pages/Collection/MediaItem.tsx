import "./MediaItem.css";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
} from "@mui/material";
import toast, { Toaster } from "react-hot-toast";
import ClearIcon from "@mui/icons-material/Clear";
import { useState } from "react";
import axios from "axios";
import ItemCard from "../Home/ItemCard";

function MediaItem(props: any) {
  var mediaType = props.item.media_type;
  var mediaTypeReadable =
    mediaType.charAt(0).toUpperCase() + mediaType.slice(1);
  if (props.item.media_type === "tvshow") {
    mediaType = "tv";
    mediaTypeReadable = "TV Show";
  }
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const handleDeleteClickOpen = () => {
    setIsDeleteDialogOpen(true);
  };
  const handleDeleteDialogClose = () => {
    setIsDeleteDialogOpen(false);
  };

  const handleDeleteItem = () => {
    if (props.item) {
      var payload = {
        media_type: props.item.media_type,
        media_source: props.item.media_source,
        source_id: props.item.source_id,
      };
      axios
        .delete(`/api/v1/collection/${props.collectionID}`, { data: payload })
        .then((res) => {
          setIsDeleteDialogOpen(false);
          window.location.reload();
        })
        .catch((err) => {
          console.log(err);
          toast.error("Failed to remove item from collection");
          setIsDeleteDialogOpen(false);
        });
    }
  };
  return (
    <>
      <div className="media-item-container">
        <div>
          <ItemCard item={props.item} itemType={"poster"} />
        </div>
        <div className="media-item-main-container">
          <div className="media-item-title-container">
            <a
              href={`/${mediaType}/${props.item.media_source}-${props.item.source_id}`}
              className="a-no-style"
            >
              <span className="media-item-title">{props.item.media_title}</span>
            </a>
            {props.item.release_date ? (
              <>
                <span className="media-item-separator">|</span>
                <span className="media-item-date">
                  {props.item.release_date.slice(0, 4)}
                </span>
              </>
            ) : (
              ""
            )}
          </div>
          <div className="media-item-secondary">{mediaTypeReadable}</div>
          <div className="media-item-description">
            {props.item.overview
              ? props.item.overview
              : "No description available."}
          </div>
        </div>
        <div className="media-item-actions-container">
          {props.showDeleteButton ? (
            <IconButton onClick={handleDeleteClickOpen}>
              <ClearIcon />
            </IconButton>
          ) : (
            ""
          )}
        </div>
      </div>
      <Dialog
        open={isDeleteDialogOpen}
        onClose={handleDeleteDialogClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">{"Delete this item?"}</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            This action cannot be reversed.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteDialogClose}>Cancel</Button>
          <Button onClick={handleDeleteItem}>Delete</Button>
        </DialogActions>
      </Dialog>
      <Toaster
        toastOptions={{
          duration: 5000,
        }}
      />
    </>
  );
}

export default MediaItem;
