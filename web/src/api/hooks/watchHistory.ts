import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createMovieWatchHistory, createTVWatchHistory, fetchTVSeasonHistory, fetchWatchActivity, fetchWatchStats } from "../services/watchHistory";

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
    queryKey: ["watch-stats", startTime, endTime],
    queryFn: () => fetchWatchStats(startTime, endTime),
  });
};

export const useTVSeasonHistory = (
  mediaSource: string,
  sourceID: string,
  seasonNumber: number,
  enabled: boolean = true,
) => {
  return useQuery({
    queryKey: ["tv-season-history", mediaSource, sourceID, seasonNumber],
    queryFn: () => fetchTVSeasonHistory(mediaSource, sourceID, seasonNumber),
    enabled: enabled && !!mediaSource && !!sourceID && seasonNumber !== undefined,
  });
};

export const useAddTVWatchHistoryMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      mediaSource,
      sourceID,
      episodeIDs,
      watchedAt,
      seasonNumber,
      episodeNumber,
    }: {
      mediaSource: string;
      sourceID: string;
      episodeIDs: number[];
      watchedAt?: string;
      seasonNumber?: number;
      episodeNumber?: number;
    }) =>
      createTVWatchHistory(
        mediaSource,
        sourceID,
        episodeIDs,
        watchedAt,
        seasonNumber,
        episodeNumber,
      ),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["watch-activity"] });
      queryClient.invalidateQueries({ queryKey: ["watch-stats"] });
      queryClient.invalidateQueries({ 
        queryKey: ["tv-season-history", variables.mediaSource, variables.sourceID] 
      });
    },
  });
};

export const useAddMovieWatchActivityMutation = () => {
  const queryClient = useQueryClient();
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["watch-activity"] });
      queryClient.invalidateQueries({ queryKey: ["watch-stats"] });
    },
  });
};
