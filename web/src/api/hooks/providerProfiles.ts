import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createProviderProfile, deleteProviderProfile, fetchProviderProfiles } from "../services/providerProfiles";

export const useProviderProfiles = () => {
  return useQuery({
    queryKey: ["provider-profiles"],
    queryFn: fetchProviderProfiles,
  });
};

export const useCreateProviderProfile = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (profile: {name: string, manifestURL: string}) => createProviderProfile(profile.name, profile.manifestURL),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["provider-profiles"] });
    }
  });
};

export const useDeleteProviderProfile = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteProviderProfile(id),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["provider-profiles"] });
    }
  });
};