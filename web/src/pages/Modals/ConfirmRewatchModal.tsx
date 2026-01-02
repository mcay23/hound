import { Dialog, Button } from "@mui/material";
import "./ConfirmRewatchModal.css";
import axios from "axios";
import toast from "react-hot-toast";

function ConfirmRewatchModal(props: any) {
  const { onClose, open, mediaSource, sourceID } = props;
  const handleConfirm = () => {
    axios
      .post(`/api/v1/tv/${mediaSource}-${sourceID}/history/rewatch`)
      .then((res) => {
        toast.success("Rewatch created successfully");
        onClose();
      })
      .catch((err) => {
        console.log(err);
        if (err.response.status === 400) {
          toast.error("Your current rewatch is already empty!");
        } else {
          toast.error("Error creating rewatch");
        }
      });
    onClose();
  };
  return (
    <Dialog onClose={onClose} open={open} disableScrollLock={false}>
      <div className="confirm-rewatch-modal-container">
        <h4>Confirm Rewatch</h4>
        <p>
          Are you sure you want to rewatch this show? This will archive your
          current progress.
        </p>
        <div className="confirm-rewatch-modal-buttons d-flex justify-content-end">
          <Button onClick={onClose}>Cancel</Button>
          <Button onClick={handleConfirm}>Confirm</Button>
        </div>
      </div>
    </Dialog>
  );
}

export default ConfirmRewatchModal;
