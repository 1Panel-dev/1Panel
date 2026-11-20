import assert from 'node:assert/strict';
import { register } from 'node:module';
import test from 'node:test';
import { crc32 } from 'node:zlib';

// monitor-export.ts only needs dateFormatRFC3339 from '@/utils/date', which pulls in i18n; stub it for node
const hooks = `
export async function resolve(specifier, context, next) {
    if (specifier === '@/utils/date') {
        return {
            url: 'data:text/javascript,export const dateFormatRFC3339 = (d) => new Date(d).toISOString();',
            shortCircuit: true,
        };
    }
    return next(specifier, context);
}`;
register('data:text/javascript,' + encodeURIComponent(hooks));
const { createMonitorExportWriter, XLSX_MAX_DATA_ROWS } = await import('../src/utils/monitor-export.ts');

const BOM = '﻿';
const decode = (bytes) => new TextDecoder('utf-8', { ignoreBOM: true }).decode(bytes);
const blobText = async (blob) => decode(new Uint8Array(await blob.arrayBuffer()));

// independent zip reader: walks the central directory, inflates each entry and checks its CRC and size
const readZip = async (blob) => {
    const buf = new Uint8Array(await blob.arrayBuffer());
    const view = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
    const endOffset = buf.length - 22;
    assert.equal(view.getUint32(endOffset, true), 0x06054b50);
    const count = view.getUint16(endOffset + 10, true);
    let pos = view.getUint32(endOffset + 16, true);
    const files = new Map();
    for (let i = 0; i < count; i++) {
        assert.equal(view.getUint32(pos, true), 0x02014b50);
        const method = view.getUint16(pos + 10, true);
        const crc = view.getUint32(pos + 16, true);
        const compressedSize = view.getUint32(pos + 20, true);
        const size = view.getUint32(pos + 24, true);
        const nameLength = view.getUint16(pos + 28, true);
        const offset = view.getUint32(pos + 42, true);
        const name = decode(buf.subarray(pos + 46, pos + 46 + nameLength));

        assert.equal(view.getUint32(offset, true), 0x04034b50);
        assert.equal(view.getUint32(offset + 14, true), crc);
        assert.equal(view.getUint32(offset + 18, true), compressedSize);
        const dataStart = offset + 30 + view.getUint16(offset + 26, true);
        const raw = buf.subarray(dataStart, dataStart + compressedSize);
        const data =
            method === 8
                ? new Uint8Array(
                      await new Response(
                          new Blob([raw]).stream().pipeThrough(new DecompressionStream('deflate-raw')),
                      ).arrayBuffer(),
                  )
                : raw;
        assert.equal(data.length, size, name);
        assert.equal(crc32(data), crc, name);
        files.set(name, { method, text: decode(data) });
        pos += 46 + nameLength;
    }
    return files;
};

const withoutCompressionStream = async (fn) => {
    const original = globalThis.CompressionStream;
    delete globalThis.CompressionStream;
    try {
        return await fn();
    } finally {
        globalThis.CompressionStream = original;
    }
};

const base = { name: 'base', fields: ['time', 'cpu'], headers: ['Time', 'CPU (%)'] };
const io = { name: 'io', fields: ['time', 'device'], headers: ['Time', 'Disk'] };
const t0 = new Date('2026-01-01T00:00:00.000Z');
const t1 = new Date('2026-01-01T00:05:00.000Z');
const t2 = new Date('2026-01-01T00:10:00.000Z');

test('streams a single csv across batches into the same content as one pass', async () => {
    const writer = createMonitorExportWriter([base], 'csv', 'export');
    await writer.append('base', [[t0, 12.5]]);
    await writer.append('base', []);
    await writer.append('base', [
        [t1, '=cmd'],
        [t2, null],
    ]);
    const file = await writer.finish();
    assert.equal(file.fileName, 'export-base.csv');
    assert.equal(writer.rowCount('base'), 3);
    assert.equal(
        await blobText(file.content),
        `${BOM}Time,CPU (%)\r\n${t0.toISOString()},12.5\r\n${t1.toISOString()},'=cmd\r\n${t2.toISOString()},`,
    );
});

