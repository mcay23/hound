import axios from "axios";

export const fetchTrendingTVShows = async () => {
  const { data } = await axios.get("/api/v1/catalog/trending-shows");
  return data;
};

export const fetchTrendingMovies = async () => {
  const { data } = await axios.get("/api/v1/catalog/trending-movies");
  return data;
};

export const fetchBackdrops = async () => {
  const { data } = await axios.get("/api/v1/backdrop");
  return data;
};

export const fetchContinueWatching = async () => {
  const { data } = await axios.get("/api/v1/continue_watching");
  return data;
};

export const fetchUserHomeRows = async () => {
  const { data } = await axios.get("/api/v1/home");
  return data;
};

export const fetchHomeRow = async (homeRowIndex: number) => {
  const { data } = await axios.get(`/api/v1/home/${homeRowIndex}`);
  return data;
};