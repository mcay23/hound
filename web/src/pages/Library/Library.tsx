import "./Library.css";
import { useState } from "react";
import toast from "react-hot-toast";
import CollectionCard from "./CollectionCover";
import CollectionFormDialog, {
  CollectionFormData,
} from "./CollectionFormDialog";
import HorizontalSection from "../Home/HorizontalSection";
import { LinearProgress } from "@mui/material";
import Footer from "../Footer";
import {
  useCollections,
  useCollectionContents,
  useRecentCollectionItems,
  useCreateCollection,
  usePublicCollections,
} from "../../api/hooks/collections";
import { useNavigate } from "react-router-dom";
import MPVElectronPlayer from "../VideoPlayer/MPVElectronPlayer";

const initialCollectionState: CollectionFormData = {
  collection_title: "",
  description: "",
  is_public: true,
};

function Library(props: any) {
  const { data: collections = [], isLoading: isCollectionsLoading } =
    useCollections();
  const {
    data: publicCollections = [],
    isLoading: isPublicCollectionsLoading,
  } = usePublicCollections();
  const { data: recentItems = [], isLoading: isRecentLoading } =
    useRecentCollectionItems();
  const createMutation = useCreateCollection();
  const [isCreateCollectionDialogOpen, setIsCreateCollectionDialogOpen] =
    useState(false);
  const { data: libraryData = [] } = useCollectionContents(
    "hound-library",
    20,
    0,
  );

  const handleCollectionDialogClose = () => {
    setIsCreateCollectionDialogOpen(false);
  };

  const handleCreateCollection = (data: CollectionFormData) => {
    if (data.collection_title === "") {
      toast.error("Title required");
      return;
    }
    if (data.description === "") {
      toast.error("Description required");
      return;
    }
    createMutation.mutate(data, {
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
  const isLoaded =
    !isCollectionsLoading && !isRecentLoading && !isPublicCollectionsLoading;
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
            <MPVElectronPlayer />
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
          <div className="library-public-collections-section">
            <div className="library-collections-header">Public Collections</div>
            <div className="library-collections-container">
              {publicCollections?.map((item: any) => (
                <CollectionCard
                  data={item}
                  key={item["collection_id"]}
                  showCaption={true}
                  dark
                />
              ))}
            </div>
          </div>
        </div>
      ) : (
        <LinearProgress className="progress-margin" />
      )}
      <CollectionFormDialog
        open={isCreateCollectionDialogOpen}
        title="Create New Collection"
        submitLabel="Create"
        initialData={initialCollectionState}
        onClose={handleCollectionDialogClose}
        onSubmit={handleCreateCollection}
      />
      <Footer />
    </>
  );
}

export default Library;
