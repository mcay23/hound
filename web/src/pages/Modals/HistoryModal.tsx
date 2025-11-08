import { Dialog } from "@mui/material";
import "./HistoryModal.css";
import HistoryTable from "../Comments/HistoryTable";

function HistoryModal(props: any) {
  const { onClose, open } = props;
  return (
    <Dialog
      onClose={onClose}
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
