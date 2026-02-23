import axios from "axios";

export const fetchProviders = async (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  season?: number,
  episode?: number
) => {
  const { data } = await axios.get(
    `/api/v1/${mediaType}/${mediaSource}-${sourceId}/providers`,
    {
      params: { season, episode },
    }
  );
  return data;
};
