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
        label: 'plaintext',
        value: ['txt'],
    },
    {
        label: 'json',
        value: ['json'],
    },
    {
        label: 'markdown',
        value: ['md'],
    },
    {
        label: 'yaml',
        value: ['yml', 'yaml'],
    },
    {
        label: 'php',
        value: ['php'],
    },
    {
        label: 'sql',
        value: ['sql'],
    },
    {
        label: 'go',
        value: ['go'],
    },
    {
        label: 'html',
        value: ['html'],
    },
    {
        label: 'javascript',
        value: ['js'],
    },
    {
        label: 'java',
        value: ['java'],
    },
    {
        label: 'kotlin',
        value: ['kt'],
    },
    {
        label: 'python',
        value: ['py'],
    },
    {
        label: 'redis',
        value: ['redis'],
    },
    {
        label: 'shell',
        value: ['sh'],
    },
    {
        label: 'css',
        value: ['css'],
    },
];
