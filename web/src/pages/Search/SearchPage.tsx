import axios from "axios";
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import LinearProgress from "@mui/material/LinearProgress";
import HorizontalSection from "../Home/HorizontalSection";
import SearchBar from "../Home/SearchBar";
import Topnav from "../Topnav";
import "./SearchPage.css";

function SearchPage(props: any) {
  const [searchParams] = useSearchParams();
  const [data, setData] = useState({
    tv_results: [],
    movie_results: [],
    game_results: [],
  });
  const [isLoaded, setIsLoaded] = useState(false);
  const [backdropURL, setBackdropURL] = useState("");
  var styles = {
    withBackdrop: {
      // backgroundColor: "blue",
      backgroundImage: "url(" + backdropURL + ")",
      backgroundSize: "cover",
      animation: "backgroundScroll 150s linear infinite",
    },
  };
  var query = searchParams.get("q");
  useEffect(() => {
    axios
      .get(`/api/v1/search?q=${query}`)
      .then((res) => {
        setData(res.data);
        setIsLoaded(true);
      })
      .catch((err) => {
        if (err.response.status === 500) {
          alert("500");
        }
      });
    if (backdropURL === "") {
      axios
        .get("/api/v1/backdrop")
        .then((res) => {
          setBackdropURL(res.data);
        })
        .catch((err) => {
          if (err.response.status === 500) {
            alert("500");
          }
        });
    }
  }, [backdropURL, query]);
  if (isLoaded) {
    document.title = query + " - Hound";
  }
  return (
    <>
      <Topnav />
      <div
        className="search-page-search-section"
        style={backdropURL ? styles.withBackdrop : {}}
      >
        <SearchBar />
      </div>
      {isLoaded ? (
        <div className="search-page-main-section">
          {data?.tv_results?.length > 0 ||
          data?.movie_results?.length > 0 ||
          data?.game_results?.length > 0 ? (
            <>
              <HorizontalSection
                items={data?.tv_results}
                header={"TV Shows"}
                itemType={"search"}
                itemOnClick={undefined}
              />
              <HorizontalSection
                items={data?.movie_results}
                header={"Movies"}
                itemType={"search"}
                itemOnClick={undefined}
              />
              <HorizontalSection
                items={data?.game_results}
                header={"Games"}
                itemType={"search"}
                itemOnClick={undefined}
              />
            </>
          ) : (
            <div className="collection-empty-message search-page-no-results">
              No results.
            </div>
          )}
        </div>
      ) : (
        <LinearProgress className="progress-margin" />
      )}
    </>
  );
}

export default SearchPage;
