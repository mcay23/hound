import "./IPTVProfiles.css";
import {
  Alert,
  Button,
  Card,
  CardContent,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
} from "@mui/material";
import toast from "react-hot-toast";
import { useState } from "react";
import {
  useCreateIPTVProviderMutation,
  useDeleteIPTVProviderMutation,
  useIPTVProviders,
} from "../../api/hooks/live_tv";

export default function IPTVProfiles() {
  const { data: iptvProviders, isLoading: isIPTVProvidersLoading } =
    useIPTVProviders();
  const deleteIPTVProvider = useDeleteIPTVProviderMutation();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isAddProviderDialogOpen, setIsAddProviderDialogOpen] = useState(false);
  const [selectedID, setSelectedID] = useState(-1);
  return (
    <>
      <div>
        <h2>IPTV Profiles (Experimental)</h2>
        <hr />
        <p className="iptv-provider-text">
          You can add IPTV providers to watch Live TV. Xtream and M3U8 playlists
          are supported.
        </p>
        <Alert severity="warning" className="mb-2">
          This feature is considered experimental, and has been tested with only
          a limited number of providers. Currently, links are not proxied
          through Hound, and clients may be able to access your IPTV
          credentials. Many IPTV providers have concurrent connection limits
          which Hound does not enforce.
        </Alert>
        {iptvProviders?.length === 0 && (
          <div className="text-muted">
            No IPTV profile yet, add at least one to start watching Live TV.
          </div>
        )}
        <Button
          className="mt-3"
          onClick={() => setIsAddProviderDialogOpen(true)}
          variant="contained"
          size="small"
        >
          Add Provider
        </Button>
        {isIPTVProvidersLoading ? (
          <div>Loading...</div>
        ) : (
          iptvProviders?.map((profile: any) => {
            return (
              <IPTVProviderCard
                key={profile.iptv_provider_id}
                profile={profile}
                setIsDeleteDialogOpen={setIsDeleteDialogOpen}
                setSelectedID={setSelectedID}
              />
            );
          })
        )}
      </div>
      <AddProviderModal
        open={isAddProviderDialogOpen}
        setOpen={setIsAddProviderDialogOpen}
      />
      <Dialog
        open={isDeleteDialogOpen}
        onClose={() => setIsDeleteDialogOpen(false)}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">
          {"Delete this profile?"}
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            This action cannot be reversed.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsDeleteDialogOpen(false)}>Cancel</Button>
          <Button
            color="error"
            onClick={() => {
              deleteIPTVProvider.mutate(selectedID, {
                onSuccess: () => {
                  toast.success("IPTV provider deleted");
                  setIsDeleteDialogOpen(false);
                },
                onError: () => {
                  toast.error("Failed to delete IPTV provider");
                  setIsDeleteDialogOpen(false);
                },
              });
            }}
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

function IPTVProviderCard({
  profile,
  setSelectedID,
  setIsDeleteDialogOpen,
}: {
  profile: any;
  setSelectedID: (id: number) => void;
  setIsDeleteDialogOpen: (open: boolean) => void;
}) {
  return (
    <Card
      variant="outlined"
      key={profile.iptv_provider_id}
      sx={{ boxShadow: 1 }}
      className="mt-3"
    >
      <CardContent>
        <h5>{profile.name}</h5>
        <div className="text-muted">{profile.host}</div>
        <Chip className="mt-2 mb-2 me-2" label={profile?.iptv_stream_type} />
        <div className="d-flex flex-row">
          <Button
            className="mt-2"
            variant="outlined"
            size="small"
            onClick={() => {
              setSelectedID(profile.iptv_provider_id);
              setIsDeleteDialogOpen(true);
            }}
          >
            Delete
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function AddProviderModal({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) {
  const [iptvStreamType, setIPTVStreamType] = useState("xtream");
  const [name, setName] = useState("");
  const [hostURL, setHostURL] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const addIPTVProvider = useCreateIPTVProviderMutation();
  const handleClose = () => {
    setName("");
    setHostURL("");
    setUsername("");
    setPassword("");
    setOpen(false);
  };
  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Add IPTV Provider</DialogTitle>
      <DialogContent className="provider-profile-container">
        <hr />
        <ToggleButtonGroup
          color="primary"
          value={iptvStreamType}
          exclusive
          onChange={(_, value) => setIPTVStreamType(value)}
          aria-label="Platform"
        >
          <ToggleButton value="xtream">Xtream</ToggleButton>
          <ToggleButton value="m3u8">M3U8 Playlist</ToggleButton>
        </ToggleButtonGroup>
        <TextField
          label="Profile Name"
          variant="outlined"
          fullWidth
          required
          margin="normal"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <TextField
          label="Host URL"
          variant="outlined"
          fullWidth
          required
          margin="normal"
          value={hostURL}
          onChange={(e) => setHostURL(e.target.value)}
        />
        {iptvStreamType === "xtream" && (
          <>
            <TextField
              label="Username"
              variant="outlined"
              fullWidth
              required
              margin="normal"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
            <TextField
              label="Password"
              type="password"
              variant="outlined"
              autoComplete="new-password"
              fullWidth
              required
              margin="normal"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button
          onClick={() => {
            if (
              name === "" ||
              hostURL === "" ||
              username === "" ||
              password === ""
            ) {
              toast.error("Please fill in all fields");
              return;
            }
            try {
              new URL(hostURL);
            } catch (error) {
              toast.error(
                "Please enter a valid URL (including http:// and https://)",
              );
              return;
            }
            addIPTVProvider.mutate(
              {
                iptvStreamType,
                name: name.trim(),
                host: hostURL.trim(),
                username: iptvStreamType === "xtream" ? username : null,
                password: iptvStreamType === "xtream" ? password : null,
              },
              {
                onSuccess: () => {
                  toast.success("IPTV provider added");
                  handleClose();
                },
                onError: () => {
                  toast.error("Failed to add IPTV provider");
                  handleClose();
                },
              },
            );
          }}
        >
          Add
        </Button>
      </DialogActions>
    </Dialog>
  );
}
