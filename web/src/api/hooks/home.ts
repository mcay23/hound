import { useQuery } from "@tanstack/react-query";
import {
  fetchBackdrops,
  fetchTrendingMovies,
  fetchTrendingTVShows,
  fetchContinueWatching,
} from "../services/home";

export const useTrendingTVShows = () => {
  return useQuery({
    queryKey: ["trending", "tv"],
    queryFn: fetchTrendingTVShows,
  });
};

export const useTrendingMovies = () => {
  return useQuery({
    queryKey: ["trending", "movie"],
    queryFn: fetchTrendingMovies,
  });
};

export const useBackdrops = () => {
  return useQuery({
    queryKey: ["backdrops"],
    queryFn: fetchBackdrops,
  });
};

export const useContinueWatching = () => {
  return useQuery({
    queryKey: ["continue-watching"],
    queryFn: fetchContinueWatching,
  });
};
