import * as React from "react";
import videojs from "video.js";
import "video.js/dist/video-js.css";

interface IVideoPlayerProps {
  options: any;
}

const initialOptions: any = {
  controls: true,
  controlBar: {
    volumePanel: {
      inline: false,
    },
  },
};

function VideoPlayer({ options }: IVideoPlayerProps) {
  const videoNode = React.useRef<HTMLVideoElement>(null);
  const player = React.useRef<any>();
  React.useEffect(() => {
    if (!videoNode.current) return;
    player.current = videojs(videoNode.current, {
      ...initialOptions,
      ...options,
    }).ready(function () {
      // console.log('onPlayerReady', this);
    });
    return () => {
      if (player.current) {
        player.current.dispose();
      }
    };
  }, [options]);
  return <video ref={videoNode} className="video-js vjs-fill" />;
}

export default VideoPlayer;
