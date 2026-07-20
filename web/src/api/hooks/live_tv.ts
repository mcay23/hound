import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchLiveTVChannels,
  fetchLiveTVCategories,
  fetchChannelEPGs,
  fetchIPTVProviders,
  createIPTVProvider,
  deleteIPTVProvider,
} from "../services/live_tv";

export interface LiveTVChannel {
  iptv_provider_id: number;
  order: number;
  stream_id: number;
  name: string;
  xtream_stream_type: string;
  thumbnail_url: string;
  epg_channel_id: string;
  category_id: string;
  added_at: string;
  stream_url: string;
}

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
  return useQuery<LiveTVChannel[]>({
    queryKey: ["live-tv-channels", iptvProviderID, categoryID],
    queryFn: () =>
      fetchLiveTVChannels(iptvProviderID as number, categoryID as number),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProviderID && !!categoryID,
  });
};

export const useChannelEPGs = (
  iptvProviderID: number | undefined,
  epgChannelIDs: string[] | undefined,
) => {
  return useQuery({
    queryKey: ["channel-epg", iptvProviderID, epgChannelIDs],
    queryFn: () =>
      fetchChannelEPGs(iptvProviderID as number, epgChannelIDs as string[]),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProviderID && !!epgChannelIDs && epgChannelIDs.length > 0,
  });
};

/*
IPTV Providers
*/

export const useIPTVProviders = () => {
  return useQuery({
    queryKey: ["iptv-providers"],
    queryFn: fetchIPTVProviders,
    staleTime: 1 * 60 * 1000,
  });
};

export const useCreateIPTVProviderMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (profile: {
      name: string;
      host: string;
      username: string;
      password: string;
    }) =>
      createIPTVProvider(
        profile.name,
        profile.host,
        profile.username,
        profile.password,
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["iptv-providers"] });
    },
  });
};

export const useDeleteIPTVProviderMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteIPTVProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["iptv-providers"] });
    },
  });
};
