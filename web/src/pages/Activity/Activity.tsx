import "./Activity.css";
import ActivityCalendar from "./ActivityCalendar";
import WatchStats from "./WatchStats";

function Activity(props: any) {
  return (
    <div className="activity-main-container">
      <h2>Your Watch Activity</h2>
      <hr className="mt-3 mb-4" />
      <div className="watch-stats-container">
        <WatchStats />
      </div>
      <hr className="mt-4 mb-4" />
      <div className="activity-calendar-container">
        <ActivityCalendar />
      </div>
    </div>
  );
}

export default Activity;
