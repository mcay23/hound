import { useMutation, useQuery } from "@tanstack/react-query";
import { createMovieWatchHistory, createTVWatchHistory, fetchWatchActivity, fetchWatchStats } from "../services/watchHistory";

export const useWatchActivity = (
  limit: number,
  offset: number,
  startTime?: string,
  endTime?: string,
) => {
  return useQuery({
    queryKey: ["watch-activity", limit, offset, startTime, endTime],
    queryFn: () => fetchWatchActivity(limit, offset, startTime, endTime),
  });
};

export const useWatchStats = (
  startTime?: string,
  endTime?: string,
) => {
  return useQuery({
    queryKey: ["watch-activity", startTime, endTime],
    queryFn: () => fetchWatchStats(startTime, endTime),
  });
};

export const useAddTVWatchActivityMutation = () => {
  return useMutation({
    mutationFn: ({
      mediaSource,
      sourceID,
      episodeIDs,
      watchedAt,
    }: {
      mediaSource: string;
      sourceID: string;
      episodeIDs: number[];
      watchedAt?: string;
    }) => createTVWatchHistory(mediaSource, sourceID, episodeIDs, watchedAt),
  });
};

export const useAddMovieWatchActivityMutation = () => {
  return useMutation({
    mutationFn: ({
      mediaSource,
      sourceID,
      watchedAt,
    }: {
      mediaSource: string;
      sourceID: string;
      watchedAt?: string;
    }) => createMovieWatchHistory(mediaSource, sourceID, watchedAt),
  });
};
