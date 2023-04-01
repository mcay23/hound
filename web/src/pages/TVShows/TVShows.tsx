import "./TVShows.css";
import React, { useEffect, useState } from "react";
import axios from "axios";
import Library from "../Library/Library";
import Topnav from "../Topnav";
import { Pagination } from "@mui/material";
import FilterBar from "../Library/FilterBar";
import useWindowDimensions from "../../helpers/useWindowDimensions";

function TVShows() {
  const { breakpoint } = useWindowDimensions();
  const [library, setLibrary] = useState<any[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [showTitle, setShowTitle] = useState(false);
  const [isLibraryLoaded, setIsLibraryLoaded] = useState(false);
  const itemsPerPage = 24;
  const handlePageChange = (
    event: React.ChangeEvent<unknown>,
    value: number
  ) => {
    setPage(value);
  };
  useEffect(() => {
    axios
      .get(
        "/api/v1/tv/lib?limit=" +
          itemsPerPage +
          "&offset=" +
          itemsPerPage * (page - 1)
      )
      .then((res) => {
        setLibrary(res.data.results);
        setTotalPages(Math.ceil(res.data.total_records / itemsPerPage));
        setIsLibraryLoaded(true);
      })
      .catch((err) => {
        if (err.response.status === 400) {
          alert("400");
        }
        console.log(err);
      });
  }, [page]);

  return (
    <>
      <Topnav />
      <FilterBar showTitle={showTitle} setShowTitle={setShowTitle} />
      {isLibraryLoaded ? (
        <Library
          library={library}
          breakpoint={breakpoint}
          showTitle={showTitle}
        />
      ) : (
        ""
      )}
      <div className="d-flex justify-content-center mb-4 mt-2">
        <div className="paginator-container shadow-lg">
          <Pagination
            id="paginator-component"
            defaultPage={1}
            page={page}
            onChange={handlePageChange}
            count={totalPages}
            size="large"
          />
        </div>
      </div>
    </>
  );
}

export default TVShows;
