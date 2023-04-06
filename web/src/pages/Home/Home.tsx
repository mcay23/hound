import axios from "axios";
import { useEffect, useState } from "react";
import Topnav from "../Topnav";
import HorizontalSection from "./HorizontalSection";
import SearchBar from "./SearchBar";
import "./Home.css";
import Footer from "../Footer";

function Home() {
  const [trendingTVShows, setTrendingTVShows] = useState<any[]>([]);
  const [isTrendingTVShowsLoaded, setIsTrendingTVShowsLoaded] = useState(false);
  const [trendingMovies, setTrendingMovies] = useState<any[]>([]);
  const [isTrendingMoviesLoaded, setIsTrendingMoviesLoaded] = useState(false);
  const [backdropURL, setBackdropURL] = useState("");

  var styles = {
    withBackdrop: {
      // backgroundColor: "blue",
      backgroundImage: "url(" + backdropURL + ")",
      backgroundSize: "cover",
      animation: "backgroundScroll 150s linear infinite",
    },
  };

  useEffect(() => {
    if (!isTrendingTVShowsLoaded) {
      axios
        .get("/api/v1/tv/trending")
        .then((res) => {
          setTrendingTVShows(res.data);
          setIsTrendingTVShowsLoaded(true);
        })
        .catch((err) => {
          if (err.response.status === 500) {
            alert("500");
          }
        });
    }
    if (backdropURL === "") {
      axios
        .get("/api/v1/backdrops")
        .then((res) => {
          var randomBackdrop =
            res.data.backdrop_urls[
              Math.floor(Math.random() * res.data.backdrop_urls.length)
            ];
          console.log(randomBackdrop);
          setBackdropURL(randomBackdrop);
        })
        .catch((err) => {
          if (err.response.status === 500) {
            alert("500");
          }
        });
    }
  });

  useEffect(() => {
    if (!isTrendingMoviesLoaded) {
      axios
        .get("/api/v1/movie/trending")
        .then((res) => {
          setTrendingMovies(res.data);
          setIsTrendingMoviesLoaded(true);
        })
        .catch((err) => {
          if (err.response.status === 500) {
            alert("500");
          }
        });
    }
  });

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
        {isTrendingTVShowsLoaded ? (
          <div className="home-page-primary-section">
            <HorizontalSection
              items={trendingTVShows}
              header="Trending TV Shows"
              itemType="poster"
              itemOnClick={undefined}
            />
          </div>
        ) : (
          ""
        )}
        {isTrendingMoviesLoaded ? (
          <HorizontalSection
            items={trendingMovies}
            header="Trending Movies"
            itemType="poster"
            itemOnClick={undefined}
          />
        ) : (
          ""
        )}
      </div>
      <Footer />
    </>
  );
}

export default Home;
