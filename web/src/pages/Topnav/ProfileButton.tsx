import { Button, Divider, Menu, MenuItem } from "@mui/material";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function ProfileButton() {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const navigate = useNavigate();
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = () => {
    setAnchorEl(null);
  };
  const handleNavigate = (path: string) => {
    handleClose();
    navigate(path);
  };

  return (
    <div>
      <div
        aria-controls={open ? "basic-menu" : undefined}
        aria-haspopup="true"
        aria-expanded={open ? "true" : undefined}
        style={{ cursor: "pointer" }}
        onClick={handleClick}
      >
        <p className="top-navbar-item mb-0">
          {localStorage.getItem("displayName") ||
            localStorage.getItem("username") ||
            "Settings"}
        </p>
      </div>
      <Menu
        id="basic-menu"
        disableAutoFocusItem
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
      >
        <MenuItem onClick={() => handleNavigate("/settings")}>
          My Account
        </MenuItem>
        {localStorage.getItem("role") === "admin" && (
          <MenuItem onClick={() => handleNavigate("/admin")}>
            Admin Panel
          </MenuItem>
        )}
        <MenuItem onClick={() => handleNavigate("/logout")}>Logout</MenuItem>
        <Divider sx={{ backgroundColor: "#000000", borderWidth: "1px" }} />
        <MenuItem
          component="a"
          href="https://github.com/Hound-Media-Server/hound/issues"
          target="_blank"
          rel="noopener noreferrer"
        >
          Report an Issue
        </MenuItem>
        <MenuItem
          component="a"
          href="https://reddit.com/r/HoundMediaServer"
          target="_blank"
          rel="noopener noreferrer"
        >
          Community Forum
        </MenuItem>
        <MenuItem
          component="a"
          href="https://hound-media-server.github.io/hound-site/"
          target="_blank"
          rel="noopener noreferrer"
        >
          Documentation
        </MenuItem>
      </Menu>
    </div>
  );
}
