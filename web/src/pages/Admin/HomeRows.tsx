import { useEffect, useMemo, useState } from "react";
import toast from "react-hot-toast";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";

import {
  useAvailableCatalogs,
  useDefaultHomeRows,
  useUpdateDefaultHomeRowsMutation,
} from "../../api/hooks/home";
import "./HomeRows.css";

type Catalog = {
  catalog_title: string;
  catalog_source: string;
  catalog_id: string;
};

type HomeRow = {
  title: string;
  catalog_selection: "all" | "rotate";
  item_order: "default" | "random";
  catalogs: Catalog[];
};

type AvailableCatalog = {
  catalog_title: string;
  catalog_source: string;
  catalog_id: string;
  catalog_type: string;
  media_type?: string;
  description: string;
};

const catalogSelectionOptions = [
  { value: "all", label: "Show All Catalogs" },
  { value: "rotate", label: "Rotate Between Catalogs" },
];

const itemOrderOptions = [
  { value: "default", label: "Default" },
  { value: "random", label: "Randomize" },
];

export default function HomeRows() {
  const {
    data: defaultHomeRows,
    isLoading: isHomeRowsLoading,
    error: homeRowsError,
  } = useDefaultHomeRows();
  const {
    data: catalogDefinitionsResponse,
    isLoading: isCatalogsLoading,
    error: catalogsError,
  } = useAvailableCatalogs();
  const updateDefaultHomeRows = useUpdateDefaultHomeRowsMutation();
  const [homeRows, setHomeRows] = useState<HomeRow[]>([]);
  const [isDirty, setIsDirty] = useState(false);
  const selectableCatalogs = catalogDefinitionsResponse?.catalogs ?? [];

  useEffect(() => {
    if (!defaultHomeRows?.home_rows || isDirty) {
      return;
    }
    setHomeRows(defaultHomeRows.home_rows);
  }, [defaultHomeRows, isDirty]);

  const updateRows = (nextRows: HomeRow[]) => {
    setHomeRows(nextRows);
    setIsDirty(true);
  };

  const addHomeRow = () => {
    const firstCatalog = selectableCatalogs[0];
    updateRows([
      ...homeRows,
      {
        title: "New Home Row",
        catalog_selection: "all",
        item_order: "default",
        catalogs: firstCatalog
          ? [createCatalogFromAvailable(firstCatalog)]
          : [],
      },
    ]);
  };

  const saveHomeRows = () => {
    if (homeRows.length === 0) {
      toast.error("Add at least one home row");
      return;
    }
    for (const row of homeRows) {
      if (!row.title.trim()) {
        toast.error("Every home row needs a title");
        return;
      }
      if (row.catalogs.length === 0) {
        toast.error(`"${row.title}" needs at least one catalog`);
        return;
      }
      for (const catalog of row.catalogs) {
        if (!catalog.catalog_title.trim()) {
          toast.error(`Every catalog in "${row.title}" needs a title`);
          return;
        }
      }
    }

    updateDefaultHomeRows.mutate(
      {
        user_id: -1,
        home_rows: homeRows,
      },
      {
        onSuccess: () => {
          toast.success("Default home catalogs updated");
          setIsDirty(false);
        },
        onError: (error: any) => {
          toast.error("Failed to save default home catalogs: " + error.message);
        },
      },
    );
  };

  const resetHomeRows = () => {
    setHomeRows(defaultHomeRows?.home_rows ?? []);
    setIsDirty(false);
  };

  if (isHomeRowsLoading || isCatalogsLoading) {
    return <div>Loading...</div>;
  }

  if (homeRowsError) {
    return <div>Error: {homeRowsError.message}</div>;
  }

  if (catalogsError) {
    return <div>Error: {catalogsError.message}</div>;
  }

  return (
    <div className="w-100">
      <h2>Default Home Catalogs</h2>
      <hr />
      <p className="w-50">
        Change the default home catalogs for all users here. Users can customize
        their own home catalogs from their Account Settings page. Each home row
        can have several catalogs. You can mix between different catalogs or
        rotate through them, so that only one catalog is shown at a time.
        <br />
        <br />
        If you choose Catalog Selection: rotate, the catalog title will be shown
        instead of the home row Title.
      </p>
      <div className="home-rows-content-container">
        {!catalogDefinitionsResponse?.mdblist_configured && (
          <Alert severity="warning" className="mt-3 mb-3">
            An MDBList API Key must be configured to use MDBList catalogs
          </Alert>
        )}
        <Stack spacing={2}>
          {homeRows.map((row, rowIndex) => (
            <>
              <HomeRowCard
                key={rowIndex}
                row={row}
                rowIndex={rowIndex}
                rows={homeRows}
                catalogs={selectableCatalogs}
                isMDBListConfigured={
                  catalogDefinitionsResponse?.mdblist_configured
                }
                onChange={updateRows}
              />
            </>
          ))}
        </Stack>
        {isDirty && (
          <Alert severity="info" className="mt-3">
            You have unsaved changes.
          </Alert>
        )}
        {homeRows.length === 0 && (
          <Alert severity="warning" className="mt-3">
            No home rows are configured.
          </Alert>
        )}
        <Stack
          direction="row"
          spacing={1}
          className="mt-3"
          justifyContent="flex-end"
        >
          <Button
            variant="outlined"
            size="small"
            onClick={resetHomeRows}
            disabled={!isDirty}
          >
            Reset
          </Button>
          <Button
            variant="outlined"
            size="small"
            onClick={saveHomeRows}
            disabled={updateDefaultHomeRows.isPending || homeRows.length === 0}
          >
            Save Changes
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={addHomeRow}
            disabled={selectableCatalogs.length === 0}
          >
            Add Home Row
          </Button>
        </Stack>
      </div>
    </div>
  );
}

