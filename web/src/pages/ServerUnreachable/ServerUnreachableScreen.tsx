import { useCallback } from "react";
import { clearSecureToken } from "../../utils/secureStore";
import { checkServerReachable } from "../../utils/serverHealth";
import "./ServerUnreachableScreen.css";
import { Button } from "@mui/material";
import { getBaseUrl } from "../../config/axios_config";

type Props = { onResolved: () => void };

function ServerUnreachableScreen({ onResolved }: Props) {
  const retry = useCallback(async () => {
    const ok = await checkServerReachable();
    if (ok) onResolved();
  }, [onResolved]);

  const changeHost = async () => {
    await clearSecureToken();
    localStorage.removeItem("isAuthenticated");
    localStorage.removeItem("username");
    localStorage.removeItem("role");
    localStorage.removeItem("displayName");
    window.location.href = "/login";
  };

  return (
    <div className="srv-unreach">
      <div className="srv-unreach-card">
        <p className="srv-unreach-title">
          The Hound Server at {getBaseUrl()} is unreachable
        </p>
        <p className="srv-unreach-subtitle text-muted">
          Contact your admin to check if the server is online.
        </p>
        <div className="srv-unreach-actions">
          <Button variant="contained" size="small" onClick={retry}>
            Retry
          </Button>
          <Button variant="contained" size="small" onClick={changeHost}>
            Change Host
          </Button>
        </div>
      </div>
    </div>
  );
}

export default ServerUnreachableScreen;
