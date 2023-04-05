import Masonry, { ResponsiveMasonry } from "react-responsive-masonry";
import toast, { Toaster } from "react-hot-toast";
import ItemCard from "../Home/ItemCard";
import "./Reviews.css";
import {
  Button,
  Dialog,
  DialogActions,
  FormControl,
  TextField,
} from "@mui/material";
import { useState } from "react";
import axios from "axios";
function Reviews(props: any) {
  const [createReviewData, setCreateReviewData] = useState({
    title: "",
    comment: "",
    score: 100,
    is_private: false,
    comment_type: "review",
  });
  const [isReviewDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const handleReviewDialogOpenClick = () => {
    setIsDeleteDialogOpen(true);
  };
  const handleReviewDialogClose = () => {
    setCreateReviewData({
      title: "",
      comment: "",
      score: 100,
      is_private: false,
      comment_type: "review",
    });
    setIsDeleteDialogOpen(false);
  };
  const handlePostReview = () => {
    if (createReviewData.title === "") {
      toast.error("Title required");
      return;
    }
    if (createReviewData.comment === "") {
      toast.error("Review comment required");
      return;
    }
    if (createReviewData.score > 100 || createReviewData.score < 0) {
      toast.error("Score has to be between 1-100");
      return;
    }
    axios
      .post(`/api/v1${window.location.pathname}/comments`, createReviewData)
      .then(() => {
        window.scrollTo(0, 0);
        window.location.reload();
        handleReviewDialogClose();
      })
      .catch((err) => {
        console.log(err);
        toast.error("Error posting review");
      });
  };
  if (!props.data) {
    return <></>;
  }
  var your_reviews: any[] = [];
  var other_reviews: any[] = [];
  props.data.map((item: any) => {
    // get reviews, sort user's reviews first
    if (item.comment_type === "review") {
      if (item.user_id === localStorage.getItem("username")) {
        your_reviews.push(item);
      } else {
        other_reviews.push(item);
      }
    }
    return null;
  });
  var all_reviews = your_reviews.concat(other_reviews);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.name === "score") {
      // to do add bound check
      setCreateReviewData({
        ...createReviewData,
        score: parseInt(event.target.value),
      });
      return;
    }
    setCreateReviewData({
      ...createReviewData,
      [event.target.name]: event.target.value,
    });
  };
  return (
    <>
      <div className="reviews-main-container">
        <div className="reviews-header-container">Reviews</div>
        <ResponsiveMasonry columnsCountBreakPoints={{ 350: 1, 750: 2, 900: 3 }}>
          <Masonry>
            {all_reviews.map((item: any) => (
              <ItemCard
                item={item}
                key={item.id ? item.id : item.source_id}
                showTitle={null}
                itemType={"comment"}
                itemOnClick={undefined}
              />
            ))}
          </Masonry>
        </ResponsiveMasonry>
        <Button
          variant="contained"
          className="reviews-add-review-button"
          onClick={handleReviewDialogOpenClick}
        >
          Write a Review
        </Button>
      </div>
      <Dialog
        open={isReviewDialogOpen}
        onClose={handleReviewDialogClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <div className="reviews-create-dialog-header">Post New Review</div>
        <div className="reviews-create-dialog-content">
          <FormControl fullWidth={true}>
            <TextField
              id="outlined-basic"
              className="mt-3"
              label="Title"
              variant="outlined"
              name="title"
              value={createReviewData.title}
              onChange={handleChange}
            />
            <TextField
              type="number"
              className="mt-3"
              label="Score / 100"
              name="score"
              inputProps={{ min: "0", max: "100", step: "1" }}
              value={createReviewData.score}
              onChange={handleChange}
            />
            <TextField
              id="outlined-multiline-static"
              className="mt-3"
              label="Review"
              name="comment"
              multiline
              rows={4}
              value={createReviewData.comment}
              onChange={handleChange}
            />
          </FormControl>
        </div>
        <DialogActions>
          <Button onClick={handleReviewDialogClose}>Cancel</Button>
          <Button onClick={handlePostReview}>Post</Button>
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

export default Reviews;
