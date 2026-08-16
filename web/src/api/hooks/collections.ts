import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchAllCollections,
  fetchCollectionContents,
  fetchRecentCollectionItems,
  createCollection,
  updateCollection,
} from "../services/collections";

export const useCollections = () => {
  return useQuery({
    queryKey: ["collections", "all"],
    queryFn: fetchAllCollections,
  });
};

export const useCollectionContents = (id: number | string | undefined, limit?: number, offset?: number, enabled = true) => {
  return useQuery({
    queryKey: ["collections", id, "contents"],
    queryFn: () => fetchCollectionContents(id!, limit, offset),
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

export const useUpdateCollection = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateCollection,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["collections", variables.collectionID, "contents"] });
    },
  });
};
