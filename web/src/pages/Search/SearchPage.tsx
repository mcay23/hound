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
    tv_results: {},
    movie_results: {},
  });
  const [isLoaded, setIsLoaded] = useState(false);
  var query = searchParams.get("q");

  console.log(searchParams);
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
  }, [query]);
  console.log(data);
  return (
    <>
      <Topnav />
      <div className="search-page-search-section">
        <SearchBar />
      </div>
      <div className="search-page-main-section">
        {isLoaded ? (
          <div>
            <HorizontalSection
              items={data.tv_results}
              header={"TV Shows"}
              itemType={"search"}
              itemOnClick={undefined}
            />
            <HorizontalSection
              items={data.movie_results}
              header={"Movies"}
              itemType={"search"}
              itemOnClick={undefined}
            />
          </div>
        ) : (
          <div>
            <LinearProgress />
          </div>
        )}
      </div>
    </>
  );
}

export default SearchPage;
