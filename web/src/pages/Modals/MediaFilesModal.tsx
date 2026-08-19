import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
} from "@mui/material";
import {
  useDeleteMediaFileMutation,
  useMediaFiles,
} from "../../api/hooks/media";
import { useMemo, useState } from "react";
import MUIDataTable, {
  MUIDataTableColumnOptions,
  MUIDataTableOptions,
} from "mui-datatables";
import toast from "react-hot-toast";

interface Stream {
  file_id: number;
  media_type: string;
  media_source: string;
  source_id: string;
  season_number: number;
  episode_number: number;
  episode_source_id: string;
  provider_profile_name: string;
  provider_profile_id: number;
  stream_protocol: string;
  file_origin: string;
  uri: string;
  title: string;
}

interface MediaFilesModalProps {
  mediaType: string;
  mediaSource: string;
  sourceID: string;
  onClose: () => void;
  open: boolean;
  season?: number | null;
  episode?: number | null;
}

export function MediaFilesModal({
  mediaType,
  mediaSource,
  sourceID,
  onClose,
  open,
  season,
  episode,
}: MediaFilesModalProps) {
  const { data: mediaFiles, isLoading } = useMediaFiles(
    mediaType,
    mediaSource,
    sourceID,
    season,
    episode,
  );
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [filesToDelete, setFilesToDelete] = useState<number[]>([]);
  const [rowsSelected, setRowsSelected] = useState<number[]>([]);
  const deleteMediaFileMutation = useDeleteMediaFileMutation();

  const streams = useMemo(() => {
    if (mediaType === "tvshow" || mediaType === "tv") {
      return [...(mediaFiles?.providers?.[0]?.streams ?? [])].sort(
        (a, b) =>
          a.season_number - b.season_number ||
          a.episode_number - b.episode_number,
      );
    } else {
      return mediaFiles?.providers?.[0]?.streams || [];
    }
  }, [mediaFiles, mediaType]);

  const rows = useMemo(
    () =>
      streams.map((stream: Stream) => [
        stream?.title,
        stream?.season_number,
        stream?.episode_number,
        stream?.file_origin,
        stream?.uri,
        stream?.file_id,
      ]),
    [streams],
  );

  const handleConfirmDelete = async () => {
    setDeleteConfirmOpen(false);
    const targetFiles = [...filesToDelete];
    setFilesToDelete([]);
    setRowsSelected([]);
    if (targetFiles.length === 0) return;

    const toastId = toast.loading(
      targetFiles.length === 1
        ? "Deleting media file..."
        : `Deleting ${targetFiles.length} media files...`,
    );
    let errorCount = 0;
    for (const fileID of targetFiles) {
      try {
        await deleteMediaFileMutation.mutateAsync(fileID);
      } catch (err: any) {
        errorCount++;
        console.error("Error deleting media file:", err);
        const errorMsg =
          err?.response?.data?.message ||
          err?.message ||
          "Failed to delete media file";
        toast.error(`Failed to delete media file: ${errorMsg}`);
      }
    }
    toast.dismiss(toastId);
    if (errorCount === 0) {
      toast.success(
        targetFiles.length === 1
          ? "Media file deleted successfully"
          : `${targetFiles.length} media files deleted successfully`,
      );
    }
  };

  const options: MUIDataTableOptions = {
    filterType: "checkbox",
    download: false,
    print: false,
    rowsSelected: rowsSelected,
    onRowSelectionChange: (_currentRowsSelected, allRowsSelected) => {
      setRowsSelected(allRowsSelected.map((row) => row.dataIndex));
    },
    onRowsDelete: (rowsDeleted) => {
      const idsToDelete: number[] = rowsDeleted.data
        .map((d) => rows[d.dataIndex][5])
        .filter((id): id is number => id !== undefined && id !== null);
      if (idsToDelete.length > 0) {
        setFilesToDelete(idsToDelete);
        setDeleteConfirmOpen(true);
      }
      return false;
    },
  };
  const excludeDisplay: MUIDataTableColumnOptions = {
    display: "excluded",
    filter: false,
  };
  const includeDisplay: MUIDataTableColumnOptions = {
    display: "true",
  };
  const isTV = mediaType === "tv" || mediaType === "tvshow";
  const columns = [
    "Title",
    {
      name: "Season",
      options: isTV ? includeDisplay : excludeDisplay,
    },
    {
      name: "Episode",
      options: isTV ? includeDisplay : excludeDisplay,
    },
    "File Origin",
    "File Location",
    { name: "file_id", options: excludeDisplay },
  ];

  return (
    <>
      <Dialog
        onClose={onClose}
        open={open}
        disableScrollLock={false}
        className="video-dialog"
        maxWidth={false}
      >
        {isLoading ? (
          <div className="history-no-data-header">Loading...</div>
        ) : !mediaFiles || mediaFiles?.providers?.[0]?.streams?.length === 0 ? (
          <div className="history-no-data-header">No Media Files.</div>
        ) : (
          <MUIDataTable
            title={"Your Media Files"}
            data={rows}
            columns={columns}
            options={options}
          />
        )}
      </Dialog>

      <Dialog
        open={deleteConfirmOpen}
        onClose={() => setDeleteConfirmOpen(false)}
      >
        <DialogTitle>Confirm Delete Media File</DialogTitle>
        <DialogContent>
          {filesToDelete.length === 1
            ? "Are you sure you want to delete this media file?"
            : `Are you sure you want to delete these ${filesToDelete.length} media files?`}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteConfirmOpen(false)}>Cancel</Button>
          <Button onClick={handleConfirmDelete} color="error" autoFocus>
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
