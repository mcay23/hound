import Artplayer from "artplayer";
import type { Option } from "artplayer";
import { useEffect, useRef } from "react";

interface PlayerProps extends React.HTMLAttributes<HTMLDivElement> {
  option: Option;
}

export default function VideoPlayer({ option, ...rest }: PlayerProps) {
  const $container = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!$container.current) return;

    const art = new Artplayer({
      ...option,
      container: $container.current,
    });

    return () => {
      art.destroy(false);
    };
  }, [option]);
  return <div ref={$container} {...rest}></div>;
}
