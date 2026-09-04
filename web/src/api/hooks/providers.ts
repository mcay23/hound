import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { decodeStream, fetchProviders, fetchSubtitles } from "../services/providers";
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

const getProviderStreams = (data: any) => {
  return data?.providers?.flatMap((p: any) => p.streams || []) ?? [];
};

const getMatchingStream = (streams: any[], encodedData?: string) => {
  if (!encodedData) return undefined;
  return streams.find((stream: any) => stream.encoded_data === encodedData);
};

export const useDirectStreamMutation = () => {
  return useMutation({
    mutationFn: async ({
      mediaType,
      mediaSource,
      sourceId,
      season,
      episode,
      providerProfileId,
      encodedData,
      onImmediateStream,
    }: {
      mediaType: string;
      mediaSource: string;
      sourceId: string;
      season?: number;
      episode?: number;
      providerProfileId?: number;
      encodedData?: string;
      onImmediateStream?: (stream: any) => void;
    }) => {
      const mediaFilesPromise = fetchMediaFiles(
        mediaType,
        mediaSource,
        sourceId,
        season,
        episode,
      ).catch(() => null);
      const providersPromise = fetchProviders(
        mediaType,
        mediaSource,
        sourceId,
        season,
        episode,
        providerProfileId,
      ).catch(() => null);

      let startedImmediately = false;
      const mediaFilesData = await mediaFilesPromise;
      const mediaFilesStreams = getProviderStreams(mediaFilesData);
      const matchingMediaFileStream = getMatchingStream(
        mediaFilesStreams,
        encodedData,
      );

      if (mediaFilesStreams.length > 0 && (!encodedData || matchingMediaFileStream)) {
        startedImmediately = true;
        onImmediateStream?.(matchingMediaFileStream ?? mediaFilesStreams[0]);
      }

      const providersData = await providersPromise;
      const externalStreams = getProviderStreams(providersData);
      const allStreams = [...mediaFilesStreams, ...externalStreams];
      const selectedStream =
        getMatchingStream(allStreams, encodedData) ?? allStreams[0];

      return {
        ...providersData,
        ...mediaFilesData,
        providers: null,
        streams: allStreams,
        selectedStream,
        startedImmediately,
      };
    },
  });
};

export const useSubtitles = (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  season?: number,
  episode?: number,
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ["subtitles", mediaType, mediaSource, sourceId, season, episode],
    queryFn: () =>
      fetchSubtitles(mediaType, mediaSource, sourceId, season, episode),
    enabled,
  });
};

export const useDecodeStream = (
  encodedData: string
) => {
  return useQuery({
    queryKey: ["decode-stream", encodedData],
    queryFn: () => decodeStream(encodedData),
    enabled: !!encodedData && encodedData !== "",
  });
};
