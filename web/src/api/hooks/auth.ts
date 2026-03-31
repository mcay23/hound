import { useMutation } from "@tanstack/react-query";
import { logout } from "../services/auth";

export const useLogoutMutation = () => {
  return useMutation({
    mutationFn: logout,
  });
};