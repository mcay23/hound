import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  fetchBackdrops,
  fetchTrendingMovies,
  fetchTrendingTVShows,
  fetchContinueWatching,
  fetchUserHomeRows,
  fetchHomeRow,
  fetchDefaultHomeRows,
  fetchAvailableCatalogs,
  updateDefaultHomeRows,
  updateUserHomeRows,
  resetUserHomeRows,
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

export const useHomeRow = (totalRows: number) => {
  return useQueries({
    queries: Array.from({ length: totalRows }, (_, i) => i).map((id) => ({
      queryKey: ["home-rows", id],
      queryFn: () => fetchHomeRow(id),
    })),
  });
};

export const useDefaultHomeRows = () => {
  return useQuery({
    queryKey: ["default-home-rows"],
    queryFn: fetchDefaultHomeRows,
  });
};

export const useAvailableCatalogs = () => {
  return useQuery({
    queryKey: ["available-catalogs"],
    queryFn: fetchAvailableCatalogs,
  });
};

export const useUpdateDefaultHomeRowsMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateDefaultHomeRows,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["default-home-rows"] });
    },
  });
};

export const useUpdateUserHomeRowsMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateUserHomeRows,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["home-rows"] });
    },
  });
};

export const useResetUserHomeRowsMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: resetUserHomeRows,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["home-rows"] });
    },
  });
};
