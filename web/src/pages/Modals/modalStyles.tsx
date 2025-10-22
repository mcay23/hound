export const slotPropsGlass = {
  backdrop: {
    sx: {
      backgroundColor: "rgba(0, 0, 0, 0.3)",
      backdropFilter: "blur(1.5px)",
    },
  },
};

export const paperPropsGlass = {
  sx: {
    background: "rgba(255, 255, 255, 0.4)",
    backdropFilter: "blur(10px)",
    border: "1px solid rgba(255, 255, 255, 0.2)",
    boxShadow: "0 4px 30px rgba(0, 0, 0, 0.2)",
    borderRadius: 3,
    overflow: "hidden",
    transition: "all 0.4s ease-in-out",
    transform: "scale(0.98)",
    "&.MuiDialog-paper": {
      animation: "dialogIn 0.4s ease-out forwards",
    },
    "@keyframes dialogIn": {
      from: { opacity: 0, transform: "scale(0.95) translateY(20px)" },
      to: { opacity: 1, transform: "scale(1) translateY(0)" },
    },
  },
};
