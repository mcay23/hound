import "./LiveTV.css";
import { useLiveTVCategories } from "../../api/hooks/live_tv";
import LiveVideoPlayer from "../VideoPlayer/LiveVideoPlayer";
import { useEffect, useMemo, useState } from "react";
import {
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
} from "@mui/material";
import EPGMenu, { SelectedChannel, EPGProgramme, pickText } from "./EPGGrid";

function findNowPlaying(epg: EPGProgramme[]): EPGProgramme | undefined {
  const now = new Date();
  return epg.find((prog) => {
    const start = new Date(prog.start_time);
    const stop = new Date(prog.stop_time);
    return now >= start && now < stop;
  });
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], {
    hour: "numeric",
    minute: "2-digit",
  });
}

function LiveTV(props: any) {
  const [selectedIPTVProvider, setSelectedIPTVProvider] = useState<
    number | undefined
  >(1);
  const [selectedCategoryID, setSelectedCategoryID] = useState<
    number | undefined
  >(undefined);
  const [selectedChannel, setSelectedChannel] = useState<
    SelectedChannel | undefined
  >(undefined);

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

  const nowPlaying = useMemo(
    () => (selectedChannel ? findNowPlaying(selectedChannel.epg) : undefined),
    [selectedChannel],
  );

  const sourceURL = selectedChannel?.channel.stream_url;

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
          <div className="d-flex flex-row">
            <div style={{ width: "70%" }}>
              <LiveVideoPlayer src={sourceURL || ""} />
            </div>
            <div
              style={{ width: "30%", marginTop: "1rem", marginLeft: "2rem" }}
            >
              <h1 className="text-white">
                {selectedChannel?.channel.name || "No Channel Selected"}
              </h1>
              {nowPlaying ? (
                <>
                  <h2 className="text-white">
                    Now Playing: {pickText(nowPlaying.titles)}
                  </h2>
                  <h3 className="text-white">
                    {pickText(nowPlaying.descriptions)}
                  </h3>
                  <h3 className="text-white">
                    {formatTime(nowPlaying.start_time)} -{" "}
                    {formatTime(nowPlaying.stop_time)}
                  </h3>
                </>
              ) : selectedChannel ? (
                <h2 className="text-white">No programme info available</h2>
              ) : null}
            </div>
          </div>
          <hr className="mt-3 mb-4" />
          <EPGMenu
            iptvProviderID={selectedIPTVProvider}
            categoryID={selectedCategoryID}
            selectedChannel={selectedChannel}
            setSelectedChannel={setSelectedChannel}
          />
        </div>
      </div>
    </>
  );
}

export default LiveTV;
