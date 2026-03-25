import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createProviderProfile, deleteProviderProfile, fetchProviderProfiles, updateProviderProfile } from "../services/providerProfiles";

export const useProviderProfiles = () => {
  return useQuery({
    queryKey: ["provider-profiles"],
    queryFn: fetchProviderProfiles,
  });
};

export const useCreateProviderProfileMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (profile: {name: string, manifestURL: string}) => createProviderProfile(profile.name, profile.manifestURL),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["provider-profiles"] });
    }
  });
};

export const useUpdateProviderProfileMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (profile: {id: number, isDefaultStreaming?: boolean, isDefaultDownloading?: boolean}) => 
      updateProviderProfile(profile.id, profile.isDefaultStreaming, profile.isDefaultDownloading),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["provider-profiles"] });
    }
  });
};

export const useDeleteProviderProfileMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteProviderProfile(id),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["provider-profiles"] });
    }
  });
};