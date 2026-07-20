import "./LiveTV.css";
import { useIPTVProviders, useLiveTVCategories } from "../../api/hooks/live_tv";
import LiveVideoPlayer from "../VideoPlayer/LiveVideoPlayer";
import { useEffect, useMemo, useState } from "react";
import {
  Drawer,
  FormControl,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  MenuItem,
  Select,
} from "@mui/material";
import EPGMenu, { SelectedChannel, EPGProgramme, pickText } from "./EPGGrid";

function LiveTV(props: any) {
  const [selectedIPTVProvider, setSelectedIPTVProvider] = useState<
    number | undefined
  >(undefined);
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

  const {
    data: iptvProviders,
    isLoading: iptvProvidersLoading,
    error: iptvProvidersError,
  } = useIPTVProviders();

  useEffect(() => {
    if (!selectedCategoryID) {
      setSelectedCategoryID(liveTVCategories?.[0]?.category_id || null);
    }
  }, [liveTVCategories]);

  useEffect(() => {
    if (iptvProviders?.length > 0 && !selectedIPTVProvider) {
      setSelectedIPTVProvider(iptvProviders[0]?.iptv_provider_id);
    }
  }, [iptvProviders]);

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
            width: 250,
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
          {iptvProviders?.length > 0 && (
            <FormControl sx={{ pt: 1, pr: 2, minWidth: 120 }}>
              <Select
                aria-describedby={`iptv-provider-helper-text`}
                value={selectedIPTVProvider ?? ""}
                onChange={(e) => {
                  if (e.target.value) {
                    setSelectedIPTVProvider(e.target.value as number);
                  }
                }}
                inputProps={{ "aria-label": "IPTV Provider" }}
                sx={{
                  backgroundColor: "#161f39ff",
                  color: "white",
                  "& .MuiSvgIcon-root": {
                    color: "white",
                  },
                }}
              >
                {iptvProviders?.map((iptvProvider: any) => (
                  <MenuItem
                    key={iptvProvider?.iptv_provider_id}
                    value={iptvProvider?.iptv_provider_id}
                  >
                    {iptvProvider?.name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          )}
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
                  <ListItemText
                    primary={category?.category_name}
                    sx={{
                      color:
                        selectedCategoryID === category?.category_id
                          ? "yellow"
                          : "white",
                    }}
                  />
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
              <h3 className="text-white live-tv-title">
                {selectedChannel?.channel.name || "No Channel Selected"}
              </h3>
              {nowPlaying ? (
                <>
                  <h4 className="text-warning live-tv-title mt-2 mb-0">
                    {pickText(nowPlaying.titles)}
                  </h4>
                  <p className="text-muted mt-2 mb-0">
                    {formatTime(nowPlaying.start_time)} -{" "}
                    {formatTime(nowPlaying.stop_time)}
                  </p>
                  <p className="text-white live-tv-description mt-2">
                    {pickText(nowPlaying.descriptions)}
                  </p>
                </>
              ) : selectedChannel ? (
                <p className="text-muted live-tv-description">
                  No programme info available
                </p>
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
