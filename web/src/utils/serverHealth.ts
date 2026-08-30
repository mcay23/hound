import { getBaseUrl } from "../config/axios_config";

export const checkServerReachable = async (token?: string): Promise<boolean> => {
  try {
    const headers: HeadersInit = {};
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    await fetch(`${getBaseUrl()}/api/v1/server_info`, {
      headers,
      signal: AbortSignal.timeout(5000),
    });
    return true;
  } catch {
    return false;
  }
};
