import axios from "axios";
import { SERVER_URL } from "../../config/axios_config";

export async function logout() {
  await axios
    .post(
      `${SERVER_URL}/api/v1/auth/logout`,
      {},
      {
        withCredentials: true,
      },
    )
    .catch((err) => {
      console.log(err);
    });
}