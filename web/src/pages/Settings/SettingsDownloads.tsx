import {
  Button,
  Card,
  CardContent,
  Chip,
  ChipProps,
  LinearProgress,
} from "@mui/material";
import { useDownloads } from "../../api/hooks/media";
import "./SettingsDownloads.css";
import { cancelDownload } from "../../api/services/media";

function SettingsDownloads() {
  // fetch every 2 seconds
  const { data: downloads, isLoading: isDownloadsLoading } = useDownloads(
    30,
    0,
    2000,
  );
  return (
    <div>
      <h2>Downloads</h2>
      {isDownloadsLoading ? (
        <div>Loading...</div>
      ) : (
        downloads?.tasks?.map((item: any) => {
          return <DownloadCard item={item} />;
        })
      )}
    </div>
  );
}

function DownloadCard({ item }: { item: any }) {
  const mainRecord =
    item.media_type === "movie" ? "movie_media_record" : "show_media_record";
  let title = item[mainRecord].media_title;
  if (item.media_type === "tvshow") {
    title +=
      " - S" +
      item["episode_media_record"].season_number +
      "E" +
      item["episode_media_record"].episode_number;
  }
  let statusLabel = "";
  if (item.status?.length > 0) {
    statusLabel = item.status[0]?.toUpperCase() + item.status?.slice(1);
    if (item.status === "failed") {
      statusLabel = "Download Failed";
      statusLabel += " - " + item.last_message;
    }
  }
  return (
    <Card
      variant="outlined"
      key={item.ingest_task_id}
      className="mb-2 download-card"
    >
      <CardContent>
        <h5>{title}</h5>
        <div className="text-muted">
          {"Download Type: " + item.download_type}
        </div>
        <div>{statusLabel}</div>
        {item.status === "downloading" && downloadingUI({ item })}
      </CardContent>
    </Card>
  );
}

function downloadingUI({ item }: { item: any }) {
  const downloaded = item.downloaded_bytes / 1000000;
  const total = item.total_bytes / 1000000;
  let progress = (downloaded / total) * 100;
  if (total === 0) {
    progress = 0;
  }
  return (
    <div className="mt-1">
      <div className="d-flex justify-content-between mb-1">
        <div className="text-muted">
          {total > 0 && (
            <>
              {downloaded.toFixed(2)} / {total.toFixed(0)} MB (
              {progress.toFixed(0)}
              %)
            </>
          )}
        </div>
        <div className="text-muted">
          {item.download_speed > 0 &&
            (item.download_speed / 1000000).toFixed(2) + " MB/s"}
        </div>
      </div>

      <LinearProgress variant="determinate" value={progress} />

      <div className="d-flex justify-content-end">
        <Button
          className="mt-2"
          variant="outlined"
          size="small"
          onClick={() => cancelDownload(item.ingest_task_id)}
        >
          Cancel
        </Button>
      </div>
    </div>
  );
}

function getStatusChip(status: string) {
  let color: ChipProps["color"] = "primary";
  if (status === "done") {
    color = "success";
  } else if (status === "failed") {
    color = "error";
  }
  const label = status[0].toUpperCase() + status.slice(1);
  return <Chip label={label} color={color} />;
}
export default SettingsDownloads;
