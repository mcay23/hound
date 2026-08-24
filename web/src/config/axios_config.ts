import { v4 as uuidv4 } from 'uuid';
import axios from 'axios';

export const getBaseUrl = (): string => {
  let host = localStorage.getItem("host");
  if (host) {
    host = host.trim();
    if (!host.startsWith("http://") && !host.startsWith("https://")) {
      host = "http://" + host;
    }
    return host.replace(/\/+$/, "");
  }
  return "http://localhost:2323";
};

export const setHostUrl = (host: string): string => {
  let formatted = host.trim();
  if (formatted && !formatted.startsWith("http://") && !formatted.startsWith("https://")) {
    formatted = "http://" + formatted;
  }
  formatted = formatted.replace(/\/+$/, "");
  localStorage.setItem("host", formatted);
  axios.defaults.baseURL = formatted;
  return formatted;
};

const getDeviceID = () => {
  let deviceID = localStorage.getItem("deviceID");
  if (!deviceID) {
    // we don't use crypto.randomUUID() because it's disabled in http
    deviceID = uuidv4();
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