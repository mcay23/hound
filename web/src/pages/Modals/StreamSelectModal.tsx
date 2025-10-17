import { Dialog } from "@mui/material";
import "./StreamSelectModal.css";
import "video.js/dist/video-js.css";
import { useEffect, useState } from "react";
import type { ColDef } from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";

function SelectStreamModal(props: any) {
  const { setOpen, open } = props;
  const handleClose = () => {
    setOpen(false);
  };
  const [rowData, setRowData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [colDefs] = useState<ColDef[]>([
    { field: "mission" },
    { field: "company" },
    { field: "location" },
    { field: "date" },
    { field: "price" },
    { field: "successful" },
    { field: "rocket" },
  ]);
  // Fetch data & update rowData state
  useEffect(() => {
    setLoading(true);
    fetch("https://www.ag-grid.com/example-assets/space-mission-data.json") // Fetch data from server
      .then((result) => result.json()) // Convert to JSON
      .then((rowData) => setRowData(rowData)) // Update state of `rowData`
      .then(() => setLoading(false));
  }, []);
  return (
    <Dialog
      onClose={handleClose}
      open={open}
      disableScrollLock={false}
      fullScreen
    >
      <div style={{ width: "100%", height: "100%" }}>
        <h1>Streams</h1>
        <AgGridReact rowData={rowData} loading={loading} columnDefs={colDefs} />
      </div>
    </Dialog>
  );
}

export default SelectStreamModal;
