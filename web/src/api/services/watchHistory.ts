import axios from "axios";

export interface WatchActivity {
  watch_event_id: number;
  rewatch_id: number;
  record_id: number;
  watch_type: string;
  watched_at: string;
  record_type: "movie" | "episode";
  media_source: string;
  source_id: string;
  media_title: string;
  show_media_title?: string;
  show_source_id?: string;
  season_number?: number;
  episode_number?: number;
  release_date: string;
  overview: string;
  duration: number;
}

interface GetWatchActivityResponse {
  watch_activity: WatchActivity[];
  total_records: number;
  limit: number;
  offset: number;
}

export const fetchWatchActivity = async (
  limit: number,
  offset: number,
  startTime?: string,
  endTime?: string,
) => {
  const { data } = await axios.get<GetWatchActivityResponse>(
    "/api/v1/watch_activity",
    {
      params: {
        limit,
        offset,
        start_time: startTime,
        end_time: endTime,
      },
    },
  );
  return data;
};

export interface WatchStats {
  movies_watched: number;
  shows_watched: number;
  episodes_watched: number;
  total_movies_duration: number;
  total_episodes_duration: number;
}

export const fetchWatchStats = async (startTime?: string, endTime?: string) => {
  const { data } = await axios.get<WatchStats>("/api/v1/watch_stats", {
    params: {
      start_time: startTime,
      end_time: endTime,
    },
  });
  return data;
};

export const fetchTVSeasonHistory = async (
  mediaSource: string,
  sourceID: string,
  seasonNumber: number,
) => {
  const { data } = await axios.get<any>(
    `/api/v1/tv/${mediaSource}-${sourceID}/season/${seasonNumber}/history`,
  );
  return data;
};

export const createTVWatchHistory = async (
  mediaSource: string,
  sourceID: string,
  episodeIDs: number[],
  watchedAt?: string,
  seasonNumber?: number,
  episodeNumber?: number,
) => {
  const payload: any = {
    action_type: "watch",
  };
  if (watchedAt) {
    payload.watched_at = watchedAt;
  }
  if (seasonNumber !== undefined && episodeNumber !== undefined) {
    payload.season_number = seasonNumber;
    payload.episode_number = episodeNumber;
  } else {
    payload.episode_ids = episodeIDs;
  }

  const { data } = await axios.post(
    `/api/v1/tv/${mediaSource}-${sourceID}/history`,
    payload,
  );
  return data;
};

export const createMovieWatchHistory = async (
  mediaType: string,
  sourceID: string,
  watchedAt?: string,
) => {
  const { data } = await axios.post<any>(`/api/v1/movie/${mediaType}-${sourceID}/history`, {
    action_type: "watch",
    watched_at: watchedAt,
  });
  return data;
};