for (const [mode, run] of [
    ['deflated', (fn) => fn()],
    ['stored', withoutCompressionStream],
]) {
    test(`packs several csv files into a valid zip (${mode})`, async () => {
        await run(async () => {
            const writer = createMonitorExportWriter([base, io], 'csv', 'export');
            await writer.append('base', [[t0, 1]]);
            await writer.append('io', [[t0, 'sda,1']]);
            await writer.append('base', [[t1, 2]]);
            const file = await writer.finish();
            assert.equal(file.fileName, 'export.zip');
            const files = await readZip(file.content);
            assert.deepEqual([...files.keys()], ['monitor-base.csv', 'monitor-io.csv']);
            assert.equal(files.get('monitor-base.csv').method, mode === 'deflated' ? 8 : 0);
            assert.equal(
                files.get('monitor-base.csv').text,
                `${BOM}Time,CPU (%)\r\n${t0.toISOString()},1\r\n${t1.toISOString()},2`,
            );
            assert.equal(files.get('monitor-io.csv').text, `${BOM}Time,Disk\r\n${t0.toISOString()},"sda,1"`);
        });
    });
}

test('streams json that matches JSON.stringify of the whole dataset', async () => {
    const writer = createMonitorExportWriter([base, io], 'json', 'export');
    await writer.append('base', [[t0, 1]]);
    await writer.append('base', [
        [t1, null],
        [t2, 3],
    ]);
    const file = await writer.finish();
    assert.equal(file.fileName, 'export.json');
    const expected = {
        base: [
            { time: t0.toISOString(), cpu: 1 },
            { time: t1.toISOString(), cpu: null },
            { time: t2.toISOString(), cpu: 3 },
        ],
        io: [],
    };
    assert.equal(await blobText(file.content), JSON.stringify(expected, null, 2));
});

test('writes xlsx sheets with continuous row numbers across batches', async () => {
    const writer = createMonitorExportWriter([base, io], 'xlsx', 'export');
    await writer.append('base', [[t0, 1]]);
    await writer.append('base', [[t1, 'a<b']]);
    const file = await writer.finish();
    assert.equal(file.fileName, 'export.xlsx');
    assert.deepEqual(file.truncated, []);
    const files = await readZip(file.content);
    assert.deepEqual(
        [...files.keys()],
        [
            '[Content_Types].xml',
            '_rels/.rels',
            'xl/workbook.xml',
            'xl/_rels/workbook.xml.rels',
            'xl/styles.xml',
            'xl/worksheets/sheet1.xml',
            'xl/worksheets/sheet2.xml',
        ],
    );
    assert.match(files.get('xl/workbook.xml').text, /<sheet name="base" sheetId="1" r:id="rId1"\/><sheet name="io"/);
    const sheet = files.get('xl/worksheets/sheet1.xml').text;
    assert.deepEqual(
        [...sheet.matchAll(/<row r="(\d+)">/g)].map((match) => match[1]),
        ['1', '2', '3'],
    );
    assert.match(sheet, /<c r="A2" s="1"><v>[\d.]+<\/v><\/c><c r="B2"><v>1<\/v><\/c>/);
    assert.match(sheet, /<c r="B3" t="inlineStr"><is><t>a&lt;b<\/t><\/is><\/c>/);
    assert.match(sheet, /<\/sheetData><\/worksheet>$/);
    assert.equal(
        files.get('xl/worksheets/sheet2.xml').text.match(/<row /g).length,
        1,
        'an empty table keeps its header row',
    );
});

test('stops writing an xlsx sheet at the Excel row limit and reports it', async () => {
    const writer = createMonitorExportWriter([io], 'xlsx', 'export');
    await writer.append(
        'io',
        Array.from({ length: XLSX_MAX_DATA_ROWS - 1 }, () => [1]),
    );
    await writer.append('io', [[2], [3], [4]]);
    await writer.append('io', [[5]]);
    assert.equal(writer.rowCount('io'), XLSX_MAX_DATA_ROWS);
    const file = await writer.finish();
    assert.deepEqual(file.truncated, ['io']);
    const sheet = (await readZip(file.content)).get('xl/worksheets/sheet1.xml').text;
    assert.ok(
        sheet.endsWith(
            `<row r="${XLSX_MAX_DATA_ROWS + 1}"><c r="A${XLSX_MAX_DATA_ROWS + 1}"><v>2</v></c></row></sheetData></worksheet>`,
        ),
    );
});
