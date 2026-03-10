import React, { useState } from "react";
import { Calendar, dayjsLocalizer, Views, View } from "react-big-calendar";
import dayjs from "dayjs";
import "react-big-calendar/lib/css/react-big-calendar.css";

const localizer = dayjsLocalizer(dayjs);

type CalendarEvent = {
  title: string;
  start: Date;
  end: Date;
};

const events: CalendarEvent[] = [
  {
    title: "Meeting",
    start: dayjs("2026-03-16").startOf("day").toDate(),
    end: dayjs("2026-03-16").startOf("day").toDate(),
  },
  {
    title: "Planning",
    start: new Date(2026, 2, 10, 9, 0),
    end: new Date(2026, 2, 10, 9, 0),
  },
];

const months = Array.from({ length: 12 }, (_, i) =>
  dayjs().month(i).format("MMMM"),
);

const years = Array.from({ length: 30 }, (_, i) => 2000 + i);

export default function ActivityCalendar() {
  const [date, setDate] = useState<Date>(new Date());
  const [view, setView] = useState<View>(Views.MONTH);

  const handleMonthChange = (monthIndex: number) => {
    const newDate = dayjs(date).month(monthIndex).toDate();
    setDate(newDate);
  };

  const handleYearChange = (year: number) => {
    const newDate = dayjs(date).year(year).toDate();
    setDate(newDate);
  };

  const handleNavigate = (newDate: Date) => {
    if (view === Views.AGENDA) {
      setDate(dayjs(newDate).startOf("month").toDate());
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
    <div style={{ height: "700px" }}>
      {/* Navigation Controls */}
      <div style={{ marginBottom: 12, display: "flex", gap: 10 }}>
        {/* Year Dropdown */}
        <select
          value={dayjs(date).year()}
          onChange={(e) => handleYearChange(Number(e.target.value))}
        >
          {years.map((year) => (
            <option key={year} value={year}>
              {year}
            </option>
          ))}
        </select>

        {/* Month Dropdown */}
        <select
          value={dayjs(date).month()}
          onChange={(e) => handleMonthChange(Number(e.target.value))}
        >
          {months.map((month, index) => (
            <option key={month} value={index}>
              {month}
            </option>
          ))}
        </select>
      </div>

      <Calendar<CalendarEvent>
        localizer={localizer}
        events={events}
        startAccessor="start"
        endAccessor="end"
        date={date}
        onNavigate={handleNavigate}
        view={view}
        onView={handleViewChange}
        views={[Views.MONTH, Views.AGENDA]}
        length={31}
      />
    </div>
  );
}
