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
    movies_watched: number,
    shows_watched: number,
    episodes_watched: number,
    total_movies_duration: number,
    total_episodes_duration: number,
}

export const fetchWatchStats = async (
    startTime?: string,
    endTime?: string
) => {
  const { data } = await axios.get<WatchStats>(
    "/api/v1/watch_stats",
    {
      params: {
        start_time: startTime,
        end_time: endTime,
      },
    },
  );
  return data;
};