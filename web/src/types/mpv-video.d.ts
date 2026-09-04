declare namespace JSX {
  interface IntrinsicElements {
    "mpv-video": React.DetailedHTMLProps<
      React.HTMLAttributes<HTMLElement>,
      HTMLElement
    > & {
      src?: string;
      loop?: boolean;
      volume?: number | string;
      mode?: string;
      currentTime?: number;
      rendererName?: string;
      playerId?: number;
    };
  }
}