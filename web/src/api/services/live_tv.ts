import axios from "axios";

export const fetchLiveTVCategories = async (iptvProfileID: number, categoryID: number) => {
  const { data } = await axios.get(`/api/v1/live/${iptvProfileID}/channels?category_id=${categoryID}`);
  return data;
};