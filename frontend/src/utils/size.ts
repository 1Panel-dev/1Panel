function formattedNumber(num: string) {
    return num.endsWith('.00') ? Number(num.slice(0, -3)) : Number(num);
}

export type BinarySizeUnit = 'B' | 'KiB' | 'MiB' | 'GiB' | 'TiB';

const binarySizeUnitPower: Record<BinarySizeUnit, number> = {
    B: 0,
    KiB: 1,
    MiB: 2,
    GiB: 3,
    TiB: 4,
};

export function convertBinarySize(size: number, from: BinarySizeUnit, to: BinarySizeUnit, precision = 2): number {
    if (!size || from === to) {
        return size;
    }
    const bytes = size * Math.pow(1024, binarySizeUnitPower[from]);
    return formattedNumber((bytes / Math.pow(1024, binarySizeUnitPower[to])).toFixed(precision));
}

export function computeSize(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' B';
    if (size < Math.pow(num, 2)) return formattedNumber((size / num).toFixed(2)) + ' KB';
    if (size < Math.pow(num, 3)) return formattedNumber((size / Math.pow(num, 2)).toFixed(2)) + ' MB';
    if (size < Math.pow(num, 4)) return formattedNumber((size / Math.pow(num, 3)).toFixed(2)) + ' GB';
    return formattedNumber((size / Math.pow(num, 4)).toFixed(2)) + ' TB';
}

export function computeSizeForDocker(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' B';
    if (size < Math.pow(num, 2)) return formattedNumber((size / num).toFixed(2)) + ' KiB';
    if (size < Math.pow(num, 3)) return formattedNumber((size / Math.pow(num, 2)).toFixed(2)) + ' MiB';
    if (size < Math.pow(num, 4)) return formattedNumber((size / Math.pow(num, 3)).toFixed(2)) + ' GiB';
    return formattedNumber((size / Math.pow(num, 4)).toFixed(2)) + ' TiB';
}

export function computeSize2(size: number): string {
    const num = 1000.0;
    if (size < num) return size + ' B';
    if (size < Math.pow(num, 2)) return formattedNumber((size / num).toFixed(2)) + ' KB';
    if (size < Math.pow(num, 3)) return formattedNumber((size / Math.pow(num, 2)).toFixed(2)) + ' MB';
    if (size < Math.pow(num, 4)) return formattedNumber((size / Math.pow(num, 3)).toFixed(2)) + ' GB';
    return formattedNumber((size / Math.pow(num, 4)).toFixed(2)) + ' TB';
}

export function computeCPU(size: number): string {
    const num = 1000;
    if (size < num) return size + ' ns';
    if (size < Math.pow(num, 2)) return formattedNumber((size / num).toFixed(2)) + ' μs';
    if (size < Math.pow(num, 3)) return formattedNumber((size / Math.pow(num, 2)).toFixed(2)) + ' ms';
    return formattedNumber((size / Math.pow(num, 3)).toFixed(2)) + ' s';
}

export function splitSize(size: number): any {
    const num = 1024.0;
    if (size < num) return { size: Number(size), unit: 'B' };
    if (size < Math.pow(num, 2)) return { size: formattedNumber((size / num).toFixed(2)), unit: 'KB' };
    if (size < Math.pow(num, 3))
        return { size: formattedNumber((size / Number(Math.pow(num, 2).toFixed(2))).toFixed(2)), unit: 'MB' };
    if (size < Math.pow(num, 4))
        return { size: formattedNumber((size / Number(Math.pow(num, 3))).toFixed(2)), unit: 'GB' };
    return { size: formattedNumber((size / Number(Math.pow(num, 4))).toFixed(2)), unit: 'TB' };
}

export function computeSizeFromMB(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' MB';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' GB';
    return (size / Math.pow(num, 3)).toFixed(2) + ' TB';
}

export function computeSizeFromKB(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' KB';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' MB';
    if (size < Math.pow(num, 3)) return (size / Math.pow(num, 2)).toFixed(2) + ' GB';
    return (size / Math.pow(num, 3)).toFixed(2) + ' TB';
}

export function computeSizeFromByte(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' B';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' KB';
    if (size < Math.pow(num, 3)) return (size / Math.pow(num, 2)).toFixed(2) + ' MB';
    if (size < Math.pow(num, 4)) return (size / Math.pow(num, 3)).toFixed(2) + ' GB';
    return (size / Math.pow(num, 4)).toFixed(2) + ' TB';
}

export function computeSizeFromKBs(size: number): string {
    const num = 1024.0;
    if (size < num) return size + ' KB/s';
    if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + ' MB/s';
    if (size < Math.pow(num, 3)) return (size / Math.pow(num, 2)).toFixed(2) + ' GB/s';
    return (size / Math.pow(num, 3)).toFixed(2) + ' TB/s';
}
