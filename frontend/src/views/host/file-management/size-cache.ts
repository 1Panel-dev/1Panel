import type { File } from '@/api/interface/file';

type FileSizeItem = Pick<File.File, 'path' | 'size' | 'isDir'> & Partial<Pick<File.File, 'dirSize'>>;

export type DirSizeItem = Pick<File.DepthDirSizeRes, 'path' | 'size'>;

export const updateDirSizeCache = (cache: Map<string, number>, items: DirSizeItem[]) => {
    items.forEach((item) => {
        cache.set(item.path, item.size);
    });
};

export const mergeDirSizeCache = <T extends FileSizeItem>(items: T[], cache: Map<string, number>): T[] => {
    if (cache.size === 0) {
        return items;
    }
    return items.map((item) => {
        if (!item.isDir || !cache.has(item.path)) {
            return item;
        }
        return { ...item, dirSize: cache.get(item.path)! };
    });
};

export const sortFilesByDisplaySize = <T extends FileSizeItem>(items: T[], order?: string | null): T[] => {
    if (order !== 'ascending' && order !== 'descending') {
        return items;
    }
    const direction = order === 'ascending' ? 1 : -1;
    return [...items].sort((prev, next) => {
        const prevSize = prev.isDir && prev.dirSize !== undefined ? prev.dirSize : prev.size;
        const nextSize = next.isDir && next.dirSize !== undefined ? next.dirSize : next.size;
        return (prevSize - nextSize) * direction;
    });
};
