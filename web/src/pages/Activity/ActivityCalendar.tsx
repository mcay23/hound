import React, { useMemo, useState } from "react";
import { Calendar, dayjsLocalizer, Views, View } from "react-big-calendar";
import "react-big-calendar/lib/css/react-big-calendar.css";
import { useWatchActivity } from "../../api/hooks/watchHistory";
import { WatchActivity as WatchActivityType } from "../../api/services/watchHistory";
import { FormControl, InputLabel, MenuItem, Select } from "@mui/material";
import dayjs from "dayjs";
import { Spinner } from "react-bootstrap";

const localizer = dayjsLocalizer(dayjs);

const months = Array.from({ length: 12 }, (_, i) =>
  dayjs().month(i).format("MMMM"),
);

const years = Array.from({ length: 10 }, (_, i) => dayjs().year() - 9 + i);

export default function ActivityCalendar() {
  const [date, setDate] = useState<Date>(new Date());
  const [view, setView] = useState<View>(Views.MONTH);

  const startTime = dayjs(date).startOf("month").toISOString();
  const endTime = dayjs(date).endOf("month").toISOString();

  // fetch watch activity for the selected month
  const { data, isLoading } = useWatchActivity(1000, 0, startTime, endTime);

  // format the title in the month view so the season/episode is still shown
  // eg. This show has a long title - S1E2 -> This show has...S1E2
  const formatTitle = useMemo(
    () => (showTitle: string, seasonEpisode: string) => {
      return (
        <div
          style={{
            display: "flex",
            whiteSpace: "nowrap",
            overflow: "hidden",
            width: "100%",
          }}
          title={`${showTitle} - ${seasonEpisode}`}
        >
          <span
            style={{
              flexShrink: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {showTitle}
          </span>
          <span style={{ flexShrink: 0 }}>{` - ${seasonEpisode}`}</span>
        </div>
      );
    },
    [],
  );

  const events = useMemo(() => {
    return (data?.watch_activity || []).map((activity: WatchActivityType) => {
      const timestamp = dayjs(activity.watched_at);
      let title: React.ReactNode = activity.media_title;

      if (activity.record_type === "episode") {
        const seasonEpisode = `S${activity.season_number}E${activity.episode_number}`;
        const showTitle = activity.show_media_title || "Unknown Show";
        if (view === Views.MONTH) {
          title = formatTitle(showTitle, seasonEpisode);
        } else if (view === Views.AGENDA) {
          title = `${showTitle} - ${seasonEpisode}: ${activity.media_title}`;
        } else {
          title = `${showTitle} - ${seasonEpisode}`;
        }
      }
      return {
        title,
        start: timestamp.toDate(),
        end: timestamp.toDate(),
        allDay: false,
      };
    });
  }, [data?.watch_activity, view, formatTitle]);

  // header for the agenda view
  const formats = useMemo(
    () => ({
      agendaHeaderFormat: (
        { start }: { start: Date },
        culture?: string,
        localizer?: any,
      ) => localizer.format(start, "MMMM YYYY", culture),
    }),
    [],
  );

  const handleMonthChange = (monthIndex: number) => {
    const newDate = dayjs(date).month(monthIndex).toDate();
    setDate(newDate);
  };

  const handleYearChange = (year: number) => {
    const newDate = dayjs(date).year(year).toDate();
    setDate(newDate);
  };

  const handleNavigate = (newDate: Date, _view: View, action: string) => {
    if (view === Views.AGENDA) {
      if (action === "NEXT") {
        setDate(dayjs(date).add(1, "month").startOf("month").toDate());
      } else if (action === "PREV") {
        setDate(dayjs(date).subtract(1, "month").startOf("month").toDate());
      } else if (action === "TODAY") {
        setDate(dayjs().startOf("month").toDate());
      } else {
        setDate(dayjs(newDate).startOf("month").toDate());
      }
    } else {
      setDate(newDate);
    }
  };

  const handleViewChange = (newView: View) => {
    setView(newView);
    if (newView === Views.AGENDA) {
      setDate(dayjs(date).startOf("month").toDate());
    }
  };

  return (
    <div style={{ height: "80%" }}>
      <div style={{ marginBottom: 12, display: "flex", gap: 10 }}>
        {/* Year Dropdown */}
        <FormControl>
          <InputLabel>Year</InputLabel>
          <Select
            size="small"
            value={dayjs(date).year()}
            onChange={(e) => handleYearChange(Number(e.target.value))}
            label={"year"}
          >
            {years.map((year) => (
              <MenuItem key={year} value={year}>
                {year}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        {/* Month Dropdown */}
        <FormControl>
          <InputLabel>Month</InputLabel>
          <Select
            size="small"
            value={dayjs(date).month()}
            onChange={(e) => handleMonthChange(Number(e.target.value))}
            label={"month"}
          >
            {months.map((month, index) => (
              <MenuItem key={month} value={index}>
                {month}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </div>

      {isLoading ? (
        <>
          <div className="mt-5 d-flex justify-content-center">
            <Spinner />
          </div>
          <div className="mt-2 d-flex justify-content-center">
            Loading Calendar...
          </div>
        </>
      ) : (
        <Calendar
          localizer={localizer}
          events={events}
          startAccessor="start"
          endAccessor="end"
          date={date}
          onNavigate={(newDate, view, action) =>
            handleNavigate(newDate, view, action)
          }
          view={view}
          onView={handleViewChange}
          views={[Views.MONTH, Views.AGENDA]}
          length={
            new Date(date.getFullYear(), date.getMonth() + 1, 0).getDate() - 1
          }
          popup
          messages={{
            date: "Watch Date",
            time: "Time",
            event: "Watch Activity",
            month: "Calendar",
            agenda: "List",
          }}
          formats={formats}
        />
      )}
    </div>
  );
}
