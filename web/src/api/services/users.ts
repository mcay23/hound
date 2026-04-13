import axios from "axios";

export const fetchUsers = async () => {
  const { data } = await axios.get<any>(
    `/api/v1/users`,
  );
  return data;
};

export const createUser = async (username: string, displayName: string, password: string) => {
  const { data } = await axios.post<any>(`/api/v1/users`, {
    username,
    display_name: displayName,
    password,
  });
  return data;
};

export const deleteUser = async (id: number) => {
  const { data } = await axios.delete<any>(`/api/v1/users/${id}`);
  return data;
};

export const resetUserPassword = async (userID: number, newPassword: string) => {
  const { data } = await axios.post<any>(`/api/v1/users/${userID}/password`, {
    "new_password": newPassword,
  });
  return data;
};