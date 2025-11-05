import axios from "axios";
import toast from "react-hot-toast";

/*
    Gets stream from providers
*/
export async function fetchStreams(
  media_type: string,
  media_source: string,
  source_id: string,
  mode: string,
  setStreams: Function,
  setMainStream: Function,
  setIsStreamModalOpen: Function,
  setIsSelectStreamModalOpen: Function,
  setStreamButtonLoading: Function,
  setStreamSelectButtonLoading: Function
) {
  axios
    .get(`/api/v1/${media_type}/${media_source}-${source_id}/providers`)
    .then((res) => {
      setStreams(res.data);
      if (res.data.data.providers[0].streams.length > 0) {
        setMainStream(res.data.data.providers[0].streams[0]);
      } else {
        toast.error("No streams found");
      }
      if (res.data.data.providers[0].streams.length > 5 && mode === "direct") {
        setIsStreamModalOpen(true);
      } else if (res.data.data.providers[0].streams.length > 0) {
        setIsSelectStreamModalOpen(true);
      }
    })
    .catch((err) => {
      if (err.response.status === 500) {
        toast.error("Error getting streams");
      }
    })
    .finally(() => {
      if (mode === "direct") {
        setStreamButtonLoading(false);
      } else if (mode === "select") {
        setStreamSelectButtonLoading(false);
      }
    });
}

// const handleStreamButtonClick = (type: string) => {
//     if (type === "direct") {
//       setStreamButtonLoading(true);
//     } else if (type === "select") {
//       setStreamSelectButtonLoading(true);
//     }
//     if (!streams) {
//       axios
//         .get(
//           `/api/v1/movie/${props.data.media_source}-${props.data.source_id}/providers`
//         )
//         .then((res) => {
//           setStreams(res.data);
//           if (res.data.data.providers[0].streams.length > 0) {
//             setMainStream(res.data.data.providers[0].streams[0]);
//           } else {
//             toast.error("No streams found");
//           }
//           if (
//             res.data.data.providers[0].streams.length > 5 &&
//             type === "direct"
//           ) {
//             setIsStreamModalOpen(true);
//           } else if (res.data.data.providers[0].streams.length > 0) {
//             setIsSelectStreamModalOpen(true);
//           }
//         })
//         .catch((err) => {
//           if (err.response.status === 500) {
//             toast.error("Error getting streams");
//           }
//         })
//         .finally(() => {
//           if (type === "direct") {
//             setStreamButtonLoading(false);
//           } else if (type === "select") {
//             setStreamSelectButtonLoading(false);
//           }
//         });
//     } else {
//       if (type === "direct") {
//         if (streams.data.providers[0].streams.length > 0) {
//           setIsStreamModalOpen(true);
//         }
//         setStreamButtonLoading(false);
//       } else if (type === "select") {
//         if (streams.data.providers[0].streams.length > 0) {
//           setIsSelectStreamModalOpen(true);
//         }
//         setStreamSelectButtonLoading(false);
//       }
//     }
//   };
