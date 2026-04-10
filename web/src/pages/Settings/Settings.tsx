import {
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
} from "@mui/material";
import "./Settings.css";
import { useState } from "react";
import ApiKeys from "./ApiKeys";
import GeneralSettings from "./GeneralSettings";

export default function Settings(props: any) {
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
                position: "fixed",
                top: 100,
                left: 30,
                height: "calc(100vh - 100px)",
              },
            }}
          >
            <div>
              <h2>User Settings</h2>
            </div>
            <List>
              {["General Settings", "API Keys"].map((text, index) => (
                <ListItem key={text} disablePadding>
                  <ListItemButton onClick={() => setActiveTab(index)}>
                    {/* <ListItemIcon>
                  {index % 2 === 0 ? <InboxIcon /> : <MailIcon />}
                </ListItemIcon> */}
                    <ListItemText primary={text} />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          </Drawer>
          <div className="d-flex settings-content">
            {activeTab === 0 && <GeneralSettings />}
            {activeTab === 1 && <ApiKeys />}
          </div>
        </div>
      </div>
    </>
  );
}
