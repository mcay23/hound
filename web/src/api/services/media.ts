import axios from "axios";

interface GetIngestTasksResponse {
  total_records: number;
  limit: number;
  offset: number;
  tasks: IngestTaskFullRecord[];
}

interface IngestTaskFullRecord {
  ingest_task_id: string;
  record_id: string;
  status: string;
  download_protocol: string;
  source_uri: string;
  file_idx: string;
  last_message: string;
  source_path: string;
  destination_path: string;
  total_bytes: number;
  downloaded_bytes: number;
  download_speed: number;
  movie_media_record: any;
  show_media_record: any;
  episode_media_record: any;
  created_at: string;
  updated_at: string;
  started_at: string;
  finished_at: string;
  media_type: string;
}

export const MatchTypeString = "match_string" as const;
export const MatchTypeInfoHash = "info_hash" as const;

export type DownloadPreference =
  | {
      match_type: typeof MatchTypeString;
      string_match_preference: {
        match_string: string;
        case_sensitive: boolean;
      };
    }
  | {
      match_type: typeof MatchTypeInfoHash;
      info_hash_preference: {
        info_hash: string;
      };
    };

export interface SeasonDownloadPreferences {
  episodes_to_download: number[];
  strict_match: boolean;
  skip_downloaded_episodes: boolean;
  preference_list: DownloadPreference[];
}

export const fetchDownloads = async (limit: number, offset: number) => {
  const { data } = await axios.get<GetIngestTasksResponse>("/api/v1/ingest", {
    params: {
      limit: limit,
      offset: offset,
    },
  });
  return data;
};

export const cancelDownload = async (taskID: number) => {
  const { data } = await axios.post(`/api/v1/ingest/${taskID}/cancel`);
  return data;
};

export const fetchMediaFiles = async (
  mediaType: string,
  mediaSource: string,
  sourceID: string,
  season?: number | null,
  episode?: number | null,
) => {
  const { data } = await axios.get<any>(
    `/api/v1/${mediaType}/${mediaSource}-${sourceID}/media_files`,
    {
      params: mediaType === "tv" ? { season, episode } : {},
    },
  );
  return data;
};

type DownloadSeasonParams = {
  mediaType: string;
  mediaSource: string;
  sourceID: string;
  seasonNum?: number | null;
  preferences?: SeasonDownloadPreferences;
};

export const downloadSeason = async ({
  mediaType,
  mediaSource,
  sourceID,
  seasonNum,
  preferences,
}: DownloadSeasonParams) => {
  const { data } = await axios.post(
    `/api/v1/${mediaType}/${mediaSource}-${sourceID}/${seasonNum}/download`,
    preferences,
  );
  return data;
};
