import { useQuery } from "@tanstack/react-query";
import { fetchWatchActivity, fetchWatchStats } from "../services/watchHistory";

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
