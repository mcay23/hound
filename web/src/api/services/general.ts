import axios from "axios";

export const fetchServerInfo = async () => {
    const { data } = await axios.get("/api/v1/server_info");
    return data;
}
