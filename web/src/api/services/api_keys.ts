import axios from "axios";

export const fetchApiKeys = async () => {
  const { data } = await axios.get<any>(
    `/api/v1/api_keys`,
  );
  return data;
};

export const createApiKey = async (name: string) => {
  const { data } = await axios.post<any>(`/api/v1/api_keys`, {
    name,
  });
  return data;
};

export const deleteApiKey = async (keyId: number) => {
  const { data } = await axios.delete<any>(`/api/v1/api_keys/${keyId}`);
  return data;
};
