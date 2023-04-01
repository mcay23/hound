import { useState, useEffect } from "react";

const xs = 0,
  sm = 600,
  md = 900,
  lg = 1200,
  xl = 1536;

function getWindowDimensions() {
  const { innerWidth: width, innerHeight: height } = window;
  let breakpoint = "xs";
  if (width >= xl) {
    breakpoint = "xl";
  } else if (width >= lg) {
    breakpoint = "lg";
  } else if (width >= md) {
    breakpoint = "md";
  } else if (width >= sm) {
    breakpoint = "sm";
  }
  return {
    width,
    height,
    breakpoint,
  };
}

export default function useWindowDimensions() {
  const [windowDimensions, setWindowDimensions] = useState(
    getWindowDimensions()
  );

  useEffect(() => {
    function handleResize() {
      setWindowDimensions(getWindowDimensions());
    }

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return windowDimensions;
}
