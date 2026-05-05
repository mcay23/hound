import { useQuery } from "@tanstack/react-query";
import { fetchServerInfo } from "../services/general";

export const useServerInfo = () => {
  return useQuery({
    queryKey: ["server_info"],
    staleTime: 1000 * 60 * 10, // 10 minutes
    queryFn: fetchServerInfo,
  });
};