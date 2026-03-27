import "./ProviderProfiles.css";
import {
  useCreateProviderProfileMutation,
  useProviderProfiles,
  useUpdateProviderProfileMutation,
} from "../../api/hooks/providerProfiles";
import {
  Button,
  Card,
  CardContent,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Modal,
  TextField,
} from "@mui/material";
import { useDeleteProviderProfileMutation } from "../../api/hooks/providerProfiles";
import toast from "react-hot-toast";
import { useState } from "react";

export default function ProviderProfiles() {
  const { data: providerProfiles, isLoading: isProviderProfilesLoading } =
    useProviderProfiles();
  const deleteProviderProfile = useDeleteProviderProfileMutation();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isAddProviderDialogOpen, setIsAddProviderDialogOpen] = useState(false);
  const [selectedID, setSelectedID] = useState(-1);
  return (
    <>
      <div>
        <h2>Provider Profiles</h2>
        <hr />
        <p className="provider-profile-text">
          Add a provider to start streaming and downloading. Multiple profiles
          are useful if you want different presets for streaming and downloading
          (eg. prioritize speed/compatibility for streaming, and quality for
          downloads).
        </p>
        <p className="provider-profile-text">
          You can also set the global default profile for all users for
          streaming/downloading.
        </p>
        <p className="provider-profile-text">
          For help setting up a provider, visit the docs.
        </p>
        {providerProfiles?.length === 0 && (
          <div className="text-muted">
            No provider profiles yet, add at least one to start streaming.
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
        {isProviderProfilesLoading ? (
          <div>Loading...</div>
        ) : (
          providerProfiles?.map((profile: any) => {
            return (
              <ProviderProfile
                key={profile.provider_profile_id}
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
              deleteProviderProfile.mutate(selectedID, {
                onSuccess: () => {
                  toast.success("Provider profile deleted");
                  setIsDeleteDialogOpen(false);
                },
                onError: () => {
                  toast.error("Failed to delete provider profile");
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

function ProviderProfile({
  profile,
  setSelectedID,
  setIsDeleteDialogOpen,
}: {
  profile: any;
  setSelectedID: (id: number) => void;
  setIsDeleteDialogOpen: (open: boolean) => void;
}) {
  const updateProviderProfile = useUpdateProviderProfileMutation();
  return (
    <Card
      variant="outlined"
      key={profile.provider_profile_id}
      className="mt-3 provider-profile-card"
    >
      <CardContent className="provider-profile-card-content">
        <h5>{profile.name}</h5>
        <div className="text-muted">{profile.manifest_url}</div>
        <div className="d-flex flex-row">
          <Button
            className="mt-2"
            variant="outlined"
            size="small"
            onClick={() => {
              setSelectedID(profile.provider_profile_id);
              setIsDeleteDialogOpen(true);
            }}
          >
            Delete
          </Button>
          {!profile.is_default_streaming && (
            <Button
              className="ms-2 mt-2"
              variant="outlined"
              size="small"
              onClick={() => {
                updateProviderProfile.mutate({
                  id: profile.provider_profile_id,
                  isDefaultStreaming: true,
                });
              }}
            >
              Set as Default for Streaming
            </Button>
          )}
          {!profile.is_default_downloading && (
            <Button
              className="ms-2 mt-2"
              variant="outlined"
              size="small"
              onClick={() => {
                updateProviderProfile.mutate({
                  id: profile.provider_profile_id,
                  isDefaultDownloading: true,
                });
              }}
            >
              Set as Default for Downloading
            </Button>
          )}
        </div>
        {(profile.is_default_streaming || profile.is_default_downloading) && (
          <div className="d-flex flex-row mt-3">
            {profile.is_default_streaming && (
              <Chip className="me-2" label="Default for Streaming" />
            )}
            {profile.is_default_downloading && (
              <Chip className="me-2" label="Default for Downloading" />
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function parseManifestURL(manifestURL: string) {
  const url = new URL(manifestURL);
  return url.origin;
}

function AddProviderModal({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) {
  const [name, setName] = useState("");
  const [manifestURL, setManifestURL] = useState("");
  const addProviderProfile = useCreateProviderProfileMutation();
  const handleClose = () => {
    setName("");
    setManifestURL("");
    setOpen(false);
  };
  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Add Provider</DialogTitle>
      <DialogContent className="provider-profile-container">
        <hr />
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
          label="Manifest URL"
          variant="outlined"
          fullWidth
          required
          margin="normal"
          value={manifestURL}
          onChange={(e) => setManifestURL(e.target.value)}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button
          onClick={() => {
            if (name === "" || manifestURL === "") {
              toast.error("Please fill in all fields");
              return;
            }
            try {
              new URL(manifestURL);
            } catch (error) {
              toast.error(
                "Please enter a valid URL (including http:// and https://)",
              );
              return;
            }
            addProviderProfile.mutate(
              {
                name,
                manifestURL,
              },
              {
                onSuccess: () => {
                  toast.success("Provider profile added");
                  handleClose();
                },
                onError: () => {
                  toast.error("Failed to add provider profile");
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
