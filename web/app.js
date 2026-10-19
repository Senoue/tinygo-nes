"use strict";
const canvas = document.getElementById("screen");
const ctx = canvas.getContext("2d");
const picture = ctx.createImageData(256, 240);
const pixels = new Uint8Array(picture.data.buffer);
const status = document.getElementById("status");
const file = document.getElementById("rom");
const pause = document.getElementById("pause");
const reset = document.getElementById("reset");
const keys = new Set();
const mapping = { KeyZ: 0, KeyX: 1, ShiftRight: 2, Enter: 3,
  ArrowUp: 4, ArrowDown: 5, ArrowLeft: 6, ArrowRight: 7 };
let rom, name, running = false, previous = null, elapsed = 0, loadID = 0;
function setRunning(value) {
  running = value;
  keys.clear(); previous = null; elapsed = 0;
  pause.textContent = value ? "一時停止" : "再開";
  if (rom) status.textContent = `${name} — ${value ? "実行中" : "一時停止"}`;
}
function fail(error) { setRunning(false); status.textContent = String(error); }
for (const type of ["keydown", "keyup"]) {
  window.addEventListener(type, event => {
    if (!(event.code in mapping) || event.target.matches("input, button")) return;
    event.preventDefault();
    if (type === "keydown") keys.add(event.code); else keys.delete(event.code);
  });
}
window.addEventListener("blur", () => { if (rom) setRunning(false); });
document.addEventListener("visibilitychange", () => { if (document.hidden && rom) setRunning(false); });
file.addEventListener("change", async () => {
  const selected = file.files[0];
  if (!selected) return;
  const id = ++loadID;
  setRunning(false);
  try {
    if (selected.size > 8 * 1024 * 1024) throw new Error("ROMは8 MiB以下を選択してください。");
    const data = new Uint8Array(await selected.arrayBuffer());
    if (id !== loadID) return;
    const error = nesLoadROM(data);
    if (error) throw new Error(error);
    rom = data; name = selected.name;
    pause.disabled = reset.disabled = false;
    setRunning(true); file.blur();
  } catch (error) { if (id === loadID) fail(error); }
});
pause.addEventListener("click", () => { setRunning(!running); pause.blur(); });
reset.addEventListener("click", () => {
  const error = nesLoadROM(rom);
  if (error) fail(error); else setRunning(true);
  reset.blur();
});
function tick(now) {
  if (running) {
    if (previous === null) elapsed = 1000 / 60;
    else elapsed += Math.min(now - previous, 50);
    previous = now;
    let mask = 0;
    for (const key of keys) mask |= 1 << mapping[key];
    let drawn = false;
    while (elapsed >= 1000 / 60) {
      const error = nesFrame(mask, pixels);
      if (error) { fail(error); break; }
      elapsed -= 1000 / 60; drawn = true;
    }
    if (drawn) ctx.putImageData(picture, 0, 0);
  }
  requestAnimationFrame(tick);
}
(async () => {
  try {
    const go = new Go();
    const response = await fetch("main.wasm");
    if (!response.ok) throw new Error(`WASM読み込み失敗: HTTP ${response.status}`);
    const { instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
    go.run(instance).catch(error => { file.disabled = pause.disabled = reset.disabled = true; fail(error); });
    file.disabled = false;
    status.textContent = "ROMを選択してください。";
    requestAnimationFrame(tick);
  } catch (error) { fail(error); }
})();
