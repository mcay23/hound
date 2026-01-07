import axios from "axios";

export const fetchTrendingTVShows = async () => {
  const { data } = await axios.get("/api/v1/tv/trending");
  return data;
};

export const fetchTrendingMovies = async () => {
  const { data } = await axios.get("/api/v1/movie/trending");
  return data;
};

export const fetchBackdrops = async () => {
  const { data } = await axios.get("/api/v1/backdrops");
  return data;
};

export const fetchContinueWatching = async () => {
  const { data } = await axios.get("/api/v1/continue_watching");
  return data;
};
