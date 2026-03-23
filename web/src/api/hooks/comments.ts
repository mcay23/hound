import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createComment, deleteComment, fetchComments } from "../services/comments";

export const useComments = (
  commentType: "review" | "comment" | "note",
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  seasonNumber?: number,
  episodeNumber?: number,
) => {
  return useQuery({
    queryKey: ["comments", commentType, mediaType, mediaSource, sourceId, seasonNumber, episodeNumber],
    queryFn: () => fetchComments(mediaType, mediaSource, sourceId, commentType, seasonNumber, episodeNumber),
  });
};

export const useCreateComment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => createComment(
      data.mediaType,
      data.mediaSource,
      data.sourceId,
      data.commentType,
      data.comment,
      data.score,
      data.title,
      data.seasonNumber,
      data.episodeNumber,
    ),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["comments", variables.commentType, variables.mediaType, variables.mediaSource, variables.sourceId, variables.seasonNumber, variables.episodeNumber] });
    },
  });
};

export const useDeleteComment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (commentId: number) => deleteComment(commentId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["comments"]});
    },
  });
};