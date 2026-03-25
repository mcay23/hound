import axios from "axios";

export const fetchProviderProfiles = async () => {
  const { data } = await axios.get<any>(
    `/api/v1/provider_profiles`,
  );
  return data.sort((a: any, b: any) => a.provider_profile_id > b.provider_profile_id ? 1 : -1);
};

export const createProviderProfile = async (
  name: string,
  manifestURL: string,
) => {
  const { data } = await axios.post<any>(`/api/v1/provider_profiles`, {
    name,
    manifest_url: manifestURL,
  });
  return data;
};

export const updateProviderProfile = async (
  id: number,
  isDefaultStreaming?: boolean,
  isDefaultDownloading?: boolean,
) => {
  const { data } = await axios.put<any>(`/api/v1/provider_profiles/${id}`, {
    is_default_streaming: isDefaultStreaming,
    is_default_downloading: isDefaultDownloading,
  });
  return data;
};

export const deleteProviderProfile = async(
    id: number,
) => {
    const { data } = await axios.delete<any>(`/api/v1/provider_profiles/${id}`);
    return data;
}