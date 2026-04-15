import {
  Card,
  CardContent,
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
import UserList from "./Users";
import { useServerInfo } from "../../api/hooks/general";

export default function Admin(props: any) {
  const [activeTab, setActiveTab] = useState(0);
  const { data: serverInfo, isLoading: isServerInfoLoading } = useServerInfo();
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
            <div className="p-2">
              <Card variant="outlined">
                <div className="p-3">
                  <p>Version: {serverInfo?.version}</p>
                  <p>Server ID: {serverInfo?.server_id}</p>
                </div>
              </Card>
            </div>
          </Drawer>
          <div className="d-flex admin-content">
            {activeTab === 0 && <Downloads />}
            {activeTab === 1 && <UserList />}
            {activeTab === 2 && <ProviderProfiles />}
          </div>
        </div>
      </div>
    </>
  );
}
