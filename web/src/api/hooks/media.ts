import { useMutation, useQuery } from "@tanstack/react-query";
import { cancelDownload, fetchDownloads } from "../services/media";

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