import { useQuery } from "@tanstack/react-query";
import { fetchLiveTVCategories } from "../services/live_tv";

export const useLiveTVCategories = (iptvProfileID: number, categoryID: number) => {
  return useQuery({
    queryKey: ["live-tv-categories", iptvProfileID, categoryID],
    queryFn: () => fetchLiveTVCategories(iptvProfileID, categoryID),
  });
};