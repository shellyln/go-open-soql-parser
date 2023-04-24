// deno run --allow-read deno.js

import * as path from "https://deno.land/std@0.57.0/path/mod.ts";
import * as _ from './wasm_exec.js';

const __filename = path.fromFileUrl(import.meta.url);
const __dirname = path.dirname(path.fromFileUrl(import.meta.url));

const go = new Go();
const f = await Deno.open(__dirname + '/go.wasm');
const mod = await WebAssembly.compile(await Deno.readAll(f));
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
