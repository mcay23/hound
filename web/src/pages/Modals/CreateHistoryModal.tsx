import {
  Button,
  Dialog,
  DialogActions,
  FormControl,
  TextField,
} from "@mui/material";
import { DatePicker } from "@mui/x-date-pickers";
import toast, { Toaster } from "react-hot-toast";
import axios from "axios";
import dayjs, { Dayjs } from "dayjs";
import React, { useState } from "react";

function CreateHistoryModal(props: any) {
  const { onClose, open } = props;
  const [createHistoryData, setCreateHistoryData] = useState({
    is_private: true,
    action_type: "watch",
    watched_at: "",
  });
  const [date, setDate] = React.useState<Dayjs | null>(dayjs());
  const handleClose = () => {
    setCreateHistoryData({
      is_private: true,
      action_type: "watch",
      watched_at: "",
    });
    setDate(dayjs());
    onClose();
  };
  if (createHistoryData.watched_at === "" && date) {
    setCreateHistoryData({
      ...createHistoryData,
      watched_at: date.toISOString(),
    });
  }
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setCreateHistoryData({
      ...createHistoryData,
      [event.target.name]: event.target.value,
    });
  };
  const createHistoryHandler = () => {
    if (createHistoryData.watched_at === "") {
      alert(
        "unset date (bug, should not be possible, please report in github)"
      );
      return;
    }
    var payload = { ...createHistoryData };
    // full season insertion
    if (props.type === "season") {
      payload = { ...createHistoryData };
    }
    axios
      .post(`/api/v1${window.location.pathname}/history`, payload)
      .then(() => {
        toast.success("Added To Watch History");
        handleClose();
      })
      .catch((err) => {
        console.log(err);
      });
  };
  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <div className="reviews-create-dialog-header">
          {props.type === "season"
            ? "Mark Season as Watched"
            : "Add Watch History"}
        </div>
        <div className="reviews-create-dialog-content">
          <FormControl fullWidth={true}>
            <DatePicker
              className="mt-2"
              value={date}
              defaultValue={dayjs("2022-04-17")}
              disableFuture
              onChange={(newValue) => {
                setDate(newValue);
                if (newValue) {
                  setCreateHistoryData({
                    ...createHistoryData,
                    watched_at: newValue.toISOString(),
                  });
                }
              }}
            />
            {/* {props.type === "movie" ? (
              <TextField
                id="outlined-multiline-static"
                className="mt-3"
                label="Notes"
                name="comment"
                multiline
                rows={4}
                value={createHistoryData.comment}
                onChange={handleChange}
              />
            ) : (
              ""
            )} */}
          </FormControl>
        </div>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button onClick={createHistoryHandler}>Ok</Button>
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

export default CreateHistoryModal;
