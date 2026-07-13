import "./LiveTV.css";
import {
  useLiveTVChannels,
  useLiveTVCategories,
} from "../../api/hooks/live_tv";
import LiveVideoPlayer from "../VideoPlayer/LiveVideoPlayer";
import { useEffect, useState } from "react";
import {
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
} from "@mui/material";
import EPGMenu from "./EPGGrid";

function LiveTV(props: any) {
  const [selectedIPTVProvider, setSelectedIPTVProvider] = useState<
    number | undefined
  >(2);
  const [selectedCategoryID, setSelectedCategoryID] = useState<
    number | undefined
  >(undefined);
  const [sourceURL, setSourceURL] = useState<string | undefined>(undefined);
  const { data: liveTVCategories } = useLiveTVCategories(2);

  useEffect(() => {
    if (!selectedCategoryID) {
      setSelectedCategoryID(liveTVCategories?.[0]?.category_id || null);
    }
  }, [liveTVCategories]);

  return (
    <>
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
            <h2>Categories</h2>
          </div>
          <List>
            {liveTVCategories?.map((category: any) => (
              <ListItem key={category?.category_id} disablePadding>
                <ListItemButton
                  onClick={() => setSelectedCategoryID(category?.category_id)}
                >
                  <ListItemText primary={category?.category_name} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Drawer>
        <div className="live-tv-content">
          <h2>Live TV</h2>
          <LiveVideoPlayer src={sourceURL || ""} />
          <hr className="mt-3 mb-4" />
          <EPGMenu
            iptvProfileID={selectedIPTVProvider}
            categoryID={selectedCategoryID}
          />
        </div>
      </div>
    </>
  );
}

export default LiveTV;
