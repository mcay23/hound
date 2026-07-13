import axios from "axios";

export const fetchLiveTVCategories = async (iptvProfileID: number) => {
  const { data } = await axios.get(`/api/v1/live/${iptvProfileID}/categories`);
  return data;
};

export const fetchLiveTVChannels = async (iptvProfileID: number, categoryID: number) => {
  const { data } = await axios.get(`/api/v1/live/${iptvProfileID}/channels?category_id=${categoryID}`);
  return data.sort((a: any, b: any) => a.order > b.order ? 1 : -1);
};

export const fetchChannelEPG = async (iptvProfileID: number, epgChannelID: string) => {
  const { data } = await axios.get(`/api/v1/live/${iptvProfileID}/epg?epg_channel_id=${epgChannelID}`);
  return data
};