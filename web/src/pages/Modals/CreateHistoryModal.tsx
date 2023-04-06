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
import { create } from "domain";

function CreateHistoryModal(props: any) {
  const { onClose, open } = props;
  const [createHistoryData, setCreateHistoryData] = useState({
    comment: "",
    is_private: true,
    comment_type: "history",
    tag_data: "",
    start_date: "",
    end_date: "",
  });
  const [date, setDate] = React.useState<Dayjs | null>(dayjs());
  const handleClose = () => {
    setCreateHistoryData({
      comment: "",
      is_private: true,
      comment_type: "history",
      tag_data: "",
      start_date: "",
      end_date: "",
    });
    setDate(dayjs());
    onClose();
  };
  if (createHistoryData.start_date === "" && date) {
    setCreateHistoryData({
      ...createHistoryData,
      start_date: date.toISOString(),
      end_date: date.toISOString(),
    });
  }
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setCreateHistoryData({
      ...createHistoryData,
      [event.target.name]: event.target.value,
    });
  };
  const createHistoryHandler = () => {
    if (createHistoryData.start_date === "") {
      alert("unset date (bug, should not be possible)");
      return;
    }
    var payload = { ...createHistoryData };
    if (props.type === "season") {
      payload = { ...createHistoryData, tag_data: `S${props.seasonNumber}` };
    }
    console.log(payload, props.seasonNumber);
    axios
      .post(`/api/v1${window.location.pathname}/comments`, payload)
      .then(() => {
        toast.success("Added To Watch History");
        handleClose();
      })
      .catch((err) => {
        console.log(err);
      });
  };
  console.log(createHistoryData);
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
                    start_date: newValue.toISOString(),
                    end_date: newValue.toISOString(),
                  });
                }
              }}
            />
            {props.type === "movie" ? (
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
            )}
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
