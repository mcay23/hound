import {
  Dialog,
  Divider,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import "./AddToCollectionModal.css";
import { useEffect, useState } from "react";
import axios from "axios";

function AddToCollectionModal(props: any) {
  const { onClose, open, item } = props;

  const handleClose = () => {
    onClose();
  };

  const handleListItemClick = (collectionID: number) => {
    // add item to collection
    if (item && item["source_id"]) {
      var payload = {
        media_type: item["media_type"],
        media_source: item["media_source"],
        source_id: item["source_id"].toString(),
      };
      axios
        .post(`/api/v1/collection/${collectionID}`, payload)
        .then((res) => {
          console.log("add item success");
        })
        .catch((err) => {
          console.log("AXIOS ERROR: ", err);
        });
    }
    onClose();
  };

  const handleCreateNewCollection = () => {
    onClose();
  };

  const [data, setData] = useState([]);

  useEffect(() => {
    axios
      .get(`/api/v1/collection/all`)
      .then((res) => {
        console.log(res.data);
        setData(res.data);
      })
      .catch((err) => {
        if (err.response.status === 500) {
          alert("500");
        }
      });
  }, [props.open]);

  return (
    <Dialog
      onClose={handleClose}
      open={open}
      className="add-to-collection-dialog"
    >
      <div className="add-to-collection-dialog-content">
        <div className="add-to-collection-dialog-header">Add To Collection</div>
        <Divider variant="middle">⸱</Divider>
        {data ? (
          <List sx={{ pt: 0 }}>
            {data.map((item) => (
              <ListItem
                disableGutters
                className="pt-0 pb-0"
                key={item["collection_id"]}
              >
                <ListItemButton
                  onClick={() => handleListItemClick(item["collection_id"])}
                  key={item}
                >
                  <ListItemText
                    className="add-to-collection-dialog-choice"
                    primary={item["collection_title"]}
                  />
                </ListItemButton>
              </ListItem>
            ))}
            <ListItem disableGutters className="pt-0 pb-0">
              <ListItemButton onClick={() => handleCreateNewCollection()}>
                {/* <ListItemAvatar>
              <Avatar>
                <AddIcon />
              </Avatar>
            </ListItemAvatar> */}
                <AddIcon />
                <ListItemText
                  className="add-to-collection-dialog-button"
                  primary="New Collection"
                />
              </ListItemButton>
            </ListItem>
          </List>
        ) : (
          ""
        )}
      </div>
    </Dialog>
  );
}

export default AddToCollectionModal;
