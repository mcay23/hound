import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
} from "@mui/material";
import ClearIcon from "@mui/icons-material/Clear";
import toast, { Toaster } from "react-hot-toast";
import "./CommentCard.css";
import { useState } from "react";
import axios from "axios";

function CommentCard(props: any) {
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const handleDeleteClickOpen = () => {
    setIsDeleteDialogOpen(true);
  };
  const handleDeleteDialogClose = () => {
    setIsDeleteDialogOpen(false);
  };
  const handleDeleteItem = () => {
    if (props.item) {
      axios
        .delete(`/api/v1/comments/${props.item.comment_id}`)
        .then((res) => {
          setIsDeleteDialogOpen(false);
          window.scrollTo(0, 0);
          window.location.reload();
        })
        .catch((err) => {
          console.log(err);
          toast.error("Failed to remove review");
          setIsDeleteDialogOpen(false);
        });
    }
  };
  return (
    <>
      <div className="comment-card-container">
        <div className="w-100">
          <div className="comment-card-title-section">
            <div className="review-score-icon">{props.item.score}</div>
            <div className="comment-card-title">{props.item.title}</div>
          </div>
          <div className="comment-card-author">
            {"by " +
              props.item.user_id +
              "     ⸱     " +
              new Date(props.item.updated_at).toLocaleDateString("en-US")}
          </div>
          <div className="comment-card-divider">
            <hr />
          </div>
          <div className="comment-card-content">{props.item.comment}</div>
        </div>
        <div className="comment-card-actions-container">
          {props.item.user_id === localStorage.getItem("username") ? (
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
        <DialogTitle id="alert-dialog-title">
          {"Delete this review?"}
        </DialogTitle>
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

export default CommentCard;
