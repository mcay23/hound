import "./LiveTV.css";
import { useLiveTVCategories } from "../../api/hooks/live_tv";
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
  >(1);
  const [selectedCategoryID, setSelectedCategoryID] = useState<
    number | undefined
  >(undefined);
  const [sourceURL, setSourceURL] = useState<string | undefined>(undefined);
  const {
    data: liveTVCategories,
    isLoading: liveTVCategoriesLoading,
    error: liveTVCategoriesError,
  } = useLiveTVCategories(selectedIPTVProvider);

  useEffect(() => {
    if (!selectedCategoryID) {
      setSelectedCategoryID(liveTVCategories?.[0]?.category_id || null);
    }
  }, [liveTVCategories]);

  return (
    <>
      <div className="d-flex" style={{ backgroundColor: "#111319" }}>
        <Drawer
          variant="permanent"
          sx={{
            zIndex: 1,
            width: 300,
            flexShrink: 0,
            backgroundColor: "#111319",
            "& .MuiDrawer-paper": {
              scrollbarColor: "#4f5668 #111319",
              "&::-webkit-scrollbar-track": {
                background: "#111319",
              },
              "&::-webkit-scrollbar-thumb": {
                backgroundColor: "#4f5668",
                borderRadius: 8,
              },
              "&::-webkit-scrollbar-thumb:hover": {
                backgroundColor: "#6b7280",
              },
              backgroundColor: "#111319",
              color: "white",
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
            {liveTVCategoriesLoading && <div className="mt-2">Loading...</div>}
            {liveTVCategoriesError && (
              <div>
                Error Fetching Categories: {liveTVCategoriesError.message}
              </div>
            )}
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
          <LiveVideoPlayer src={sourceURL || ""} />
          <hr className="mt-3 mb-4" />
          <EPGMenu
            iptvProviderID={selectedIPTVProvider}
            categoryID={selectedCategoryID}
            setSourceURL={setSourceURL}
            sourceURL={sourceURL}
          />
        </div>
      </div>
    </>
  );
}

export default LiveTV;
