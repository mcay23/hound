const getBaseUrl = () => {
  if (process.env.NODE_ENV === "production") {
    // relative url in prod
    return "";
  }
  return "http://localhost:2323";
};

const getDeviceID = () => {
  let deviceID = localStorage.getItem("deviceID");
  if (!deviceID) {
    deviceID = crypto.randomUUID();
    localStorage.setItem("deviceID", deviceID);
  }
  return deviceID;
}

export const SERVER_URL = getBaseUrl();
export const AXIOS_CONFIG = {
  "withCredentials": true,
  "headers": {
    "Content-Type": "application/json;charset=UTF-8",
    "X-Client-Id": "hound-web",
    "X-Client-Platform": "web",
    "X-Device-Id": getDeviceID()
  }
}