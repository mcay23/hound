import { useMutation, useQuery } from "@tanstack/react-query";
import { cancelDownload, fetchDownloads, fetchMediaFiles } from "../services/media";

export const useDownloads = (
  limit: number,
  offset: number,
  refetchInterval?: number,
) => {
  return useQuery({
    queryKey: ["downloads", limit, offset],
    queryFn: () => fetchDownloads(limit, offset),
    refetchInterval,
  });
};

export const useCancelDownload = (taskID: number) => {
  return useMutation({
    mutationFn: () => cancelDownload(taskID),
  });
};

export const useMediaFiles = (
  mediaType: string,
  mediaSource: string,
  sourceID: string,
  season?: number | null,
  episode?: number | null
) => {
  return useQuery({
    queryKey: ["media-files", mediaType, mediaSource, sourceID, season, episode],
    queryFn: () => fetchMediaFiles(mediaType, mediaSource, sourceID, season, episode),
  });
};