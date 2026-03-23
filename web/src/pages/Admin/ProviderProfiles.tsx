import "./ProviderProfiles.css";
import {
  useCreateProviderProfile,
  useProviderProfiles,
} from "../../api/hooks/providerProfiles";
import {
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Modal,
  TextField,
} from "@mui/material";
import { useDeleteProviderProfile } from "../../api/hooks/providerProfiles";
import toast from "react-hot-toast";
import { useState } from "react";

export default function ProviderProfiles() {
  const { data: providerProfiles, isLoading: isProviderProfilesLoading } =
    useProviderProfiles();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isAddProviderDialogOpen, setIsAddProviderDialogOpen] = useState(false);
  const [selectedID, setSelectedID] = useState(-1);
  const deleteProviderProfile = useDeleteProviderProfile();
  return (
    <>
      <div>
        <h2>Provider Profiles</h2>
        <hr />
        <p className="provider-profile-text">
          Add a provider to start streaming and downloading. Multiple profiles
          are useful if you want different presets for streaming and downloading
          (eg. prioritize performance for streaming, quality for downloads,
          etc.)
        </p>
        {providerProfiles?.length === 0 && (
          <div className="text-muted">
            No provider profiles yet, add at least one to start streaming.
          </div>
        )}
        <Button
          className="mt-3"
          onClick={() => setIsAddProviderDialogOpen(true)}
          variant="outlined"
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
          {"Delete this review?"}
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            This action cannot be reversed.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsDeleteDialogOpen(false)}>Cancel</Button>
          <Button
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
  return (
    <Card
      variant="outlined"
      key={profile.provider_profile_id}
      className="mt-3 provider-profile-card"
    >
      <CardContent className="provider-profile-card-content">
        <h5>{profile.name}</h5>
        <div className="text-muted">{profile.manifest_url}</div>
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
  const addProviderProfile = useCreateProviderProfile();
  return (
    <Dialog
      open={open}
      onClose={() => {
        setName("");
        setManifestURL("");
        setOpen(false);
      }}
    >
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
        <Button onClick={() => setOpen(false)}>Cancel</Button>
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
                  setOpen(false);
                },
                onError: () => {
                  toast.error("Failed to add provider profile");
                  setOpen(false);
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
