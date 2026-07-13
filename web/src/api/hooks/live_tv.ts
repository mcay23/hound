import { useQuery } from "@tanstack/react-query";
import { fetchLiveTVChannels, fetchLiveTVCategories, fetchChannelEPGs } from "../services/live_tv";

export const useLiveTVCategories = (iptvProfileID: number) => {
  return useQuery({
    queryKey: ["live-tv-categories", iptvProfileID],
    queryFn: () => fetchLiveTVCategories(iptvProfileID),
    staleTime: 30 * 60 * 1000,
  });
};

export const useLiveTVChannels = (
  iptvProfileID: number | undefined,
  categoryID: number | undefined,
) => {
  return useQuery({
    queryKey: ["live-tv-channels", iptvProfileID, categoryID],
    queryFn: () => fetchLiveTVChannels(iptvProfileID as number, categoryID as number),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProfileID && !!categoryID,
  });
};

export const useChannelEPGs = (iptvProfileID: number | undefined, epgChannelIDs: string[] | undefined) => {
  return useQuery({
    queryKey: ["channel-epg", iptvProfileID, epgChannelIDs],
    queryFn: () => fetchChannelEPGs(iptvProfileID as number, epgChannelIDs as string[]),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProfileID && !!epgChannelIDs && epgChannelIDs.length > 0,
  });
};
