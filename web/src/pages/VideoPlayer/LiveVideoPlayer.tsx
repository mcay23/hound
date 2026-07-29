import React, { useEffect, useRef, useState } from "react";
import videojs from "video.js";
import Player from "video.js/dist/types/player";
import "video.js/dist/video-js.css";
import { copyToClipboard } from "../../helpers/helpers";
import toast from "react-hot-toast";
import { Alert, Button } from "@mui/material";

export default function LiveVideoPlayer({ src }: { src: string }) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const playerRef = useRef<Player | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  useEffect(() => {
    if (!containerRef.current) return;
    if (!playerRef.current) {
      const videoElement = document.createElement("video-js");
      containerRef.current.appendChild(videoElement);
      playerRef.current = videojs(videoElement, {
        controls: true,
        responsive: true,
        fluid: true,
        preload: "auto",
        autoplay: true,
        liveui: true,
        sources: [
          {
            src,
            type: "application/x-mpegURL",
          },
        ],
      });
      playerRef.current.on("error", () => {
        const err = playerRef.current?.error();
        console.error("Video.js Error Code:", err?.code);
        console.error("Video.js Error Message:", err?.message);
        if (!err) {
          setErrorMessage("Unknown error");
          return;
        }
        setErrorMessage(
          err.message ??
            `Video.js error ${err.code}${err.status ? ` (${err.status})` : ""}`,
        );
      });
    } else {
      setErrorMessage(null);
      playerRef.current.src({
        src,
        type: "application/x-mpegURL",
      });
    }
  }, [src]);

  useEffect(() => {
    return () => {
      if (playerRef.current) {
        playerRef.current.dispose();
        playerRef.current = null;
      }
    };
  }, []);

  return (
    <>
      <div ref={containerRef} />
      <div className="mt-3 d-flex justify-content-between align-items-center">
        <div>
          {errorMessage && (
            <>
              <p className="text-danger">
                Error opening stream, try opening in a new tab or using an
                external player:
                <br /> <span className="text-muted">{errorMessage}</span>
              </p>
            </>
          )}
        </div>
        <div>
          <Button
            variant="outlined"
            className="me-2"
            onClick={() => window.open(src, "_blank")}
          >
            Open in New Tab
          </Button>
          <Button
            variant="outlined"
            onClick={async () => {
              try {
                await copyToClipboard(src || "");
                toast.success("Copied to clipboard");
              } catch (err) {
                toast.error("Failed to copy: " + err);
              }
            }}
          >
            Copy Link
          </Button>
        </div>
      </div>
    </>
  );
}
