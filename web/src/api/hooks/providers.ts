import { useMutation, useQuery } from "@tanstack/react-query";
import { fetchProviders } from "../services/providers";
import { fetchMediaFiles } from "../services/media";

export const useProviders = (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  season?: number,
  episode?: number,
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ["providers", mediaType, mediaSource, sourceId, season, episode],
    queryFn: () =>
      fetchProviders(mediaType, mediaSource, sourceId, season, episode),
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
    }: {
      mediaType: string;
      mediaSource: string;
      sourceId: string;
      season?: number;
      episode?: number;
    }) => fetchProviders(mediaType, mediaSource, sourceId, season, episode),
  });
};

export const useUnifiedStreams = (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  season?: number,
  episode?: number,
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ["unified-streams", mediaType, mediaSource, sourceId, season, episode],
    queryFn: async () => {
      const [mediaFilesData, providersData] = await Promise.all([
        fetchMediaFiles(mediaType, mediaSource, sourceId, season, episode).catch((err) => {console.log("mediaFilesData failed", err); return null}),
        fetchProviders(mediaType, mediaSource, sourceId, season, episode).catch((err) => {console.log("providersData failed", err); return null}),
      ]);

      // media files return their own ProviderResponseObject mimicking the main providers response
      const mediaFilesProviders = mediaFilesData?.providers || [];
      const externalProviders = providersData?.providers || [];

      const mediaFilesStreams = mediaFilesProviders.flatMap((p: any) => p.streams || []);
      const externalStreams = externalProviders.flatMap((p: any) => p.streams || []);
      const allStreams = [...mediaFilesStreams, ...externalStreams];
      
      return {
        ...providersData,
        ...mediaFilesData,
        streams: allStreams,
      };
    },
    enabled,
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
    }: {
      mediaType: string;
      mediaSource: string;
      sourceId: string;
      season?: number;
      episode?: number;
    }) => {
      const [mediaFilesData, providersData] = await Promise.all([
        fetchMediaFiles(mediaType, mediaSource, sourceId, season, episode).catch(() => null),
        fetchProviders(mediaType, mediaSource, sourceId, season, episode).catch(() => null),
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