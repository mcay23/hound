import { useEffect, useState, useMemo } from "react";
import Topnav from "../Topnav";
import HorizontalSection from "./HorizontalSection";
import SearchBar from "./SearchBar";
import "./Home.css";
import Footer from "../Footer";
import {
  useBackdrops,
  useContinueWatching,
  useTrendingMovies,
  useTrendingTVShows,
} from "../../api/hooks/home";

function Home() {
  const { data: trendingTVShows = [], isLoading: isTrendingTVShowsLoading } =
    useTrendingTVShows();
  const { data: trendingMovies = [], isLoading: isTrendingMoviesLoading } =
    useTrendingMovies();
  const { data: backdropsData } = useBackdrops();
  const { data: continueWatchingData, isLoading: isContinueWatchingLoading } =
    useContinueWatching();

  const [backdropURL, setBackdropURL] = useState("");

  const styles = useMemo(
    () => ({
      withBackdrop: {
        backgroundImage: "url(" + backdropURL + ")",
        backgroundSize: "cover",
        animation: "backgroundScroll 150s linear infinite",
      },
    }),
    [backdropURL]
  );

  useEffect(() => {
    if (backdropsData && !backdropURL) {
      const urls = backdropsData;
      const randomBackdrop = urls[Math.floor(Math.random() * urls.length)];
      setBackdropURL(randomBackdrop);
    }
  }, [backdropsData, backdropURL]);

  return (
    <>
      <Topnav />
      <div
        className="home-page-search-section"
        style={backdropURL ? styles.withBackdrop : {}}
      >
        <SearchBar />
      </div>
      <div className="home-page-main-section">
        {!isTrendingTVShowsLoading ? (
          <div className="home-page-primary-section">
            <HorizontalSection
              items={trendingTVShows}
              header="Trending TV Shows"
              itemType="poster"
              itemOnClick={undefined}
            />
          </div>
        ) : (
          <div className="home-page-placeholder"></div>
        )}
        {!isTrendingMoviesLoading && (
          <HorizontalSection
            items={trendingMovies}
            header="Trending Movies"
            itemType="poster"
            itemOnClick={undefined}
          />
        )}
        {!isContinueWatchingLoading ? (
          <div className="mt-5">
            <HorizontalSection
              items={continueWatchingData}
              header="Continue Watching"
              itemType="watch_tile"
              itemOnClick={undefined}
            />
          </div>
        ) : (
          <div className="home-page-placeholder"></div>
        )}
      </div>
      <Footer />
    </>
  );
}

export default Home;
