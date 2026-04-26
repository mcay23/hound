import {
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  InputAdornment,
  TextField,
  Typography,
} from "@mui/material";
import {
  useCreateApiKeyMutation,
  useDeleteApiKeyMutation,
  useApiKeys,
} from "../../api/hooks/api_keys";
import { useState } from "react";
import toast from "react-hot-toast";
import { ContentCopy, Visibility, VisibilityOff } from "@mui/icons-material";
import { copyToClipboard } from "../../helpers/helpers";

export default function ApiKeys() {
  const { data, isLoading, error } = useApiKeys();
  const [isAddUserDialogOpen, setIsAddUserDialogOpen] = useState(false);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="w-100">
      <h2>API Keys</h2>
      <hr />
      <p className="settings-content-text">
        Add an API Key for third-party applications. Each application should
        have its own API Key so you can manage access easily.
      </p>
      <p>WARNING: Don't share this to applications you don't trust!</p>
      <Button
        className="mt-2"
        variant="contained"
        size="small"
        onClick={() => {
          setIsAddUserDialogOpen(true);
        }}
      >
        Add API Key
      </Button>
      {data?.map((apiKey: any) => (
        <ApiKeyCard key={apiKey.key_id} apiKey={apiKey} />
      ))}
      <AddApiKeyModal
        open={isAddUserDialogOpen}
        onClose={() => setIsAddUserDialogOpen(false)}
      />
    </div>
  );
}

function ApiKeyCard({ apiKey }: { apiKey: any }) {
  const deleteApiKeyMutation = useDeleteApiKeyMutation();
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

  const handleDeleteApiKey = () => {
    deleteApiKeyMutation.mutate(apiKey.key_id, {
      onSuccess: () => {
        toast.success("API Key deleted successfully");
        setIsDeleteModalOpen(false);
      },
      onError: (error) => {
        toast.error("Error deleting API key: " + error.message);
      },
    });
  };

  return (
    <Card variant="outlined" key={apiKey.key_id} className="mt-3">
      <CardContent>
        <Typography variant="h6">{apiKey.name}</Typography>
        <Typography variant="body2" color="text.secondary">
          Created At: {new Date(apiKey.created_at).toLocaleString()}
        </Typography>
        <ApiKeyField apiKey={apiKey.api_key} />
        <div className="d-flex flex-row mt-2">
          <Button
            variant="outlined"
            size="small"
            color="error"
            onClick={() => setIsDeleteModalOpen(true)}
          >
            Revoke Key
          </Button>
        </div>
        <ConfirmDeleteApiKeyModal
          open={isDeleteModalOpen}
          onClose={() => setIsDeleteModalOpen(false)}
          onConfirm={handleDeleteApiKey}
          name={apiKey.name}
        />
      </CardContent>
    </Card>
  );
}

function ApiKeyField({ apiKey }: { apiKey: string }) {
  const [show, setShow] = useState(false);
  return (
    <TextField
      label="API Key"
      value={show ? apiKey : "••••••••••••••"}
      size="small"
      className="mt-3 mb-3"
      InputProps={{
        readOnly: true,
        endAdornment: (
          <InputAdornment position="end">
            <IconButton
              onClick={() => setShow((s) => !s)}
              onMouseDown={(e) => e.preventDefault()}
              edge="end"
            >
              {show ? <VisibilityOff /> : <Visibility />}
            </IconButton>
            <IconButton
              className="ms-2"
              onClick={async () => {
                try {
                  await copyToClipboard(apiKey);
                  toast.success("Copied to clipboard");
                } catch (err) {
                  toast.error("Failed to copy: " + err);
                }
              }}
            >
              <ContentCopy />
            </IconButton>
          </InputAdornment>
        ),
      }}
    />
  );
}

function ConfirmDeleteApiKeyModal({
  open,
  onClose,
  onConfirm,
  name,
}: {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  name: string;
}) {
  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Confirm Delete API Key</DialogTitle>
      <DialogContent>
        Are you sure you want to delete the API key "{name}"? This action cannot
        be undone.
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={onConfirm} color="error">
          Delete
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function AddApiKeyModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const createApiKeyMutation = useCreateApiKeyMutation();
  const [name, setName] = useState("");

  const handleClose = () => {
    setName("");
    onClose();
  };

  const handleCreateApiKey = () => {
    if (!name.trim()) {
      toast.error("Please enter a name for the API key");
      return;
    }
    createApiKeyMutation.mutate(name, {
      onSuccess: () => {
        handleClose();
        toast.success("API Key created successfully");
      },
      onError: (error) => {
        toast.error("Error creating API key: " + error.message);
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Create New API Key</DialogTitle>
      <DialogContent>
        <TextField
          label="Name"
          variant="outlined"
          fullWidth
          required
          margin="normal"
          value={name}
          onChange={(e) => setName(e.target.value)}
          autoFocus
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleCreateApiKey}>Create</Button>
      </DialogActions>
    </Dialog>
  );
}
