import { useQueries, useQuery } from "@tanstack/react-query";
import {
  fetchBackdrops,
  fetchTrendingMovies,
  fetchTrendingTVShows,
  fetchContinueWatching,
  fetchUserHomeRows,
  fetchHomeRow,
  fetchDefaultHomeRows,
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

export const useUserHomeRows = () => {
  return useQuery({
    queryKey: ["home-rows"],
    queryFn: fetchUserHomeRows,
  });
};

export const useHomeRow = (length: number) => {
  return useQueries({
    queries: Array.from({ length: length }, (_, i) => i).map((id) => ({
      queryKey: ["home-rows", id],
      queryFn: () => fetchHomeRow(id),
    })),
  })
};

export const useDefaultHomeRows = () => {
  return useQuery({
    queryKey: ["default-home-rows"],
    queryFn: fetchDefaultHomeRows,
  });
};