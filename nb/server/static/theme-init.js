// Apply saved theme or system preference before first paint.
(function () {
  var s = localStorage.getItem("editor-theme");
  if (
    s === "dark" ||
    (!s && matchMedia("(prefers-color-scheme: dark)").matches)
  ) {
    document.documentElement.setAttribute("data-theme", "dark");
  }
})();
