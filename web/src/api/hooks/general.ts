import { useQuery } from "@tanstack/react-query";
import { fetchServerInfo } from "../services/general";

export const useServerInfo = () => {
    return useQuery({
        queryKey: ["server_info"],
        queryFn: fetchServerInfo,
    });
}