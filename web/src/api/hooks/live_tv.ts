import { useQuery } from "@tanstack/react-query";
import { fetchLiveTVChannels, fetchLiveTVCategories, fetchChannelEPGs } from "../services/live_tv";

export const useLiveTVCategories = (iptvProviderID: number | undefined) => {
  return useQuery({
    queryKey: ["live-tv-categories", iptvProviderID],
    queryFn: () => fetchLiveTVCategories(iptvProviderID as number),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProviderID,
  });
};

export const useLiveTVChannels = (
  iptvProviderID: number | undefined,
  categoryID: number | undefined,
) => {
  return useQuery({
    queryKey: ["live-tv-channels", iptvProviderID, categoryID],
    queryFn: () => fetchLiveTVChannels(iptvProviderID as number, categoryID as number),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProviderID && !!categoryID,
  });
};

export const useChannelEPGs = (iptvProviderID: number | undefined, epgChannelIDs: string[] | undefined) => {
  return useQuery({
    queryKey: ["channel-epg", iptvProviderID, epgChannelIDs],
    queryFn: () => fetchChannelEPGs(iptvProviderID as number, epgChannelIDs as string[]),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProviderID && !!epgChannelIDs && epgChannelIDs.length > 0,
  });
};
