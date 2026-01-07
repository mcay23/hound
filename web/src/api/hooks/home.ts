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
    select: (data) => data.data,
  });
};

export const useTrendingMovies = () => {
  return useQuery({
    queryKey: ["trending", "movie"],
    queryFn: fetchTrendingMovies,
    select: (data) => data.data,
  });
};

export const useBackdrops = () => {
  return useQuery({
    queryKey: ["backdrops"],
    queryFn: fetchBackdrops,
    select: (data) => data.data,
  });
};

export const useContinueWatching = () => {
  return useQuery({
    queryKey: ["continue-watching"],
    queryFn: fetchContinueWatching,
    select: (data) => data.data,
  });
};