import axios from "axios";
import MUIDataTable, {
  MUIDataTableColumnOptions,
  MUIDataTableOptions,
} from "mui-datatables";
import React, { useEffect, useState } from "react";
import "./HistoryTable.css";

function HistoryTable(props: any) {
  const arr: any[] = [];
  const [data, setData] = useState([arr]);
  const [isDataLoaded, setIsDataLoaded] = useState(false);
  var mediaType =
    props.data.media_type === "tvshow" ? "tv" : props.data.media_type;
  useEffect(() => {
    if (!isDataLoaded) {
      axios
        .get(
          `/api/v1/${mediaType}/${props.data.media_source}-${props.data.source_id}/history`,
        )
        .then((res) => {
          var temp: any[][] = [];
          res.data.forEach((rewatch: any) => {
            rewatch?.watch_events?.forEach((item: any) => {
              temp.push([
                item.media_title,
                item.season_number,
                item.episode_number,
                item.watched_at.split("T")[0],
                item.watch_event_id,
              ]);
            });
          });
          setData(temp);
          setIsDataLoaded(true);
        });
    }
  }, [isDataLoaded, mediaType, props.data.media_source, props.data.source_id]);
  const options: MUIDataTableOptions = {
    filterType: "checkbox",
    download: false,
    print: false,
    onRowsDelete: (rowsDeleted) => {
      const idsToDelete = rowsDeleted.data.map((d) => {
        return data[d.dataIndex][4];
      }); // array of all ids to to be deleted
      const payload = {
        watch_event_ids: idsToDelete,
      };
      axios
        .post(
          `/api/v1/${mediaType}/${props.data.media_source}-${props.data.source_id}/history/delete`,
          payload,
        )
        .catch((err) => {
          console.log(err);
        });
    },
  };
  const excludeDisplay: MUIDataTableColumnOptions = {
    display: "excluded",
    filter: false,
  };
  const includeDisplay: MUIDataTableColumnOptions = {
    display: "true",
  };
  const columns = [
    "Title",
    {
      name: "Season",
      options: mediaType === "tv" ? includeDisplay : excludeDisplay,
    },
    {
      name: "Episode",
      options: mediaType === "tv" ? includeDisplay : excludeDisplay,
    },
    "Watch Date",
    { name: "watch_event_id", options: excludeDisplay },
  ];
  if (!isDataLoaded || data.length === 0) {
    return <div className="history-no-data-header">No watch data.</div>;
  }
  return (
    <MUIDataTable
      title={"Your Watch History"}
      data={data}
      columns={columns}
      options={options}
    />
  );
}

export default HistoryTable;
