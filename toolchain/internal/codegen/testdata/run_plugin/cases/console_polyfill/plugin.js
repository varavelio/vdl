exports.generate = () => {
  console.log("log", 1);
  console.error("error", 2);
  console.warn("warn", 3);
  console.info("info", 4);

  return {
    files: [
      {
        path: "console.txt",
        content: "ok",
      },
    ],
  };
};
