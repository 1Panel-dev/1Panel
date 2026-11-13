import { dateFormatRFC3339 } from '@/utils/date';

export interface MonitorExportColumns {
    name: string;
    fields: string[];
    headers: string[];
}

export interface MonitorExportFile {
    content: Blob;
    fileName: string;
    type: string;
    // names of tables whose rows were cut to fit format limits
    truncated: string[];
}

// Writes an export incrementally: every batch of rows is encoded as soon as it arrives and moved into
// Blob chunks, so finished batches leave the JS heap (browsers may page large blobs to disk) and memory
// stays flat no matter how long the exported range is.
export interface MonitorExportWriter {
    // rows of one table must be appended in time order
    append(name: string, rows: any[][]): Promise<void>;
    rowCount(name: string): number;
    finish(): Promise<MonitorExportFile>;
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

const csvEscape = (value: string) => {
    // neutralize formula execution when a non-numeric cell starts with = + - @ or a control char
    let sanitized = value;
    if (/^[=+\-@\t]/.test(sanitized) && Number.isNaN(Number(sanitized))) {
        sanitized = `'${sanitized}`;
    }
    return /[",\n\r]/.test(sanitized) ? `"${sanitized.replace(/"/g, '""')}"` : sanitized;
};

const csvLine = (cells: any[]) => cells.map((cell) => csvEscape(formatCell(cell))).join(',');

const jsonItem = (columns: MonitorExportColumns, row: any[]) => {
    const item: Record<string, any> = {};
    columns.fields.forEach((field, i) => {
        item[field] = row[i] instanceof Date ? formatCell(row[i]) : (row[i] ?? null);
    });
    // indented to sit inside the top-level object exactly like JSON.stringify(data, null, 2)
    return '    ' + JSON.stringify(item, null, 2).replace(/\n/g, '\n    ');
};

const FLUSH_BYTES = 4 * 1024 * 1024;

// collects output parts and periodically folds them into a Blob so the source strings and arrays can be freed
class BlobBuffer {
    private blobs: Blob[] = [];
    private pending: Array<string | Uint8Array<ArrayBuffer>> = [];
    private pendingSize = 0;

    push(part: string | Uint8Array<ArrayBuffer>) {
        if (part.length === 0) {
            return;
        }
        this.pending.push(part);
        this.pendingSize += typeof part === 'string' ? part.length : part.byteLength;
        if (this.pendingSize >= FLUSH_BYTES) {
            this.flush();
        }
    }

    toBlob(): Blob {
        this.flush();
        return new Blob(this.blobs);
    }

    private flush() {
        if (this.pending.length > 0) {
            this.blobs.push(new Blob(this.pending));
            this.pending = [];
            this.pendingSize = 0;
        }
    }
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

const crc32Update = (crc: number, data: Uint8Array): number => {
    for (let i = 0; i < data.length; i++) {
        crc = crcTable[(crc ^ data[i]) & 0xff] ^ (crc >>> 8);
    }
    return crc;
};

const dosDateTime = (date: Date) => {
    const time =
        ((date.getHours() & 0x1f) << 11) | ((date.getMinutes() & 0x3f) << 5) | ((date.getSeconds() >> 1) & 0x1f);
    const day =
        (((date.getFullYear() - 1980) & 0x7f) << 9) | (((date.getMonth() + 1) & 0xf) << 5) | (date.getDate() & 0x1f);
    return { time, day };
};

// deflate-raw through the browser's native CompressionStream; null means unsupported, the entry is stored instead
const createDeflateStream = (): CompressionStream | null => {
    if (typeof CompressionStream === 'undefined') {
        return null;
    }
    try {
        return new CompressionStream('deflate-raw');
    } catch {
        // engines that predate the 'deflate-raw' format throw on construction
        return null;
    }
};

interface ZipEntryData {
    nameBytes: Uint8Array<ArrayBuffer>;
    crc: number;
    size: number;
    method: number;
    data: Blob;
}

const encoder = new TextEncoder();

// streams one zip entry: tracks the CRC and size of the raw bytes while compressing them on the fly
class ZipEntryWriter {
    private crc = 0xffffffff;
    private size = 0;
    private out = new BlobBuffer();
    private writer: WritableStreamDefaultWriter<BufferSource> | null = null;
    private drained: Promise<void> = Promise.resolve();
    private method = 0;
    private name: string;

    constructor(name: string) {
        this.name = name;
        const stream = createDeflateStream();
        if (stream) {
            this.method = 8;
            this.writer = stream.writable.getWriter();
            this.drained = (async () => {
                const reader = stream.readable.getReader();
                for (;;) {
                    const { done, value } = await reader.read();
                    if (done) {
                        return;
                    }
                    this.out.push(value as Uint8Array<ArrayBuffer>);
                }
            })();
            // surfaced through close(); keeps an early stream failure from becoming an unhandled rejection
            this.drained.catch(() => undefined);
        }
    }

    async write(content: string | Uint8Array<ArrayBuffer>) {
        const bytes = typeof content === 'string' ? encoder.encode(content) : content;
        if (bytes.length === 0) {
            return;
        }
        this.crc = crc32Update(this.crc, bytes);
        this.size += bytes.length;
        if (this.writer) {
            await this.writer.write(bytes);
        } else {
            this.out.push(bytes);
        }
    }

    async close(): Promise<ZipEntryData> {
        if (this.writer) {
            await this.writer.close();
            await this.drained;
        }
        return {
            nameBytes: encoder.encode(this.name),
            crc: (this.crc ^ 0xffffffff) >>> 0,
            size: this.size,
            method: this.method,
            data: this.out.toBlob(),
        };
    }
}

const buildZip = (entries: ZipEntryData[]): Blob => {
    const parts: BlobPart[] = [];
    const central: Uint8Array<ArrayBuffer>[] = [];
    let offset = 0;
    const { time, day } = dosDateTime(new Date());

    for (const item of entries) {
        const local = new DataView(new ArrayBuffer(30));
        local.setUint32(0, 0x04034b50, true);
        local.setUint16(4, 20, true);
        local.setUint16(6, 0x0800, true);
        local.setUint16(8, item.method, true);
        local.setUint16(10, time, true);
        local.setUint16(12, day, true);
        local.setUint32(14, item.crc, true);
        local.setUint32(18, item.data.size, true);
        local.setUint32(22, item.size, true);
        local.setUint16(26, item.nameBytes.length, true);
        parts.push(new Uint8Array(local.buffer), item.nameBytes, item.data);

        const header = new DataView(new ArrayBuffer(46));
        header.setUint32(0, 0x02014b50, true);
        header.setUint16(4, 20, true);
        header.setUint16(6, 20, true);
        header.setUint16(8, 0x0800, true);
        header.setUint16(10, item.method, true);
        header.setUint16(12, time, true);
        header.setUint16(14, day, true);
        header.setUint32(16, item.crc, true);
        header.setUint32(20, item.data.size, true);
        header.setUint32(24, item.size, true);
        header.setUint16(28, item.nameBytes.length, true);
        header.setUint32(42, offset, true);
        central.push(new Uint8Array(header.buffer), item.nameBytes);

        offset += 30 + item.nameBytes.length + item.data.size;
    }

    const centralSize = central.reduce((sum, chunk) => sum + chunk.length, 0);
    const end = new DataView(new ArrayBuffer(22));
    end.setUint32(0, 0x06054b50, true);
    end.setUint16(8, entries.length, true);
    end.setUint16(10, entries.length, true);
    end.setUint32(12, centralSize, true);
    end.setUint32(16, offset, true);
    return new Blob([...parts, ...central, new Uint8Array(end.buffer)]);
};

const zipStatic = async (name: string, content: string) => {
    const entry = new ZipEntryWriter(name);
    await entry.write(content);
    return entry.close();
};

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

const xlsxRow = (index: number, cells: any[], isHeader = false) => {
    const parts = [`<row r="${index}">`];
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
    return parts.join('');
};

const xlsxStaticEntries = (tables: MonitorExportColumns[]) => {
    const sheets = tables
        .map((table, i) => `<sheet name="${xmlEscape(table.name)}" sheetId="${i + 1}" r:id="rId${i + 1}"/>`)
        .join('');
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
    return [
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
                `<sheets>${sheets}</sheets></workbook>`,
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
};

// Excel worksheets are limited to 1,048,576 rows (header included)
export const XLSX_MAX_DATA_ROWS = 1048575;

interface TableState {
    columns: MonitorExportColumns;
    rows: number;
    truncated: boolean;
}

const createTableStates = (tables: MonitorExportColumns[]) => {
    const states = new Map<string, TableState>();
    for (const columns of tables) {
        states.set(columns.name, { columns, rows: 0, truncated: false });
    }
    const get = (name: string) => {
        const state = states.get(name);
        if (!state) {
            throw new Error(`unknown export table: ${name}`);
        }
        return state;
    };
    return { states, get };
};

const truncatedNames = (states: Map<string, TableState>) =>
    [...states.values()].filter((state) => state.truncated).map((state) => state.columns.name);

const createCSVWriter = (tables: MonitorExportColumns[], baseName: string): MonitorExportWriter => {
    const { get } = createTableStates(tables);
    if (tables.length === 1) {
        const buffer = new BlobBuffer();
        buffer.push(String.fromCharCode(0xfeff) + csvLine(tables[0].headers));
        return {
            async append(name, rows) {
                const state = get(name);
                if (rows.length > 0) {
                    buffer.push(rows.map((row) => '\r\n' + csvLine(row)).join(''));
                    state.rows += rows.length;
                }
            },
            rowCount: (name) => get(name).rows,
            async finish() {
                return {
                    content: buffer.toBlob(),
                    fileName: `${baseName}-${tables[0].name}.csv`,
                    type: 'text/csv',
                    truncated: [],
                };
            },
        };
    }
    // several metrics: one csv per metric packed into a zip, each entry compressed while it is written
    const entries = new Map<string, ZipEntryWriter>();
    const ready = Promise.all(
        tables.map((columns) => {
            const entry = new ZipEntryWriter(`monitor-${columns.name}.csv`);
            entries.set(columns.name, entry);
            return entry.write(String.fromCharCode(0xfeff) + csvLine(columns.headers));
        }),
    );
    // awaited by append/finish; this only keeps an early failure from being reported as unhandled
    ready.catch(() => undefined);
    return {
        async append(name, rows) {
            await ready;
            const state = get(name);
            if (rows.length > 0) {
                await entries.get(name)!.write(rows.map((row) => '\r\n' + csvLine(row)).join(''));
                state.rows += rows.length;
            }
        },
        rowCount: (name) => get(name).rows,
        async finish() {
            await ready;
            const closed = await Promise.all(tables.map((columns) => entries.get(columns.name)!.close()));
            return {
                content: buildZip(closed),
                fileName: `${baseName}.zip`,
                type: 'application/zip',
                truncated: [],
            };
        },
    };
};

const createJSONWriter = (tables: MonitorExportColumns[], baseName: string): MonitorExportWriter => {
    const { get } = createTableStates(tables);
    const buffers = new Map(tables.map((columns) => [columns.name, new BlobBuffer()]));
    return {
        async append(name, rows) {
            const state = get(name);
            if (rows.length === 0) {
                return;
            }
            const items = rows.map((row) => jsonItem(state.columns, row)).join(',\n');
            buffers.get(name)!.push(state.rows > 0 ? ',\n' + items : items);
            state.rows += rows.length;
        },
        rowCount: (name) => get(name).rows,
        async finish() {
            const parts: BlobPart[] = ['{\n'];
            tables.forEach((columns, i) => {
                const key = `  ${JSON.stringify(columns.name)}: `;
                const separator = i < tables.length - 1 ? ',\n' : '\n';
                if (get(columns.name).rows === 0) {
                    parts.push(`${key}[]${separator}`);
                } else {
                    parts.push(`${key}[\n`, buffers.get(columns.name)!.toBlob(), `\n  ]${separator}`);
                }
            });
            parts.push('}');
            return {
                content: new Blob(parts),
                fileName: `${baseName}.json`,
                type: 'application/json',
                truncated: [],
            };
        },
    };
};

const createXLSXWriter = (tables: MonitorExportColumns[], baseName: string): MonitorExportWriter => {
    const { states, get } = createTableStates(tables);
    const sheets = new Map<string, ZipEntryWriter>();
    const ready = Promise.all(
        tables.map((columns, i) => {
            const sheet = new ZipEntryWriter(`xl/worksheets/sheet${i + 1}.xml`);
            sheets.set(columns.name, sheet);
            return sheet.write(
                `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
                    `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>` +
                    xlsxRow(1, columns.headers, true),
            );
        }),
    );
    ready.catch(() => undefined);
    return {
        async append(name, rows) {
            await ready;
            const state = get(name);
            const room = XLSX_MAX_DATA_ROWS - state.rows;
            if (rows.length > room) {
                state.truncated = true;
            }
            const accepted = rows.length > room ? rows.slice(0, Math.max(0, room)) : rows;
            if (accepted.length === 0) {
                return;
            }
            // header occupies row 1, so data row n lands on sheet row n + 1
            const xml = accepted.map((row, i) => xlsxRow(state.rows + i + 2, row)).join('');
            await sheets.get(name)!.write(xml);
            state.rows += accepted.length;
        },
        rowCount: (name) => get(name).rows,
        async finish() {
            await ready;
            const staticEntries = await Promise.all(
                xlsxStaticEntries(tables).map((entry) => zipStatic(entry.name, entry.content)),
            );
            const sheetEntries = await Promise.all(
                tables.map(async (columns) => {
                    const sheet = sheets.get(columns.name)!;
                    await sheet.write(`</sheetData></worksheet>`);
                    return sheet.close();
                }),
            );
            return {
                content: buildZip([...staticEntries, ...sheetEntries]),
                fileName: `${baseName}.xlsx`,
                type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
                truncated: truncatedNames(states),
            };
        },
    };
};

export function createMonitorExportWriter(
    tables: MonitorExportColumns[],
    format: string,
    baseName: string,
): MonitorExportWriter {
    switch (format) {
        case 'json':
            return createJSONWriter(tables, baseName);
        case 'xlsx':
            return createXLSXWriter(tables, baseName);
        default:
            return createCSVWriter(tables, baseName);
    }
}