function HomeRowCard({
  row,
  rowIndex,
  rows,
  catalogs,
  isMDBListConfigured,
  onChange,
}: {
  row: HomeRow;
  rowIndex: number;
  rows: HomeRow[];
  catalogs: AvailableCatalog[];
  isMDBListConfigured: boolean;
  onChange: (rows: HomeRow[]) => void;
}) {
  const [isMDBListDialogOpen, setIsMDBListDialogOpen] = useState(false);

  const updateRow = (nextRow: HomeRow) => {
    const nextRows = [...rows];
    nextRows[rowIndex] = nextRow;
    onChange(nextRows);
  };

  const addCatalog = () => {
    if (catalogs.length === 0) {
      toast.error("No catalogs are available");
      return;
    }
    updateRow({
      ...row,
      catalogs: [...row.catalogs, createCatalogFromAvailable(catalogs[0])],
    });
  };

  const addMDBListCatalog = (catalogID: string) => {
    const catalogTitle = createMDBListCatalogTitle(catalogID);
    updateRow({
      ...row,
      catalogs: [
        ...row.catalogs,
        {
          catalog_title: catalogTitle,
          catalog_source: "mdblist",
          catalog_id: catalogID,
        },
      ],
    });
    setIsMDBListDialogOpen(false);
  };

  return (
    <Card variant="outlined" sx={{ boxShadow: 2 }}>
      <CardContent>
        <Stack spacing={2}>
          <Stack direction="row" justifyContent="space-between" spacing={1}>
            <h4>
              {rowIndex + 1 + " - "}
              {row.title ? row.title : "New Home Row"}
            </h4>
            <Stack direction="row" spacing={1}>
              <Button
                variant="outlined"
                size="small"
                onClick={() => onChange(moveItem(rows, rowIndex, rowIndex - 1))}
                disabled={rowIndex === 0}
              >
                Up
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={() => onChange(moveItem(rows, rowIndex, rowIndex + 1))}
                disabled={rowIndex === rows.length - 1}
              >
                Down
              </Button>
              <Button
                variant="outlined"
                size="small"
                color="error"
                onClick={() => onChange(rows.filter((_, i) => i !== rowIndex))}
              >
                Remove
              </Button>
            </Stack>
          </Stack>

          <TextField
            label="Home Row Title"
            value={row.title}
            onChange={(event) =>
              updateRow({ ...row, title: event.target.value })
            }
            className="mb-2"
            size="small"
            fullWidth
          />

          <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
            <FormControl fullWidth size="small">
              <InputLabel>Catalog Selection</InputLabel>
              <Select
                label="Catalog Selection"
                value={row.catalog_selection}
                onChange={(event) =>
                  updateRow({
                    ...row,
                    catalog_selection: event.target
                      .value as HomeRow["catalog_selection"],
                  })
                }
              >
                {catalogSelectionOptions.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControl fullWidth size="small">
              <InputLabel>Item Order</InputLabel>
              <Select
                label="Item Order"
                value={row.item_order}
                onChange={(event) =>
                  updateRow({
                    ...row,
                    item_order: event.target.value as HomeRow["item_order"],
                  })
                }
              >
                {itemOrderOptions.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Stack>
          <Divider />
          <Stack direction="row" justifyContent="space-between">
            <h5>Catalogs</h5>
            <Stack direction="row" spacing={1}>
              <Button
                variant="outlined"
                size="small"
                onClick={addCatalog}
                disabled={catalogs.length === 0}
              >
                Add Catalog
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={() => setIsMDBListDialogOpen(true)}
                disabled={!isMDBListConfigured}
              >
                Add MDBList Catalog
              </Button>
            </Stack>
          </Stack>

          <Stack spacing={1.5}>
            {row.catalogs.map((catalog, catalogIndex) => (
              <CatalogEditor
                key={`${catalog.catalog_source}-${catalog.catalog_id}-${catalogIndex}`}
                catalog={catalog}
                catalogIndex={catalogIndex}
                row={row}
                availableCatalogs={catalogs}
                onChange={updateRow}
              />
            ))}
          </Stack>
        </Stack>
      </CardContent>
      <AddMDBListCatalogDialog
        open={isMDBListDialogOpen}
        onClose={() => setIsMDBListDialogOpen(false)}
        onAdd={addMDBListCatalog}
      />
    </Card>
  );
}

function CatalogEditor({
  catalog,
  catalogIndex,
  row,
  availableCatalogs,
  onChange,
}: {
  catalog: Catalog;
  catalogIndex: number;
  row: HomeRow;
  availableCatalogs: AvailableCatalog[];
  onChange: (row: HomeRow) => void;
}) {
  const selectedCatalogKey = getCatalogKey(catalog);
  const isMDBListCatalog = catalog.catalog_source === "mdblist";
  const [mdbListURL, setMDBListURL] = useState(
    getMDBListURL(catalog.catalog_id),
  );

  const updateCatalog = (nextCatalog: Catalog) => {
    const nextCatalogs = [...row.catalogs];
    nextCatalogs[catalogIndex] = nextCatalog;
    onChange({ ...row, catalogs: nextCatalogs });
  };

  useEffect(() => {
    if (isMDBListCatalog) {
      setMDBListURL(getMDBListURL(catalog.catalog_id));
    }
  }, [catalog.catalog_id, isMDBListCatalog]);

  const selectedAvailableCatalog = availableCatalogs.find(
    (availableCatalog) =>
      getCatalogKey(availableCatalog) === selectedCatalogKey,
  );
  const isCatalogSelectable = Boolean(selectedAvailableCatalog);

  return (
    <Box
      sx={{
        border: "1px solid #ddd",
        borderRadius: 1,
        p: 2,
        backgroundColor: "#f8f8f8",
      }}
    >
      <Stack spacing={1.5}>
        <Stack direction="row" justifyContent="space-between" spacing={1}>
          <Typography variant="subtitle2">
            Catalog {catalogIndex + 1}
          </Typography>
          <Stack direction="row" spacing={1}>
            <Button
              variant="outlined"
              size="small"
              onClick={() =>
                onChange({
                  ...row,
                  catalogs: moveItem(
                    row.catalogs,
                    catalogIndex,
                    catalogIndex - 1,
                  ),
                })
              }
              disabled={catalogIndex === 0}
            >
              Up
            </Button>
            <Button
              variant="outlined"
              size="small"
              onClick={() =>
                onChange({
                  ...row,
                  catalogs: moveItem(
                    row.catalogs,
                    catalogIndex,
                    catalogIndex + 1,
                  ),
                })
              }
              disabled={catalogIndex === row.catalogs.length - 1}
            >
              Down
            </Button>
            <Button
              variant="outlined"
              size="small"
              color="error"
              onClick={() =>
                onChange({
                  ...row,
                  catalogs: row.catalogs.filter((_, i) => i !== catalogIndex),
                })
              }
            >
              Remove
            </Button>
          </Stack>
        </Stack>

        <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
          <TextField
            label="Catalog Title"
            value={catalog.catalog_title}
            onChange={(event) =>
              updateCatalog({
                ...catalog,
                catalog_title: event.target.value,
              })
            }
            size="small"
            fullWidth
          />
          {isMDBListCatalog ? (
            <TextField
              label="MDBList URL"
              value={mdbListURL}
              onChange={(event) => {
                const nextURL = event.target.value;
                const nextCatalogID = parseMDBListCatalogID(nextURL);
                setMDBListURL(nextURL);
                if (nextCatalogID) {
                  updateCatalog({
                    ...catalog,
                    catalog_id: nextCatalogID,
                  });
                }
              }}
              onBlur={() => {
                const nextCatalogID = parseMDBListCatalogID(mdbListURL);
                if (!nextCatalogID) {
                  toast.error("Enter a valid MDBList list URL");
                  setMDBListURL(getMDBListURL(catalog.catalog_id));
                }
              }}
              size="small"
              fullWidth
            />
          ) : (
            <FormControl fullWidth size="small">
              <InputLabel>Catalog</InputLabel>
              <Select
                label="Catalog"
                value={selectedCatalogKey}
                onChange={(event) => {
                  const nextAvailableCatalog = availableCatalogs.find(
                    (availableCatalog) =>
                      getCatalogKey(availableCatalog) === event.target.value,
                  );
                  if (!nextAvailableCatalog) {
                    return;
                  }
                  updateCatalog(
                    createCatalogFromAvailable(nextAvailableCatalog),
                  );
                }}
              >
                {!isCatalogSelectable && (
                  <MenuItem value={selectedCatalogKey} disabled>
                    {catalog.catalog_title || catalog.catalog_id}
                  </MenuItem>
                )}
                {availableCatalogs.map((availableCatalog) => (
                  <MenuItem
                    key={getCatalogKey(availableCatalog)}
                    value={getCatalogKey(availableCatalog)}
                  >
                    {availableCatalog.catalog_title}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          )}
        </Stack>

        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <Chip size="small" label={catalog.catalog_source} />
          <Chip
            size="small"
            label={catalog.catalog_id}
            clickable={catalog.catalog_source === "mdblist"}
            onClick={
              catalog.catalog_source === "mdblist"
                ? () => window.open(getMDBListURL(catalog.catalog_id), "_blank")
                : undefined
            }
          />
          {selectedAvailableCatalog?.catalog_type && (
            <Chip size="small" label={selectedAvailableCatalog.catalog_type} />
          )}
          {selectedAvailableCatalog?.media_type && (
            <Chip size="small" label={selectedAvailableCatalog.media_type} />
          )}
        </Stack>
        {selectedAvailableCatalog?.description && (
          <Typography variant="body2" color="text.secondary">
            {selectedAvailableCatalog.description}
          </Typography>
        )}
      </Stack>
    </Box>
  );
}

function AddMDBListCatalogDialog({
  open,
  onClose,
  onAdd,
}: {
  open: boolean;
  onClose: () => void;
  onAdd: (catalogID: string) => void;
}) {
  const [listURL, setListURL] = useState("");

  const handleClose = () => {
    setListURL("");
    onClose();
  };

  const handleAdd = () => {
    const catalogID = parseMDBListCatalogID(listURL);
    if (!catalogID) {
      toast.error("Enter a valid MDBList list URL");
      return;
    }
    onAdd(catalogID);
    setListURL("");
  };

  return (
    <Dialog open={open} onClose={handleClose} fullWidth maxWidth="sm">
      <DialogTitle>Add MDBList Catalog</DialogTitle>
      <DialogContent>
        <p>
          Paste your MDBList list URL here
          <br />
          eg. https://mdblist.com/lists/garycrawfordgc/latest-tv-shows
        </p>
        <TextField
          label="MDBList URL"
          value={listURL}
          onChange={(event) => setListURL(event.target.value)}
          placeholder="https://mdblist.com/lists/author/list-name"
          margin="normal"
          fullWidth
          autoFocus
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button onClick={handleAdd}>Add</Button>
      </DialogActions>
    </Dialog>
  );
}

function createCatalogFromAvailable(catalog: AvailableCatalog): Catalog {
  return {
    catalog_title: catalog.catalog_title,
    catalog_source: catalog.catalog_source,
    catalog_id: catalog.catalog_id,
  };
}

function getCatalogKey(
  catalog: Pick<Catalog, "catalog_source" | "catalog_id">,
) {
  return `${catalog.catalog_source}:${catalog.catalog_id}`;
}

function parseMDBListCatalogID(value: string) {
  const trimmed = value.trim();
  if (!trimmed) {
    return "";
  }
  try {
    const url = new URL(trimmed);
    if (url.hostname !== "mdblist.com" && url.hostname !== "www.mdblist.com") {
      return "";
    }
    const parts = url.pathname.split("/").filter(Boolean);
    if (parts.length < 3 || parts[0] !== "lists") {
      return "";
    }
    return `${parts[1]}/${parts[2]}`;
  } catch {
    const parts = trimmed.split("/").filter(Boolean);
    if (parts.length === 2) {
      return `${parts[0]}/${parts[1]}`;
    }
    return "";
  }
}

function getMDBListURL(catalogID: string) {
  return `https://mdblist.com/lists/${catalogID}`;
}

function createMDBListCatalogTitle(catalogID: string) {
  const slug = catalogID.split("/")[1] ?? catalogID;
  return slug
    .split("-")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function moveItem<T>(items: T[], fromIndex: number, toIndex: number) {
  if (toIndex < 0 || toIndex >= items.length) {
    return items;
  }
  const nextItems = [...items];
  const [item] = nextItems.splice(fromIndex, 1);
  nextItems.splice(toIndex, 0, item);
  return nextItems;
}
