import axios from "axios";

export const fetchProviderProfiles = async () => {
  const { data } = await axios.get<any>(
    `/api/v1/provider_profiles`,
  );
  return data;
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

export const deleteProviderProfile = async(
    id: number,
) => {
    const { data } = await axios.delete<any>(`/api/v1/provider_profiles/${id}`);
    return data;
}