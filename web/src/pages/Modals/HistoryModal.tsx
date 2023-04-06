import { Dialog } from "@mui/material";
import "./HistoryModal.css";
import HistoryTable from "../Comments/HistoryTable";

function HistoryModal(props: any) {
  const { onClose, open } = props;
  const handleClose = () => {
    onClose();
  };
  return (
    <Dialog
      onClose={handleClose}
      open={open}
      disableScrollLock={false}
      className="video-dialog"
      maxWidth={false}
    >
      <HistoryTable data={props.data} />
    </Dialog>
  );
}

export default HistoryModal;
