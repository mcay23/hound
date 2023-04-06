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
          `/api/v1/${mediaType}/${props.data.media_source}-${props.data.source_id}/comments?type=history`
        )
        .then((res) => {
          if (res.data) {
            var temp: any[][] = [];
            res.data.map((item: any) => {
              var seasonEpisode = item.tag_data.split("E");
              var title = item.title ? item.title : props.data.media_title;
              temp.push([
                title,
                parseInt(seasonEpisode[0].substring(1)),
                parseInt(seasonEpisode[1]),
                item.start_date.split("T")[0],
                item.comment,
                item.comment_id,
              ]);
              return false;
            });
            setData(temp);
          }
          // set data as loaded, even if null
          setIsDataLoaded(true);
        })
        .catch((err) => {
          console.log(err);
        });
    }
  });
  const options: MUIDataTableOptions = {
    filterType: "checkbox",
    download: false,
    print: false,
    onRowsDelete: (rowsDeleted) => {
      const idsToDelete = rowsDeleted.data.map((d) => {
        return data[d.dataIndex][5];
      }); // array of all ids to to be deleted
      axios.delete(`/api/v1/comments?ids=${idsToDelete}`).catch((err) => {
        console.log(err);
      });
    },
  };
  const excludeDisplay: MUIDataTableColumnOptions = {
    display: "excluded",
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
    {
      name: "Notes",
      options: mediaType === "tv" ? excludeDisplay : includeDisplay,
    },
    { name: "comment_id", options: excludeDisplay },
  ];
  if (isDataLoaded && data[0].length === 0) {
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
