import {
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
} from "@mui/material";
import "./Admin.css";
import Downloads from "./Downloads";
import { useState } from "react";
import ProviderProfiles from "./ProviderProfiles";

export default function Admin(props: any) {
  const [activeTab, setActiveTab] = useState(0);
  return (
    <>
      <div className="admin-main-container">
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
            <div className="admin-header">
              <h2>Admin Panel</h2>
            </div>
            <List>
              {["Downloads", "Users", "Provider Profiles"].map(
                (text, index) => (
                  <ListItem key={text} disablePadding>
                    <ListItemButton onClick={() => setActiveTab(index)}>
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
          <div className="admin-content">
            {activeTab === 0 && <Downloads />}
            {activeTab === 1 && <></>}
            {activeTab === 2 && <ProviderProfiles />}
          </div>
        </div>
      </div>
    </>
  );
}
