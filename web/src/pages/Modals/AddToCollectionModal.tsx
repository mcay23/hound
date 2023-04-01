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

function AddToCollectionModal(props: any) {
  const { onClose, selectedValue, open } = props;

  const handleClose = () => {
    onClose(selectedValue);
  };

  const handleListItemClick = (value: string) => {
    onClose(value);
  };

  const emails = [
    "My library",
    "top 10 movies of all time",
    "hey123",
    "favorite movies",
  ];

  return (
    <Dialog
      onClose={handleClose}
      open={open}
      className="add-to-collection-dialog"
    >
      <div className="add-to-collection-dialog-content">
        <div className="add-to-collection-dialog-header">Add To Collection</div>
        <Divider variant="middle">⸱</Divider>
        <List sx={{ pt: 0 }}>
          {emails.map((item) => (
            <ListItem disableGutters className="pt-0 pb-0" key={item}>
              <ListItemButton
                onClick={() => handleListItemClick(item)}
                key={item}
              >
                <ListItemText
                  className="add-to-collection-dialog-choice"
                  primary={item}
                />
              </ListItemButton>
            </ListItem>
          ))}
          <ListItem disableGutters className="pt-0 pb-0">
            <ListItemButton onClick={() => handleListItemClick("addAccount")}>
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
      </div>
    </Dialog>
  );
}

export default AddToCollectionModal;
