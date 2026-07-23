import axios from "axios";
import { LiveTVChannel } from "../hooks/live_tv";

export const fetchXtreamCategories = async (iptvProviderID: number) => {
  const { data } = await axios.get(`/api/v1/live/${iptvProviderID}/categories`);
  return data;
};

export const fetchLiveTVChannels = async (
  iptvProviderID: number,
  categoryID: number | undefined,
) => {
  const { data } = await axios.get(
    `/api/v1/live/${iptvProviderID}/channels${categoryID ? `?category_id=${categoryID}` : ''}`,
  );
  data.channels?.sort((a: LiveTVChannel, b: LiveTVChannel) => a.order - b.order);
  return data
};

export const fetchChannelEPGs = async (
  iptvProviderID: number,
  epgChannelIDs: string[],
) => {
  const { data } = await axios.post(`/api/v1/live/${iptvProviderID}/epg`, {
    epg_channel_ids: epgChannelIDs,
  });
  return data;
};

/*
 IPTV Providers
*/

export const fetchIPTVProviders = async () => {
  const { data } = await axios.get(`/api/v1/iptv_providers`);
  return data;
};

export const createIPTVProvider = async (
  iptvProviderType: string,
  name: string,
  host: string,
  username: string | null,
  password: string | null,
) => {
  const { data } = await axios.post<any>(`/api/v1/iptv_providers`, {
    iptv_provider_type: iptvProviderType,
    name,
    host,
    username,
    password,
  });
  return data;
};

export const deleteIPTVProvider = async (id: number) => {
  const { data } = await axios.delete<any>(`/api/v1/iptv_providers/${id}`);
  return data;
};
