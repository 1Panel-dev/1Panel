import { dateFormatRFC3339 } from '@/utils/date';
import { downloadBlob } from '@/utils/file';

export interface MonitorExportTable {
    name: string;
    fields: string[];
    headers: string[];
    rows: any[][];
}

const formatCell = (cell: any): string => {
    if (cell === null || cell === undefined) {
        return '';
    }
    if (cell instanceof Date) {
        // RFC3339 keeps the UTC offset so exported timestamps stay unambiguous across DST transitions
        return dateFormatRFC3339(cell);
    }
    return String(cell);
};

const isNumberCell = (cell: any): boolean => typeof cell === 'number' && Number.isFinite(cell);

export function buildMonitorCSV(table: MonitorExportTable): string {
    const escape = (value: string) => {
        // neutralize formula execution when a non-numeric cell starts with = + - @ or a control char
        let sanitized = value;
        if (/^[=+\-@\t]/.test(sanitized) && Number.isNaN(Number(sanitized))) {
            sanitized = `'${sanitized}`;
        }
        return /[",\n\r]/.test(sanitized) ? `"${sanitized.replace(/"/g, '""')}"` : sanitized;
    };
    const lines = [table.headers.map(escape).join(',')];
    for (const row of table.rows) {
        lines.push(row.map((cell) => escape(formatCell(cell))).join(','));
    }
    return String.fromCharCode(0xfeff) + lines.join('\r\n');
}

export function buildMonitorJSON(tables: MonitorExportTable[]): string {
    const data: Record<string, any[]> = {};
    for (const table of tables) {
        data[table.name] = table.rows.map((row) => {
            const item: Record<string, any> = {};
            table.fields.forEach((field, i) => {
                item[field] = row[i] instanceof Date ? formatCell(row[i]) : (row[i] ?? null);
            });
            return item;
        });
    }
    return JSON.stringify(data, null, 2);
}

// minimal zip writer for csv bundles and xlsx containers: deflate when the runtime offers it, store otherwise
const crcTable = (() => {
    const table = new Uint32Array(256);
    for (let n = 0; n < 256; n++) {
        let c = n;
        for (let k = 0; k < 8; k++) {
            c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
        }
        table[n] = c;
    }
    return table;
})();

const crc32 = (data: Uint8Array): number => {
    let crc = 0xffffffff;
    for (let i = 0; i < data.length; i++) {
        crc = crcTable[(crc ^ data[i]) & 0xff] ^ (crc >>> 8);
    }
    return (crc ^ 0xffffffff) >>> 0;
};

const dosDateTime = (date: Date) => {
    const time =
        ((date.getHours() & 0x1f) << 11) | ((date.getMinutes() & 0x3f) << 5) | ((date.getSeconds() >> 1) & 0x1f);
    const day =
        (((date.getFullYear() - 1980) & 0x7f) << 9) | (((date.getMonth() + 1) & 0xf) << 5) | (date.getDate() & 0x1f);
    return { time, day };
};

export interface ZipEntry {
    name: string;
    content: string | Uint8Array<ArrayBuffer>;
}

// deflate-raw through the browser's native CompressionStream; null means unsupported, caller stores instead
const deflateRaw = async (data: Uint8Array<ArrayBuffer>): Promise<Uint8Array<ArrayBuffer> | null> => {
    if (typeof CompressionStream === 'undefined') {
        return null;
    }
    try {
        const stream = new Blob([data]).stream().pipeThrough(new CompressionStream('deflate-raw'));
        return new Uint8Array(await new Response(stream).arrayBuffer());
    } catch {
        // engines that predate the 'deflate-raw' format throw on construction
        return null;
    }
};

export async function zipFiles(entries: ZipEntry[]): Promise<Uint8Array<ArrayBuffer>> {
    const encoder = new TextEncoder();
    const chunks: Uint8Array[] = [];
    const central: Uint8Array[] = [];
    let offset = 0;
    const { time, day } = dosDateTime(new Date());

    const prepared = await Promise.all(
        entries.map(async (entry) => {
            const raw = typeof entry.content === 'string' ? encoder.encode(entry.content) : entry.content;
            const deflated = await deflateRaw(raw);
            const useDeflate = deflated !== null && deflated.length < raw.length;
            return {
                nameBytes: encoder.encode(entry.name),
                crc: crc32(raw),
                size: raw.length,
                method: useDeflate ? 8 : 0,
                data: useDeflate ? deflated : raw,
            };
        }),
    );

    for (const item of prepared) {
        const local = new DataView(new ArrayBuffer(30));
        local.setUint32(0, 0x04034b50, true);
        local.setUint16(4, 20, true);
        local.setUint16(6, 0x0800, true);
        local.setUint16(8, item.method, true);
        local.setUint16(10, time, true);
        local.setUint16(12, day, true);
        local.setUint32(14, item.crc, true);
        local.setUint32(18, item.data.length, true);
        local.setUint32(22, item.size, true);
        local.setUint16(26, item.nameBytes.length, true);
        chunks.push(new Uint8Array(local.buffer), item.nameBytes, item.data);

        const header = new DataView(new ArrayBuffer(46));
        header.setUint32(0, 0x02014b50, true);
        header.setUint16(4, 20, true);
        header.setUint16(6, 20, true);
        header.setUint16(8, 0x0800, true);
        header.setUint16(10, item.method, true);
        header.setUint16(12, time, true);
        header.setUint16(14, day, true);
        header.setUint32(16, item.crc, true);
        header.setUint32(20, item.data.length, true);
        header.setUint32(24, item.size, true);
        header.setUint16(28, item.nameBytes.length, true);
        header.setUint32(42, offset, true);
        central.push(new Uint8Array(header.buffer), item.nameBytes);

        offset += 30 + item.nameBytes.length + item.data.length;
    }

    const centralOffset = offset;
    let centralSize = 0;
    for (const chunk of central) {
        centralSize += chunk.length;
    }
    const end = new DataView(new ArrayBuffer(22));
    end.setUint32(0, 0x06054b50, true);
    end.setUint16(8, entries.length, true);
    end.setUint16(10, entries.length, true);
    end.setUint32(12, centralSize, true);
    end.setUint32(16, centralOffset, true);
    chunks.push(...central, new Uint8Array(end.buffer));

    const total = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const result = new Uint8Array(total);
    let pos = 0;
    for (const chunk of chunks) {
        result.set(chunk, pos);
        pos += chunk.length;
    }
    return result;
}

const xmlEscape = (value: string) =>
    value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

const xlsxColumnName = (index: number): string => {
    let name = '';
    for (let n = index; ;) {
        name = String.fromCharCode(65 + (n % 26)) + name;
        n = Math.floor(n / 26) - 1;
        if (n < 0) {
            return name;
        }
    }
};

// Excel serial keeps the local wall-clock time so the displayed value matches the panel UI
const xlsxDateSerial = (date: Date) => (date.getTime() - date.getTimezoneOffset() * 60000) / 86400000 + 25569;

const xlsxSheet = (table: MonitorExportTable): string => {
    const parts = [
        `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`,
        `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`,
    ];
    const writeRow = (index: number, cells: any[], isHeader = false) => {
        parts.push(`<row r="${index}">`);
        cells.forEach((cell, i) => {
            const ref = `${xlsxColumnName(i)}${index}`;
            if (!isHeader && cell instanceof Date) {
                parts.push(`<c r="${ref}" s="1"><v>${xlsxDateSerial(cell)}</v></c>`);
            } else if (!isHeader && isNumberCell(cell)) {
                parts.push(`<c r="${ref}"><v>${cell}</v></c>`);
            } else {
                parts.push(`<c r="${ref}" t="inlineStr"><is><t>${xmlEscape(formatCell(cell))}</t></is></c>`);
            }
        });
        parts.push(`</row>`);
    };
    writeRow(1, table.headers, true);
    table.rows.forEach((row, i) => writeRow(i + 2, row));
    parts.push(`</sheetData></worksheet>`);
    return parts.join('');
};

export function buildMonitorXLSX(tables: MonitorExportTable[]): Promise<Uint8Array<ArrayBuffer>> {
    const sheets = tables.map(
        (table, i) => `<sheet name="${xmlEscape(table.name)}" sheetId="${i + 1}" r:id="rId${i + 1}"/>`,
    );
    const overrides = tables
        .map(
            (_table, i) =>
                `<Override PartName="/xl/worksheets/sheet${i + 1}.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`,
        )
        .join('');
    const sheetRels = tables
        .map(
            (_table, i) =>
                `<Relationship Id="rId${i + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet${i + 1}.xml"/>`,
        )
        .join('');

    const entries: ZipEntry[] = [
        {
            name: '[Content_Types].xml',
            content:
                `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
                `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
                `<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
                `<Default Extension="xml" ContentType="application/xml"/>` +
                `<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
                `<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>` +
                overrides +
                `</Types>`,
        },
        {
            name: '_rels/.rels',
            content:
                `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
                `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
                `<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
                `</Relationships>`,
        },
        {
            name: 'xl/workbook.xml',
            content:
                `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
                `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
                `<sheets>${sheets.join('')}</sheets></workbook>`,
        },
        {
            name: 'xl/_rels/workbook.xml.rels',
            content:
                `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
                `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">${sheetRels}` +
                `<Relationship Id="rId${tables.length + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>` +
                `</Relationships>`,
        },
        {
            name: 'xl/styles.xml',
            content:
                `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
                `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
                `<numFmts count="1"><numFmt numFmtId="164" formatCode="yyyy-mm-dd hh:mm:ss"/></numFmts>` +
                `<fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts>` +
                `<fills count="2"><fill><patternFill/></fill><fill><patternFill patternType="gray125"/></fill></fills>` +
                `<borders count="1"><border/></borders>` +
                `<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
                `<cellXfs count="2">` +
                `<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
                `<xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
                `</cellXfs>` +
                `<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>` +
                `</styleSheet>`,
        },
    ];
    tables.forEach((table, i) => {
        entries.push({ name: `xl/worksheets/sheet${i + 1}.xml`, content: xlsxSheet(table) });
    });
    return zipFiles(entries);
}

// Excel worksheets are limited to 1,048,576 rows (header included)
export const XLSX_MAX_DATA_ROWS = 1048575;

// returns the names of tables whose rows were truncated to fit format limits
export async function exportMonitorTables(
    tables: MonitorExportTable[],
    format: string,
    baseName: string,
): Promise<string[]> {
    switch (format) {
        case 'json':
            downloadBlob(buildMonitorJSON(tables), `${baseName}.json`, 'application/json');
            return [];
        case 'xlsx': {
            const truncated: string[] = [];
            const limited = tables.map((table) => {
                if (table.rows.length <= XLSX_MAX_DATA_ROWS) {
                    return table;
                }
                truncated.push(table.name);
                return { ...table, rows: table.rows.slice(0, XLSX_MAX_DATA_ROWS) };
            });
            downloadBlob(
                await buildMonitorXLSX(limited),
                `${baseName}.xlsx`,
                'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
            );
            return truncated;
        }
        default:
            if (tables.length === 1) {
                downloadBlob(buildMonitorCSV(tables[0]), `${baseName}-${tables[0].name}.csv`, 'text/csv');
            } else {
                const bundle = await zipFiles(
                    tables.map((table) => ({ name: `monitor-${table.name}.csv`, content: buildMonitorCSV(table) })),
                );
                downloadBlob(bundle, `${baseName}.zip`, 'application/zip');
            }
            return [];
    }
}
