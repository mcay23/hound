import "./Library.css";
import Topnav from "../Topnav";
import axios from "axios";
import { useEffect, useState } from "react";
import toast, { Toaster } from "react-hot-toast";
import CollectionCard from "./CollectionCover";
import HorizontalSection from "../Home/HorizontalSection";
import {
  Button,
  Dialog,
  DialogActions,
  FormControl,
  LinearProgress,
  TextField,
} from "@mui/material";
import Footer from "../Footer";

function Library(props: any) {
  const [collections, setCollections] = useState([]);
  const [primaryCollection, setPrimaryCollection] = useState<any[]>([]);
  const [isCollectionsLoaded, setIsCollectionsLoaded] = useState(false);
  const [isCreateCollectionDialogOpen, setIsCreateCollectionDialogOpen] =
    useState(false);
  const [createCollectionData, setCreateCollectionData] = useState({
    collection_title: "",
    description: "",
    is_public: true,
  });
  const handleCollectionDialogClose = () => {
    setCreateCollectionData({
      collection_title: "",
      description: "",
      is_public: true,
    });
    setIsCreateCollectionDialogOpen(false);
  };
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setCreateCollectionData({
      ...createCollectionData,
      [event.target.name]: event.target.value,
    });
  };
  const handleCreateCollection = () => {
    if (createCollectionData.collection_title === "") {
      toast.error("Title required");
      return;
    }
    if (createCollectionData.description === "") {
      toast.error("Review comment required");
      return;
    }
    axios
      .post(`/api/v1/collection/new`, createCollectionData)
      .then(() => {
        handleCollectionDialogClose();
        window.scrollTo(0, 0);
        window.location.reload();
      })
      .catch((err) => {
        console.log(err);
        toast.error("Error creating collection");
      });
  };
  document.title = "My Collections - Hound";

  useEffect(() => {
    if (!isCollectionsLoaded) {
      const fetchData = async () => {
        var primaryCollectionID;
        await axios
          .get(`/api/v1/collection/all`)
          .then((res) => {
            setCollections(res.data);
            primaryCollectionID = res.data.find((item: any) => {
              return item.is_primary;
            }).collection_id;
          })
          .catch((err) => {
            console.log(err);
          });
        axios
          .get(`/api/v1/collection/${primaryCollectionID}?limit=20&offset=0`)
          .then((res) => {
            setPrimaryCollection(res.data.results);
            setIsCollectionsLoaded(true);
          })
          .catch((err) => {
            console.log(err);
          });
      };
      fetchData();
    }
  });

  return (
    <>
      <Topnav />
      {isCollectionsLoaded ? (
        <div className="library-main-container">
          <div className="library-top-section-container">
            <HorizontalSection
              items={primaryCollection}
              header="From Your Library"
              itemType="poster"
              itemOnClick={undefined}
            />
            {primaryCollection ? (
              ""
            ) : (
              <div className="horizontal-section-header">
                Your library is empty. Try adding some items!
              </div>
            )}
          </div>
          {
            <div className="library-collections-section">
              <div className="library-collections-header">Your Collections</div>
              <div className="library-collections-container">
                <div
                  className={"rounded collection-card-cover"}
                  id="library-collection-create-cover"
                  onClick={() => {
                    setIsCreateCollectionDialogOpen(true);
                  }}
                >
                  <div className={"collection-card-cover-inner"}>
                    Add New collection
                  </div>
                </div>
                {collections.map((item) => (
                  <CollectionCard
                    data={item}
                    key={item["collection_id"]}
                    showCaption={true}
                  />
                ))}
              </div>
            </div>
          }
        </div>
      ) : (
        <LinearProgress className="progress-margin" />
      )}
      <Dialog
        open={isCreateCollectionDialogOpen}
        onClose={handleCollectionDialogClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <div className="reviews-create-dialog-header">
          Create New Collection
        </div>
        <div className="reviews-create-dialog-content">
          <FormControl fullWidth={true}>
            <TextField
              id="outlined-basic"
              className="mt-3"
              label="Title"
              variant="outlined"
              name="collection_title"
              value={createCollectionData.collection_title}
              onChange={handleChange}
            />
            <TextField
              id="outlined-multiline-static"
              className="mt-3"
              label="Description"
              name="description"
              multiline
              rows={4}
              value={createCollectionData.description}
              onChange={handleChange}
            />
          </FormControl>
        </div>
        <DialogActions>
          <Button onClick={handleCollectionDialogClose}>Cancel</Button>
          <Button onClick={handleCreateCollection}>Create</Button>
        </DialogActions>
      </Dialog>
      <Toaster
        toastOptions={{
          duration: 5000,
        }}
      />
      <Footer />
    </>
  );
}

export default Library;
