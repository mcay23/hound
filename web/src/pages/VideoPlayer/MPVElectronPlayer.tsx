import { useEffect, useRef } from "react";

export default function MPVElectronPlayer() {
  const videoRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const video = videoRef.current as any;
    if (!video) return;

    const file =
      "D:\\Joe Hisaishi Budokan\\Joe Hisaishi in Budokan - The Big Screen [1080p, 10bit].mkv";
    video
      .open(file)
      .then(() => {
        console.log("MPV opened file");
        return video.play();
      })
      .catch((err: unknown) => {
        console.error("MPV error:", err);
      });

    return () => {
      video.destroy?.();
    };
  }, []);

  return (
    <div
      style={{
        width: "100%",
        height: "600px",
        background: "black",
      }}
    >
      <mpv-video
        ref={videoRef}
        render-mode="shared-texture"
        volume="100"
        style={{
          width: "100%",
          height: "100%",
          display: "block",
        }}
      />
    </div>
  );
}
