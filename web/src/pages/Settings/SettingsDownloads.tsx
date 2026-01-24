import {
  Button,
  Card,
  CardContent,
  Chip,
  ChipProps,
  Divider,
  LinearProgress,
  Pagination,
} from "@mui/material";
import { useDownloads } from "../../api/hooks/media";
import "./SettingsDownloads.css";
import { cancelDownload } from "../../api/services/media";
import { useState } from "react";

function SettingsDownloads() {
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalRecords, setTotalRecords] = useState(0);
  const itemsPerPage = 10;
  const handlePageChange = (
    event: React.ChangeEvent<unknown>,
    value: number,
  ) => {
    setPage(value);
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: "smooth",
    });
  };
  // fetch every 2 seconds
  const { data: downloads, isLoading: isDownloadsLoading } = useDownloads(
    itemsPerPage,
    (page - 1) * itemsPerPage,
    2000,
  );
  if (downloads && downloads.total_records !== totalRecords) {
    setTotalRecords(downloads.total_records);
    setTotalPages(Math.ceil(downloads.total_records / itemsPerPage));
  }
  return (
    <div>
      <h2>Downloads</h2>
      <hr />
      {isDownloadsLoading ? (
        <div>Loading...</div>
      ) : (
        downloads?.tasks?.map((item: any) => {
          return <DownloadCard item={item} />;
        })
      )}
      <div className="paginator-container shadow-lg">
        <Pagination
          id="paginator-component"
          defaultPage={1}
          page={page}
          onChange={handlePageChange}
          count={totalPages}
          size="large"
        />
      </div>
    </div>
  );
}

function DownloadCard({ item }: { item: any }) {
  const mainRecord =
    item.media_type === "movie" ? "movie_media_record" : "show_media_record";
  let title = item[mainRecord]?.media_title;
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
      <CardContent className="download-card-content">
        <h5>{title}</h5>
        <div className="text-muted">
          {"Download Type: " + item.download_type}
        </div>
        <div>{statusLabel}</div>
        {item.status === "downloading" && downloadingUI({ item })}
        {item.status === "done" && doneUI({ item })}
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
              {downloaded.toFixed(2)} / {total.toFixed(2)} MB (
              {progress.toFixed(0)}
              %)
            </>
          )}
        </div>
        <div className="text-muted">
          {item.connected_seeders > 0 &&
            "(" + item.connected_seeders + " seeders) "}
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

function doneUI({ item }: { item: any }) {
  return (
    <div className="mt-1">
      <div className="d-flex justify-content-between mb-1">
        <div className="text-muted">
          {(item.downloaded_bytes / 1000000).toFixed(2)} /{" "}
          {(item.total_bytes / 1000000).toFixed(2)} MB
        </div>
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
