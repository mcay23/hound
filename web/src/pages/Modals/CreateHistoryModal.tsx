import { Button, Dialog, DialogActions, FormControl } from "@mui/material";
import { DatePicker } from "@mui/x-date-pickers";
import toast from "react-hot-toast";
import dayjs, { Dayjs } from "dayjs";
import React, { useState } from "react";
import {
  useAddMovieWatchActivityMutation,
  useAddTVWatchActivityMutation,
} from "../../api/hooks/watchHistory";

interface CreateHistoryModalProps {
  onClose: () => void;
  open: boolean;
  type: "movie" | "season" | "episode";
  mediaSource: string;
  sourceID: string;
  episodeIDs?: number[];
}

function CreateHistoryModal({
  onClose,
  open,
  type,
  mediaSource,
  sourceID,
  episodeIDs,
}: CreateHistoryModalProps) {
  const [date, setDate] = useState<Dayjs | null>(dayjs());

  const addTVWatchActivityMutation = useAddTVWatchActivityMutation();
  const addMovieWatchActivityMutation = useAddMovieWatchActivityMutation();

  const handleClose = () => {
    setDate(dayjs());
    onClose();
  };

  const createHistoryHandler = async () => {
    if (!date) {
      toast.error("Please select a date");
      return;
    }
    const watchedAt = date.toISOString();
    try {
      if (type === "season" || type === "episode") {
        const cleanNumbers = episodeIDs?.map(Number).filter((n) => !isNaN(n));
        await addTVWatchActivityMutation.mutateAsync({
          mediaSource,
          sourceID,
          episodeIDs: cleanNumbers || [],
          watchedAt,
        });
      } else if (type === "movie") {
        await addMovieWatchActivityMutation.mutateAsync({
          mediaSource,
          sourceID,
          watchedAt,
        });
      }
      toast.success("Added to watch history");
      handleClose();
    } catch (err) {
      console.error(err);
      toast.error("Failed to add to watch history");
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      aria-labelledby="alert-dialog-title"
      aria-describedby="alert-dialog-description"
    >
      <div className="reviews-create-dialog-header">
        {type === "season" ? "Mark Season as Watched" : "Add Watch History"}
      </div>
      <div className="reviews-create-dialog-content">
        <FormControl fullWidth>
          <DatePicker
            className="mt-2"
            value={date}
            defaultValue={dayjs()}
            disableFuture
            onChange={(newValue) => setDate(newValue)}
          />
        </FormControl>
      </div>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button
          onClick={createHistoryHandler}
          disabled={
            addTVWatchActivityMutation.isPending ||
            addMovieWatchActivityMutation.isPending
          }
        >
          OK
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default CreateHistoryModal;
