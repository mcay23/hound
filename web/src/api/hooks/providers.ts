import { useMutation, useQuery } from "@tanstack/react-query";
import { fetchProviders } from "../services/providers";
import { fetchMediaFiles } from "../services/media";

export const useProviders = (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  season?: number,
  episode?: number,
  providerProfileId?: number,
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ["providers", mediaType, mediaSource, sourceId, season, episode, providerProfileId],
    queryFn: () =>
      fetchProviders(mediaType, mediaSource, sourceId, season, episode, providerProfileId),
    enabled,
  });
};

export const useProvidersMutation = () => {
  return useMutation({
    mutationFn: ({
      mediaType,
      mediaSource,
      sourceId,
      season,
      episode,
      providerProfileId,
    }: {
      mediaType: string;
      mediaSource: string;
      sourceId: string;
      season?: number;
      episode?: number;
      providerProfileId?: number;
    }) => fetchProviders(mediaType, mediaSource, sourceId, season, episode, providerProfileId),
  });
};

export const useUnifiedStreamsMutation = () => {
  return useMutation({
    mutationFn: async ({
      mediaType,
      mediaSource,
      sourceId,
      season,
      episode,
      providerProfileId,
    }: {
      mediaType: string;
      mediaSource: string;
      sourceId: string;
      season?: number;
      episode?: number;
      providerProfileId?: number;
    }) => {
      const [mediaFilesData, providersData] = await Promise.all([
        fetchMediaFiles(mediaType, mediaSource, sourceId, season, episode).catch(() => null),
        fetchProviders(mediaType, mediaSource, sourceId, season, episode, providerProfileId).catch(() => null),
      ]);
      const mediaFilesProviders = mediaFilesData?.providers || [];
      const externalProviders = providersData?.providers || [];

      const mediaFilesStreams = mediaFilesProviders.flatMap((p: any) => p.streams || []);
      const externalStreams = externalProviders
        .flatMap((p: any) => p.streams || []);
      const allStreams = [...mediaFilesStreams, ...externalStreams];
      
      return {
        ...providersData,
        ...mediaFilesData,
        providers: null,
        streams: allStreams,
      };
    },
  });
};