(function () {
  var d = document.documentElement;
  try {
    var theme = localStorage.getItem("im.theme");
    if (theme === "light" || theme === "dark") d.setAttribute("data-theme", theme);
    else if (window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches) d.setAttribute("data-theme", "light");
    var lang = localStorage.getItem("im.lang");
    var dir = localStorage.getItem("im.dir");
    if (lang && /^[a-z]{2,3}(-[A-Za-z0-9]+)*$/.test(lang)) d.setAttribute("lang", lang);
    if (dir === "rtl" || dir === "ltr") d.setAttribute("dir", dir);
  } catch (e) {}
})();
