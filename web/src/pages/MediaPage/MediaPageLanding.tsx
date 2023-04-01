import axios from "axios";
import { useEffect, useState } from "react";
import { useLocation } from "react-router-dom";
import Topnav from "../Topnav";
import MediaPageTV from "./MediaPageTV";
import MediaPageMovie from "./MediaPageMovie";
import { LinearProgress } from "@mui/material";

const valid_sources = ["tmdb"];

function MediaPageLanding() {
  const [data, setData] = useState<any[]>([]);
  const [isDataLoaded, setIsDataLoaded] = useState(false);
  const location = useLocation();
  // get data from api
  console.log(location.pathname);
  useEffect(() => {
    if (!isDataLoaded) {
      axios
        .get("/api/v1" + location.pathname)
        .then((res) => {
          setData(res.data);
          setIsDataLoaded(true);
        })
        .catch((err) => {
          if (err.response.status === 500) {
            alert("500");
          }
        });
    }
  });
  var pathData = location.pathname.split("/");
  const mediaType = pathData[1];
  return (
    <>
      <Topnav />
      {isDataLoaded ? (
        <>
          {mediaType === "tv" ? (
            <MediaPageTV data={data} />
          ) : (
            <MediaPageMovie data={data} />
          )}
        </>
      ) : (
        <LinearProgress />
      )}
    </>
  );
}

export default MediaPageLanding;
