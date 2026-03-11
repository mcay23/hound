import { Spinner } from "react-bootstrap";
import { useWatchStats } from "../../api/hooks/watchHistory";
import "./WatchStats.css";
import { Card } from "@mui/material";

export default function WatchStats() {
  const { data, isLoading } = useWatchStats();
  return (
    <>
      {!isLoading ? (
        <div className="watch-stats-row">
          <Card className="watch-stats-card">
            <div className="watch-stats-title">Movies Watched</div>
            <h5>{data?.movies_watched}</h5>
          </Card>
          <Card className="watch-stats-card">
            <div className="watch-stats-title">Shows Watched</div>
            <h5>{data?.shows_watched}</h5>
          </Card>
          <Card className="watch-stats-card">
            <div className="watch-stats-title">Episodes Watched</div>
            <h5>{data?.episodes_watched}</h5>
          </Card>
          {data && (
            <Card className="watch-stats-card">
              <div className="watch-stats-title">Watch Time</div>
              <h5>
                {data.total_episodes_duration + data.total_movies_duration > 60
                  ? (
                      (data.total_episodes_duration +
                        data.total_movies_duration) /
                      60
                    ).toFixed(1) + " hours"
                  : (
                      data.total_episodes_duration + data.total_movies_duration
                    ).toFixed(0) + " minutes"}
              </h5>
            </Card>
          )}
        </div>
      ) : (
        <>
          <div className="mt-5 d-flex justify-content-center">
            <Spinner />
          </div>
          <div className="mt-2 d-flex justify-content-center">
            Loading Stats...
          </div>
        </>
      )}
    </>
  );
}
