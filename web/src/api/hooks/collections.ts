import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchAllCollections,
  fetchCollectionContents,
  fetchRecentCollectionItems,
  createCollection,
} from "../services/collections";

export const useCollections = () => {
  return useQuery({
    queryKey: ["collections", "all"],
    queryFn: fetchAllCollections,
  });
};

export const useCollectionContents = (id: number | string | undefined, enabled = true) => {
  return useQuery({
    queryKey: ["collections", id, "contents"],
    queryFn: () => fetchCollectionContents(id!),
    enabled: !!id && enabled,
  });
};

export const useRecentCollectionItems = () => {
  return useQuery({
    queryKey: ["collections", "recent"],
    queryFn: fetchRecentCollectionItems,
  });
};

export const useCreateCollection = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createCollection,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["collections", "all"] });
    },
  });
};
