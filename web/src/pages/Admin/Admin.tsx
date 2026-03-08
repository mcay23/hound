import {
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
} from "@mui/material";
import "./Admin.css";
import AdminDownloads from "./AdminDownloads";

function Admin(props: any) {
  return (
    <>
      <div className="settings-main-container">
        <div className="d-flex">
          <Drawer
            variant="permanent"
            sx={{
              zIndex: 1,
              width: 300,
              flexShrink: 0,
              "& .MuiDrawer-paper": {
                width: 300,
                position: "sticky",
                top: 100,
                height: "calc(100vh - 100px)",
              },
            }}
          >
            <div className="settings-header">
              <h2>Settings</h2>
            </div>
            <List>
              {["Downloads", "Starred", "Send email", "Drafts"].map(
                (text, index) => (
                  <ListItem key={text} disablePadding>
                    <ListItemButton>
                      {/* <ListItemIcon>
                  {index % 2 === 0 ? <InboxIcon /> : <MailIcon />}
                </ListItemIcon> */}
                      <ListItemText primary={text} />
                    </ListItemButton>
                  </ListItem>
                ),
              )}
            </List>
          </Drawer>
          <div className="settings-content">
            <AdminDownloads />
          </div>
        </div>
      </div>
    </>
  );
}

export default Admin;
