import {
  Button,
  Card,
  CardContent,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from "@mui/material";
import {
  useCreateUserMutation,
  useDeleteUserMutation,
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
        <h5>{user.username}</h5>
        <div className="d-flex flex-row">
          <Button
            className="mt-2"
            variant="outlined"
            size="small"
            onClick={() => {}}
          >
            Reset Password
          </Button>
          <Button
            className="mt-2 ms-2"
            variant="outlined"
            size="small"
            color="error"
            onClick={() => setIsDeleteModalOpen(true)}
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
  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Confirm Delete User</DialogTitle>
      <DialogContent>
        Are you sure you want to delete user {username}? This action cannot be
        undone.
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
  const [displayName, setDisplayName] = useState("");
  const handleClose = () => {
    setUsername("");
    setPassword("");
    setDisplayName("");
    onClose();
  };
  const handleCreateUser = () => {
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
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleCreateUser}>Add User</Button>
      </DialogActions>
    </Dialog>
  );
}
