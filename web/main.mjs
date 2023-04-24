import { readFileSync } from 'fs';
import * as url from 'url';

globalThis.crypto = await import('node:crypto');
await import('./wasm_exec.js');

const __filename = url.fileURLToPath(import.meta.url);
const __dirname = url.fileURLToPath(new URL('.', import.meta.url));

const go = new Go();
const mod = await WebAssembly.compile(readFileSync(__dirname + '/go.wasm'));
let inst = await WebAssembly.instantiate(mod, go.importObject);

globalThis.goWasmExports = inst.exports;

async function run() {
    const goInstRan = go.run(inst);

    {
        let result = parseSoql(`select Id from Contact`);
        console.log(result);
    }

    await goInstRan;
}

run();
