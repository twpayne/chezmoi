// The light palette is always the first palette. In some cases, it is numbered
// as __palette_0 and in others it is numbered as __palette_1. The dark palette
// is numbered as __palette_1 or __palette_2, depending on the index used for
// the light palette.
var paletteSwitcherLight = document.getElementById("__palette_0");
var paletteSwitcherDark = document.getElementById("__palette_1");

if (!paletteSwitcherLight) {
  paletteSwitcherLight = paletteSwitcherDark;
  paletteSwitcherDark = document.getElementById("__palette_2");
}

paletteSwitcherLight.addEventListener("change", () => location.reload());
paletteSwitcherDark.addEventListener("change", () => location.reload());
