const getBaseUrl = () => {
  if (process.env.NODE_ENV === "production") {
    // relative url in prod
    return "";
  }
  return "http://localhost:2323";
};

export const SERVER_URL = getBaseUrl();
export const AXIOS_CONFIG = {
  "withCredentials": true,
  "headers": {
    "Content-Type": "application/json;charset=UTF-8",
    "X-Client": "desktop"
  }
}