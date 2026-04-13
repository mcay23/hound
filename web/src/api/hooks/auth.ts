import { useMutation, useQueryClient } from "@tanstack/react-query";
import { changePassword, logout } from "../services/auth";

export const useLogoutMutation = () => {
  return useMutation({
    mutationFn: logout,
  });
};

export const useChangePassword = () => {
  return useMutation({
    mutationFn: ({ oldPassword, newPassword }: { oldPassword: string; newPassword: string }) =>
      changePassword(oldPassword, newPassword),
  });
};

