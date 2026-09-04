import { useEffect, useState, useMemo } from "react";
import HorizontalSection from "./HorizontalSection";
import SearchBar from "./SearchBar";
import "./Home.css";
import {
  useBackdrops,
  useContinueWatching,
  useUserHomeRows,
  useHomeRow,
} from "../../api/hooks/home";
import Footer from "../Footer";
import { isPlatformElectron } from "../../utils/platform";

function Home() {
  const { data: backdropsData } = useBackdrops();
  const { data: continueWatchingData, isLoading: isContinueWatchingLoading } =
    useContinueWatching();
  const { data: userHomeRows, isLoading: isUserHomeRowsLoading } =
    useUserHomeRows();
  const homeRows = useHomeRow(
    isUserHomeRowsLoading ? 0 : (userHomeRows?.home_rows?.length ?? 0),
  );
  const [backdropURI, setBackdropURI] = useState("");

  const styles = useMemo(
    () => ({
      withBackdrop: {
        backgroundImage: "url(" + backdropURI + ")",
        backgroundSize: "cover",
        animation: "backgroundScroll 150s linear infinite",
      },
    }),
    [backdropURI],
  );

  useEffect(() => {
    if (backdropsData && !backdropURI) {
      setBackdropURI(backdropsData);
    }
  }, [backdropsData, backdropURI]);

  return (
    <>
      <div
        className="home-page-search-section"
        style={backdropURI ? styles.withBackdrop : {}}
      >
        <SearchBar />
      </div>
      <div className="home-page-main-section">
        {!isContinueWatchingLoading && continueWatchingData?.length > 0 ? (
          <div className="mt-3">
            <HorizontalSection
              items={continueWatchingData}
              header="Continue Watching"
              itemType="watch_tile"
              itemOnClick={undefined}
            />
          </div>
        ) : (
          <></>
        )}
        {homeRows.map((homeRow, index) => {
          if (!(homeRow?.data?.items?.length > 0)) {
            return <></>;
          }
          return (
            <div
              key={`home-row-${index}`}
              className={index === 0 ? "home-page-primary-section" : "mt-3"}
            >
              <HorizontalSection
                items={homeRow?.data?.items}
                header={homeRow?.data?.title}
                itemType={"poster"}
                itemOnClick={undefined}
              />
              {index !== 0 &&
                !homeRow.isLoading &&
                !homeRow.isError &&
                index !== homeRows?.length - 1 &&
                homeRow?.data?.items?.length > 0 && (
                  <div className="home-page-section-divider" />
                )}
            </div>
          );
        })}
      </div>
      {!isPlatformElectron && <Footer />}
    </>
  );
}

export default Home;
