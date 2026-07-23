import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchLiveTVChannels,
  fetchXtreamCategories,
  fetchChannelEPGs,
  fetchIPTVProviders,
  createIPTVProvider,
  deleteIPTVProvider,
} from "../services/live_tv";

interface LiveTVChannelResponse {
  total: number;
  added: number;
  channels: LiveTVChannel[]
}

export interface LiveTVChannel {
  iptv_provider_id: number;
  order: number;
  stream_id: number;
  name: string;
  stream_type: string;
  thumbnail_url: string;
  epg_channel_id: string;
  category_id: string;
  added_at: string;
  stream_url: string;
}

export interface IPTVProvider {
  iptv_provider_id: number;
  iptv_provider_type: string;
  name: string;
  host: string;
  username: string;
  is_default: boolean;
  last_refresh: string;
}

export const useXtreamCategories = (iptvProviderID: number | undefined, iptvProviderType: string | undefined) => {
  return useQuery({
    queryKey: ["xtream-categories", iptvProviderID],
    queryFn: () => fetchXtreamCategories(iptvProviderID as number),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProviderID && iptvProviderType === "xtream",
  });
};

export const useLiveTVChannels = (
  iptvProvider: any | undefined,
  categoryID: number | undefined,
) => {
  // categoryID must exist for xtream only
  return useQuery<LiveTVChannelResponse>({
    queryKey: ["live-tv-channels", iptvProvider?.iptv_provider_id, categoryID],
    queryFn: () =>
      fetchLiveTVChannels(iptvProvider?.iptv_provider_id as number, categoryID),
    enabled: !!iptvProvider && (!!categoryID || iptvProvider?.iptv_provider_type === "m3u8"),
  });
};

export const useChannelEPGs = (
  iptvProviderID: number | undefined,
  iptvProviderType: string | undefined,
  epgChannelIDs: string[] | undefined,
) => {
  return useQuery({
    queryKey: ["channel-epg", iptvProviderID, epgChannelIDs],
    queryFn: () =>
      fetchChannelEPGs(iptvProviderID as number, epgChannelIDs as string[]),
    staleTime: 30 * 60 * 1000,
    enabled: !!iptvProviderID && iptvProviderType === "xtream" && !!epgChannelIDs && epgChannelIDs.length > 0,
  });
};

/*
IPTV Providers
*/

export const useIPTVProviders = () => {
  return useQuery<IPTVProvider[]>({
    queryKey: ["iptv-providers"],
    queryFn: fetchIPTVProviders,
    staleTime: 1 * 60 * 1000,
  });
};

export const useCreateIPTVProviderMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (profile: {
      iptvProviderType: string;
      name: string;
      host: string;
      username: string | null;
      password: string | null;
    }) =>
      createIPTVProvider(
        profile.iptvProviderType,
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
