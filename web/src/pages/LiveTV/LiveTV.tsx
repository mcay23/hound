import "./LiveTV.css";
import { useLiveTVCategories } from "../../api/hooks/live_tv";
import LiveVideoPlayer from "../VideoPlayer/LiveVideoPlayer";
import { useEffect, useState } from "react";

function LiveTV(props: any) {
  const { data: liveTVCategories } = useLiveTVCategories(2, 47);
  console.log(liveTVCategories?.[1].stream_url);

  const [sourceURL, setSourceURL] = useState<string | undefined>(undefined);

  useEffect(() => {
    if (liveTVCategories) {
      setSourceURL(liveTVCategories[0]?.stream_url);
    }
  }, [liveTVCategories]);

  return (
    <div className="live-tv-main-container">
      <h2>Live TV</h2>
      <LiveVideoPlayer src={sourceURL || ""} />
      <hr className="mt-3 mb-4" />
      {liveTVCategories?.map((category: any) => (
        <button
          key={category.stream_id}
          onClick={() => setSourceURL(category.stream_url)}
        >
          {category.name}
        </button>
      ))}
    </div>
  );
}

export default LiveTV;
