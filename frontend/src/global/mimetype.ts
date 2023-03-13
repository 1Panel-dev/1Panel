import { CompressType } from '@/enums/files';

export const Mimetypes = new Map([
    ['application/zip', CompressType.Zip],
    ['application/x-zip', CompressType.Zip],
    ['application/x-zip-compressed', CompressType.Zip],
    ['application/x-tar', CompressType.Tar],
    ['application/x-bzip2', CompressType.Bz2],
    ['application/gzip', CompressType.TarGz],
    ['application/x-gzip', CompressType.TarGz],
    ['application/x-gunzip', CompressType.TarGz],
    ['application/gzipped', CompressType.TarGz],
    ['application/gzip-compressed', CompressType.TarGz],
    ['application/x-gzip-compressed', CompressType.TarGz],
    ['gzip/document', CompressType.TarGz],
    ['application/x-xz', CompressType.Xz],
]);

export const Languages = [
    {
        label: 'go',
        value: 'go',
    },
    {
        label: 'html',
        value: 'html',
    },
    {
        label: 'javascript',
        value: 'javascript',
    },
    {
        label: 'java',
        value: 'java',
    },
    {
        label: 'kotlin',
        value: 'kotlin',
    },
    {
        label: 'markdown',
        value: 'markdown',
    },
    {
        label: 'mysql',
        value: 'mysql',
    },
    {
        label: 'php',
        value: 'php',
    },
    {
        label: 'redis',
        value: 'redis',
    },
    {
        label: 'shell',
        value: 'shell',
    },
    {
        label: 'sql',
        value: 'sql',
    },
    {
        label: 'yaml',
        value: 'yaml',
    },
    {
        label: 'json',
        value: 'json',
    },
    {
        label: 'css',
        value: 'css',
    },
];
