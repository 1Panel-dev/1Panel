export enum CompressType {
    Zip = 'zip',
    Gz = 'gz',
    Bz2 = 'bz2',
    TarBz2 = 'tar.bz2',
    Tar = 'tar',
    TGz = 'tgz',
    TarGz = 'tar.gz',
    Xz = 'xz',
    TarXz = 'tar.xz',
    Rar = 'rar',
    '7z' = '7z',
}

export enum CompressExtension {
    zip = '.zip',
    gz = '.gz',
    'tar.bz2' = '.tar.bz2',
    bz2 = '.bz2',
    tar = '.tar',
    tgz = '.tgz',
    'tar.gz' = '.tar.gz',
    'tar.xz' = '.tar.xz',
    xz = '.xz',
    rar = '.rar',
    '7z' = '.7z',
}

export const MimetypeByExtensionObject: Record<string, string> = {
    '.zip': 'application/zip',
    '.tar': 'application/x-tar',
    '.tar.bz2': 'application/x-bzip2',
    '.bz2': 'application/x-bzip2',
    '.tar.gz': 'application/gzip',
    '.tgz': 'application/x-gzip',
    '.gz': 'application/gzip',
    '.xz': 'application/x-xz',
    '.rar': 'application/x-rar-compressed',
    '.7z': 'application/x-7z-compressed',
};
