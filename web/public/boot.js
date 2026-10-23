(function () {
  var d = document.documentElement;
  try {
    var theme = localStorage.getItem("im.theme");
    if (theme === "light" || theme === "dark") d.setAttribute("data-theme", theme);
    else if (window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches) d.setAttribute("data-theme", "light");
    var lang = localStorage.getItem("im.lang");
    if (lang === "ar") {
      d.setAttribute("lang", "ar");
      d.setAttribute("dir", "rtl");
    }
  } catch (e) {}
})();
