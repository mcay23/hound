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
  download_type: string;
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
