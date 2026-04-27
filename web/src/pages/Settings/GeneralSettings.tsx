import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from "@mui/material";
import { useChangePassword } from "../../api/hooks/auth";
import { useState } from "react";
import toast from "react-hot-toast";
import { useNavigate } from "react-router-dom";

export default function GeneralSettings() {
  const [isChangePasswordModalOpen, setIsChangePasswordModalOpen] =
    useState(false);
  return (
    <div className="w-100">
      <h2>General Settings</h2>
      <hr />
      <h4>Password</h4>
      <Button
        className="mt-3"
        onClick={() => {
          toast.error("Can't do this in the demo");
        }}
        variant="contained"
        size="small"
      >
        Change Password
      </Button>
      {/* <ChangePasswordModal
        open={isChangePasswordModalOpen}
        setOpen={setIsChangePasswordModalOpen}
      /> */}
    </div>
  );
}

function ChangePasswordModal({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) {
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const changePasswordMutation = useChangePassword();
  const handleClose = () => {
    setOldPassword("");
    setNewPassword("");
    setOpen(false);
  };
  const navigate = useNavigate();
  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Change Password</DialogTitle>
      <DialogContent className="provider-profile-container">
        <hr />
        <Alert severity="warning">
          Warning: This will log you out of all devices!
        </Alert>
        <TextField
          label="Old Password"
          variant="outlined"
          fullWidth
          required
          type="password"
          margin="normal"
          value={oldPassword}
          onChange={(e) => setOldPassword(e.target.value)}
        />
        <TextField
          label="New Password"
          variant="outlined"
          fullWidth
          required
          type="password"
          margin="normal"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
        />
        <TextField
          label="Confirm New Password"
          variant="outlined"
          fullWidth
          required
          type="password"
          margin="normal"
          value={confirmNewPassword}
          onChange={(e) => setConfirmNewPassword(e.target.value)}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button
          onClick={() => {
            if (
              oldPassword === "" ||
              newPassword === "" ||
              confirmNewPassword == ""
            ) {
              toast.error("Please fill in all fields");
              return;
            }
            if (newPassword !== confirmNewPassword) {
              toast.error("Passwords don't match!");
              return;
            }
            if (newPassword.length < 8) {
              toast.error("Password too short");
              return;
            }
            changePasswordMutation.mutate(
              {
                oldPassword,
                newPassword,
              },
              {
                onSuccess: () => {
                  toast.success("Password changed, logging out");
                  setTimeout(() => {
                    navigate("/logout");
                  }, 3500);
                  handleClose();
                },
                onError: (error: any) => {
                  const status = error?.response?.status;
                  if (status === 400) {
                    toast.error("Incorrect Password");
                  } else {
                    toast.error("Internal error changing password");
                  }
                  handleClose();
                },
              },
            );
          }}
        >
          Change Password
        </Button>
      </DialogActions>
    </Dialog>
  );
}
