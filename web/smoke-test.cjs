// Run after build.sh: node web/smoke-test.cjs
const fs = require('node:fs');
const assert = require('node:assert/strict');
require('./wasm_exec.js');
(async () => {
  const go = new Go();
  const { instance } = await WebAssembly.instantiate(fs.readFileSync(`${__dirname}/main.wasm`), go.importObject);
  go.run(instance).catch(error => { console.error(error); process.exit(1); });
  // Original NROM program: set the universal background to palette $21, then loop.
  const rom = new Uint8Array(16 + 16384 + 8192);
  rom.set([0x4e, 0x45, 0x53, 0x1a, 1, 1]);
  rom.set([0xa9,0x3f,0x8d,0x06,0x20,0xa9,0x00,0x8d,0x06,0x20,
    0xa9,0x21,0x8d,0x07,0x20,0x4c,0x0f,0x80], 16);
  rom[16+16384-4] = 0; rom[16+16384-3] = 0x80;
  assert.equal(nesLoadROM(rom), '');
  const pixels = new Uint8Array(256*240*4);
  for (let i=0;i<3;i++) assert.equal(nesFrame(0, pixels), '');
  for (let i=0;i<pixels.length;i+=4) assert.deepEqual(Array.from(pixels.subarray(i,i+4)), [100,176,255,255]);
  assert.match(nesLoadROM(new Uint8Array(2)), /short/);
  assert.equal(nesFrame(255, pixels), ''); // Failed loads preserve the working console.
  assert.equal(nesLoadROM(rom), '');
  assert.equal(nesFrame(0, pixels), '');
  console.log('WASM smoke test passed: ROM load, RGBA frame, invalid ROM, reload');
  process.exit(0);
})().catch(error => { console.error(error); process.exit(1); });
