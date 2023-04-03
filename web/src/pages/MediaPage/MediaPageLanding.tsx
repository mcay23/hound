import axios from "axios";
import { useEffect, useState } from "react";
import { useLocation } from "react-router-dom";
import Topnav from "../Topnav";
import MediaPageTV from "./MediaPageTV";
import MediaPageMovie from "./MediaPageMovie";
import { LinearProgress } from "@mui/material";
import MediaPageGame from "./MediaPageGame";

const valid_sources = ["tmdb"];

function MediaPageLanding() {
  const [data, setData] = useState<any[]>([]);
  const [isDataLoaded, setIsDataLoaded] = useState(false);
  const location = useLocation();
  // get data from api
  useEffect(() => {
    if (!isDataLoaded) {
      // backend api happens to have same path as fe path
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
  var mediaComponent;
  switch (mediaType) {
    case "tv":
      mediaComponent = <MediaPageTV data={data} />;
      break;
    case "movie":
      mediaComponent = <MediaPageMovie data={data} />;
      break;
    case "game":
      mediaComponent = <MediaPageGame data={data} />;
      break;
  }
  return (
    <>
      <Topnav />
      {isDataLoaded ? <>{mediaComponent}</> : <LinearProgress />}
    </>
  );
}

export default MediaPageLanding;
