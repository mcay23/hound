import axios from "axios";

export const fetchLiveTVCategories = async (iptvProviderID: number) => {
  const { data } = await axios.get(`/api/v1/live/${iptvProviderID}/categories`);
  return data;
};

export const fetchLiveTVChannels = async (iptvProviderID: number, categoryID: number) => {
  const { data } = await axios.get(`/api/v1/live/${iptvProviderID}/channels?category_id=${categoryID}`);
  return data.sort((a: any, b: any) => a.order > b.order ? 1 : -1);
};

export const fetchChannelEPGs = async (iptvProviderID: number, epgChannelIDs: string[]) => {
  const { data } = await axios.post(`/api/v1/live/${iptvProviderID}/epg`, { epg_channel_ids: epgChannelIDs });
  return data
};