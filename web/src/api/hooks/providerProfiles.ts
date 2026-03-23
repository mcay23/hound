import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createProviderProfile, deleteProviderProfile, fetchProviderProfiles } from "../services/providerProfiles";

export const useProviderProfiles = () => {
  return useQuery({
    queryKey: ["provider-profiles"],
    queryFn: fetchProviderProfiles,
  });
};

export const useCreateProviderProfile = () => {
  return useMutation({
    mutationFn: (profile: any) => createProviderProfile(profile.name, profile.manifestURL),
  });
};

export const useDeleteProviderProfile = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteProviderProfile(id),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["provider-profiles"] });
    }
  });
};