import { useEffect, useRef } from "react";
import { useLogoutMutation } from "../api/hooks/auth";

export default function Logout() {
  const logout = useLogoutMutation();
  const logoutRequested = useRef(false);

  useEffect(() => {
    if (logoutRequested.current) return;
    logoutRequested.current = true;

    const performLogout = async () => {
      try {
        await logout.mutateAsync();
      } catch (e) {
        console.error("Logout mutation failed", e);
      } finally {
        localStorage.removeItem("isAuthenticated");
        localStorage.removeItem("username");
        localStorage.removeItem("role");
        window.location.href = "/login";
      }
    };

    performLogout();
  }, [logout]);

  return <></>;
}
