import axios from "axios";

export interface Comment {
  comment_id: number;
  comment_type: "review" | "comment" | "note";
  user_id: string;
  record_id: number;
  is_public: boolean;
  title: string;
  comment: string;
  score: number;
  created_at: string;
  updated_at: string;
}

export const fetchComments = async (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  commentType: "review" | "comment" | "note",
  seasonNumber?: number,
  episodeNumber?: number,
) => {
  const { data } = await axios.get<any>(
    `/api/v1/${mediaType === "tvshow" ? "tv" : mediaType}/${mediaSource}-${sourceId}/comments`,
    {
      params: {
        type: commentType,
        season_number: seasonNumber,
        episode_number: episodeNumber,
      },
    },
  );
  return data;
};

export const createComment = async (
  mediaType: string,
  mediaSource: string,
  sourceId: string,
  commentType: "review" | "comment" | "note",
  comment: string,
  score?: number,
  title?: string,
  seasonNumber?: number,
  episodeNumber?: number,
) => {
  const { data } = await axios.post<any>(
    `/api/v1/${mediaType === "tvshow" ? "tv" : mediaType}/${mediaSource}-${sourceId}/comments`,
    {
      comment_type: commentType,
      title,
      comment,
      score,
      season_number: seasonNumber,
      episode_number: episodeNumber,
    },
  );
  return data;
};

export const deleteComment = async (commentId: number) => {
  const { data } = await axios.delete<any>(`/api/v1/comments/${commentId}`);
  return data;
};