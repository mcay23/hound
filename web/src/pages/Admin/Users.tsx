import {
  Alert,
  Button,
  Card,
  CardContent,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  FormGroup,
  TextField,
} from "@mui/material";
import {
  useCreateUserMutation,
  useDeleteUserMutation,
  useResetUserPassword,
  useUsers,
} from "../../api/hooks/users";
import "./Users.css";
import { useState } from "react";
import toast from "react-hot-toast";

export default function UserList() {
  const { data, isLoading, error } = useUsers();
  const [isAddUserDialogOpen, setIsAddUserDialogOpen] = useState(false);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="w-100">
      <h2>Users</h2>
      <hr />
      <div className="d-flex">
        <Alert severity="warning" className="mb-2">
          On release, a paid license will be required for multi-user support.
          This restriction is disabled for all users in the Beta only. Learn
          more at the Hound Docs website.
        </Alert>
      </div>
      <Button
        className="mt-2"
        variant="contained"
        size="small"
        onClick={() => {
          setIsAddUserDialogOpen(true);
        }}
      >
        Add User
      </Button>
      {data.map((user: any) => (
        <UserCard key={user.id} user={user} />
      ))}
      <AddUserModal
        open={isAddUserDialogOpen}
        onClose={() => setIsAddUserDialogOpen(false)}
      />
    </div>
  );
}

function UserCard({ user }: { user: any }) {
  const deleteUserMutation = useDeleteUserMutation();
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isResetModalOpen, setIsResetModalOpen] = useState(false);

  const handleDeleteUser = () => {
    deleteUserMutation.mutate(user.user_id, {
      onSuccess: () => {
        toast.success("User deleted successfully");
        setIsDeleteModalOpen(false);
      },
      onError: (error) => {
        toast.error("Error deleting user: " + error.message);
      },
    });
  };

  return (
    <Card variant="outlined" key={user.user_id} className="mt-3">
      <CardContent>
        <h5>{user.display_name}</h5>
        <h6>Username: {user.username}</h6>
        <div className="d-flex flex-row">
          <Button
            className="mt-2"
            variant="outlined"
            size="small"
            onClick={() => {
              if (user?.is_admin) {
                toast.error(
                  "Please reset admin password through the 'My Account' page",
                );
                return;
              }
              setIsResetModalOpen(true);
            }}
          >
            Reset Password
          </Button>
          <Button
            className="mt-2 ms-2"
            variant="outlined"
            size="small"
            color="error"
            onClick={() => {
              if (user?.is_admin) {
                toast.error("Cannot delete admin user");
                return;
              }
              setIsDeleteModalOpen(true);
            }}
          >
            Delete User
          </Button>
        </div>
        {user.is_admin && <Chip className="mt-2" label="Admin" />}
        <ConfirmDeleteUserModal
          open={isDeleteModalOpen}
          onClose={() => setIsDeleteModalOpen(false)}
          onConfirm={handleDeleteUser}
          username={user.username}
        />
        <ResetUserPasswordModal
          open={isResetModalOpen}
          setOpen={setIsResetModalOpen}
          userID={user.user_id}
        />
      </CardContent>
    </Card>
  );
}

function ConfirmDeleteUserModal({
  open,
  onClose,
  onConfirm,
  username,
}: {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  username: string;
}) {
  const [checked, setChecked] = useState(false);
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setChecked(event.target.checked);
  };

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Confirm Delete User</DialogTitle>
      <DialogContent>
        Are you sure you want to delete user {username}? This action cannot be
        undone.
        <br />
        <br />
        <FormControlLabel
          control={<Checkbox checked={checked} onChange={handleChange} />}
          label="Yes, I want to delete this user"
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button
          onClick={() => {
            if (checked) {
              onConfirm();
            } else {
              toast.error("Please check the box to proceed.");
            }
          }}
          color="error"
        >
          Delete
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function AddUserModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const createUserMutation = useCreateUserMutation();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const handleClose = () => {
    setUsername("");
    setPassword("");
    setConfirmPassword("");
    setDisplayName("");
    onClose();
  };
  const handleCreateUser = () => {
    if (password !== confirmPassword) {
      toast.error("Passwords don't match!");
      return;
    }
    if (password.length < 8) {
      toast.error("Password too short");
      return;
    }
    if (
      username === "" ||
      displayName === "" ||
      password === "" ||
      confirmPassword === ""
    ) {
      toast.error("Please fill in all fields");
      return;
    }
    createUserMutation.mutate(
      {
        username,
        displayName,
        password,
      },
      {
        onSuccess: () => {
          handleClose();
          toast.success("User created successfully");
        },
        onError: (error) => {
          toast.error("Error creating user: " + error.message);
        },
      },
    );
  };
  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Create New User</DialogTitle>
      <DialogContent>
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
          label="Display Name"
          variant="outlined"
          fullWidth
          required
          margin="normal"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
        />
        <TextField
          label="Password"
          variant="outlined"
          fullWidth
          required
          margin="normal"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          inputProps={{
            autocomplete: "new-password",
            form: {
              autocomplete: "off",
            },
          }}
        />
        <TextField
          label="Confirm Password"
          variant="outlined"
          fullWidth
          required
          margin="normal"
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          inputProps={{
            autocomplete: "new-password",
            form: {
              autocomplete: "off",
            },
          }}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleCreateUser}>Add User</Button>
      </DialogActions>
    </Dialog>
  );
}

function ResetUserPasswordModal({
  open,
  setOpen,
  userID,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
  userID: number;
}) {
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const resetUserPasswordMutation = useResetUserPassword();
  const handleClose = () => {
    setNewPassword("");
    setOpen(false);
  };
  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Reset Password</DialogTitle>
      <DialogContent className="provider-profile-container">
        <hr />
        <Alert severity="warning">
          Warning: This will log the user out of all their devices!
        </Alert>
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
            if (newPassword === "" || confirmNewPassword == "") {
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
            resetUserPasswordMutation.mutate(
              {
                userID,
                newPassword,
              },
              {
                onSuccess: () => {
                  toast.success("Password reset success");
                  handleClose();
                },
                onError: (error: any) => {
                  toast.error("Internal error changing password");
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
