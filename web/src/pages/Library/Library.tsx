import "./Library.css";
import { useState } from "react";
import toast from "react-hot-toast";
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
import {
  useCollections,
  useCollectionContents,
  useRecentCollectionItems,
  useCreateCollection,
} from "../../api/hooks/collections";
import { useNavigate } from "react-router-dom";

function Library(props: any) {
  const { data: collections = [], isLoading: isCollectionsLoading } =
    useCollections();
  const { data: recentItems = [], isLoading: isRecentLoading } =
    useRecentCollectionItems();
  const createMutation = useCreateCollection();
  const [isCreateCollectionDialogOpen, setIsCreateCollectionDialogOpen] =
    useState(false);
  const [createCollectionData, setCreateCollectionData] = useState({
    collection_title: "",
    description: "",
    is_public: true,
  });
  const { data: libraryData = [] } = useCollectionContents(
    "hound-library",
    20,
    0,
  );

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
      toast.error("Description required");
      return;
    }
    createMutation.mutate(createCollectionData, {
      onSuccess: () => {
        handleCollectionDialogClose();
        window.scrollTo(0, 0);
      },
      onError: (err) => {
        console.log(err);
        toast.error("Error creating collection");
      },
    });
  };

  document.title = "My Collections - Hound";
  const isLoaded = !isCollectionsLoading && !isRecentLoading;
  const navigate = useNavigate();

  return (
    <>
      {isLoaded ? (
        <div className="library-main-container">
          <div className="library-top-section-container">
            <HorizontalSection
              items={recentItems}
              header="Recently Added"
              itemType="poster"
              itemOnClick={undefined}
            />
            {!(recentItems?.length > 0) && (
              <div className="horizontal-section-header ps-5 pt-5 pb-5">
                Your collections are empty. Try adding some items!
              </div>
            )}
          </div>
          <div className="library-top-section-container">
            <HorizontalSection
              items={libraryData?.records}
              header="In Your Library"
              headerHref="/collection/hound-library"
              itemType="poster"
              itemOnClick={undefined}
            />
            {!(libraryData?.records?.length > 0) && (
              <div className="horizontal-section-header ps-5 pt-5 pb-5">
                Your Library is empty. Try downloading some media!
              </div>
            )}
          </div>
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
              <div
                className={"rounded collection-card-cover"}
                id="library-collection-create-cover"
                onClick={() => {
                  navigate("/collection/hound-library");
                }}
              >
                <div className={"collection-card-cover-inner"}>
                  Hound Library
                </div>
              </div>
              {collections?.map((item: any) => (
                <CollectionCard
                  data={item}
                  key={item["collection_id"]}
                  showCaption={true}
                />
              ))}
            </div>
          </div>
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
      <Footer />
    </>
  );
}

export default Library;
