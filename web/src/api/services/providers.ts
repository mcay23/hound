import axios from "axios";

export const fetchProviders = async (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  season?: number,
  episode?: number,
  providerProfileId?: number
) => {
  const { data } = await axios.get(
    `/api/v1/${mediaType}/${mediaSource}-${sourceId}/providers`,
    {
      params: { season, episode, provider_profile_id: providerProfileId },
    }
  );
  return data;
};

export const fetchSubtitles = async (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  season?: number,
  episode?: number,
) => {
  const { data } = await axios.get(
    `/api/v1/${mediaType}/${mediaSource}-${sourceId}/subtitles`,
    {
      params: { season, episode },
    }
  );
  return data;
};
