import axios from "axios";

export const fetchAllCollections = async () => {
  const { data } = await axios.get("/api/v1/collection/all");
  return data;
};

export const fetchCollectionContents = async (id: number | string, limit = 20, offset = 0) => {
  const { data } = await axios.get(`/api/v1/collection/${id}?limit=${limit}&offset=${offset}`);
  return data;
};

export const fetchRecentCollectionItems = async () => {
  const { data } = await axios.get("/api/v1/collection/recent");
  return data;
};

export const createCollection = async (collectionData: {
  collection_title: string;
  description: string;
  is_public: boolean;
}) => {
  const { data } = await axios.post("/api/v1/collection/new", collectionData);
  return data;
};
