import { useMemo, useState } from "react";
import toast from "react-hot-toast";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
} from "@mui/material";

import {
  useAvailableCatalogs,
  useResetUserHomeRowsMutation,
  useUpdateUserHomeRowsMutation,
  useUserHomeRows,
} from "../../api/hooks/home";
import {
  useCollections,
  usePublicCollections,
} from "../../api/hooks/collections";
import { HomeRowsEditor } from "../Admin/HomeRows";

type CollectionItem = {
  collection_id: number;
  collection_title: string;
  owner_display_name?: string;
  owner_username?: string;
};

export default function UserHomeRows() {
  const {
    data: userHomeRows,
    isLoading: isUserHomeRowsLoading,
    error: userHomeRowsError,
  } = useUserHomeRows();
  const {
    data: catalogDefinitionsResponse,
    isLoading: isCatalogsLoading,
    error: catalogsError,
  } = useAvailableCatalogs();
  const { data: userCollections = [] } = useCollections();
  const { data: publicCollections = [] } = usePublicCollections();
  const updateUserHomeRows = useUpdateUserHomeRowsMutation();
  const resetUserHomeRows = useResetUserHomeRowsMutation();
  const [isResetDialogOpen, setIsResetDialogOpen] = useState(false);
  const [editorKey, setEditorKey] = useState(0);

  // dedupe user + public collections
  const combinedCollections = useMemo<CollectionItem[]>(() => {
    const seen = new Set<number>();
    const result: CollectionItem[] = [];
    for (const c of [
      ...(userCollections as CollectionItem[]),
      ...(publicCollections as CollectionItem[]),
    ]) {
      if (!seen.has(c.collection_id)) {
        seen.add(c.collection_id);
        result.push(c);
      }
    }
    return result;
  }, [userCollections, publicCollections]);

  const handleResetToDefaults = () => {
    resetUserHomeRows.mutate(undefined, {
      onSuccess: () => {
        toast.success("Home layout reset to defaults");
        setIsResetDialogOpen(false);
        setEditorKey((key) => key + 1);
      },
      onError: (error: any) => {
        toast.error("Failed to reset home layout: " + error.message);
      },
    });
  };

  return (
    <>
      <HomeRowsEditor
        key={editorKey}
        title="Home Layout"
        description={
          <>
            Customize the rows and catalogs shown on your home page. Each home
            row can have several catalogs. You can mix between different
            catalogs or rotate through them, so that only one catalog is shown
            at a time.
            <br />
            <br />
            If you choose Catalog Selection: rotate, the catalog title will be
            shown instead of the home row title.
          </>
        }
        homeRowsData={userHomeRows}
        catalogDefinitionsResponse={catalogDefinitionsResponse}
        collections={combinedCollections}
        isLoading={isUserHomeRowsLoading || isCatalogsLoading}
        error={userHomeRowsError ?? catalogsError}
        onSave={(homeRows) =>
          updateUserHomeRows.mutateAsync({
            home_rows: homeRows,
          })
        }
        isSaving={updateUserHomeRows.isPending}
        footerActions={
          <Button
            variant="outlined"
            size="small"
            color="error"
            onClick={() => setIsResetDialogOpen(true)}
            disabled={
              updateUserHomeRows.isPending || resetUserHomeRows.isPending
            }
          >
            Reset to Defaults
          </Button>
        }
      />
      <Dialog
        open={isResetDialogOpen}
        onClose={() => setIsResetDialogOpen(false)}
      >
        <DialogTitle>Reset home layout?</DialogTitle>
        <DialogContent>
          This will remove your personal home layout. Your home page will use
          the server default layout and continue following future default
          changes.
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsResetDialogOpen(false)}>Cancel</Button>
          <Button color="error" onClick={handleResetToDefaults}>
            Reset
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
