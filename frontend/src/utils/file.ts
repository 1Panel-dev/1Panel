import { Languages } from '@/global/mimetype';

const icons = new Map([
    ['.zip', 'p-file-zip'],
    ['.gz', 'p-file-zip'],
    ['.tar.bz2', 'p-file-zip'],
    ['.bz2', 'p-file-zip'],
    ['.xz', 'p-file-zip'],
    ['.tar.xz', 'p-file-zip'],
    ['.tar', 'p-file-zip'],
    ['.tar.gz', 'p-file-zip'],
    ['.war', 'p-file-zip'],
    ['.tgz', 'p-file-zip'],
    ['.7z', 'p-file-zip'],
    ['.rar', 'p-file-zip'],
    ['.mp3', 'p-file-mp3'],
    ['.svg', 'p-file-svg'],
    ['.txt', 'p-file-txt'],
    ['.html', 'p-file-html'],
    ['.word', 'p-file-word'],
    ['.ppt', 'p-file-ppt'],
    ['.jpg', 'p-file-jpg'],
    ['.jpeg', 'p-file-jpg'],
    ['.png', 'p-file-png'],
    ['.xlsx', 'p-file-excel'],
    ['.doc', 'p-file-word'],
    ['.xls', 'p-file-excel'],
    ['.docx', 'p-file-word'],
    ['.pdf', 'p-file-pdf'],
    ['.bmp', 'p-file-png'],
    ['.gif', 'p-file-png'],
    ['.tiff', 'p-file-png'],
    ['.ico', 'p-file-png'],
    ['.webp', 'p-file-png'],
    ['.mp4', 'p-file-video'],
    ['.webm', 'p-file-video'],
    ['.mov', 'p-file-video'],
    ['.wmv', 'p-file-video'],
    ['.mkv', 'p-file-video'],
    ['.avi', 'p-file-video'],
    ['.wma', 'p-file-video'],
    ['.flv', 'p-file-video'],
    ['.wav', 'p-file-mp3'],
    ['.wma', 'p-file-mp3'],
    ['.ape', 'p-file-mp3'],
    ['.acc', 'p-file-mp3'],
    ['.ogg', 'p-file-mp3'],
    ['.flac', 'p-file-mp3'],
]);

const fileTypes = {
    image: ['.jpg', '.jpeg', '.png', '.bmp', '.gif', '.tiff', '.ico', '.svg', '.webp'],
    compress: ['.zip', '.rar', '.gz', '.war', '.tgz', '.7z', '.tar.gz', '.tar', '.bz2', '.xz', '.tar.bz2', '.tar.xz'],
    video: ['.mp4', '.webm', '.mov', '.wmv', '.mkv', '.avi', '.wma', '.flv'],
    audio: ['.mp3', '.wav', '.wma', '.ape', '.acc', '.ogg', '.flac'],
    pdf: ['.pdf'],
    word: ['.doc', '.docx'],
    excel: ['.xls', '.xlsx'],
    text: ['.iso', '.tiff', '.exe', '.so', '.bz', '.dmg', '.apk', '.pptx', '.ppt', '.xlsb'],
};

export function getIcon(extension: string): string {
    if (icons.get(extension) != undefined) {
        return String(icons.get(extension));
    }
    return 'p-file-normal';
}

export const getFileType = (extension: string) => {
    let type = 'text';
    Object.entries(fileTypes).forEach(([key, extensions]) => {
        if (extensions.includes(extension.toLowerCase())) {
            type = key;
        }
    });
    return type;
};

const convertTypes = ['image', 'video', 'audio'] as const;
type ConvertType = (typeof convertTypes)[number];

export function isConvertible(extension: string, mimeType: string): boolean {
    return convertTypes.includes(getFileType(extension) as ConvertType) && /^(image|audio|video)\//.test(mimeType);
}

export const isSensitiveLinuxPath = (path: string) => {
    const sensitivePath = [
        '/',
        '/bin',
        '/sbin',
        '/usr/bin',
        '/usr/sbin',
        '/usr/local/bin',
        '/etc',
        '/lib',
        '/lib64',
        '/usr/lib',
        '/home',
        '/tmp',
        '/var',
        '/dev',
        '/proc',
        '/sys',
    ];
    return sensitivePath.indexOf(path) !== -1;
};

function buildFileDownloadUrl(filePath: string, currentNode: string): string {
    const base = `${import.meta.env.VITE_API_URL as string}/files/download?operateNode=${currentNode}&`;
    return base + 'path=' + encodeURIComponent(filePath);
}

export function buildFileSharePageUrl(code: string, currentNode: string): string {
    const shareUrl = new URL(`/s/${encodeURIComponent(code)}`, window.location.origin);
    shareUrl.searchParams.set('operateNode', currentNode);
    return shareUrl.toString();
}

export function buildFileShareDownloadUrl(code: string, currentNode: string, password?: string): string {
    const apiBase = import.meta.env.VITE_API_URL as string;
    const normalizedBase = apiBase.replace(/\/$/, '');
    const shareUrl = /^https?:\/\//i.test(normalizedBase)
        ? new URL(`${normalizedBase}/files/share/download`)
        : new URL(`${normalizedBase}/files/share/download`, window.location.origin);
    shareUrl.searchParams.set('operateNode', currentNode);
    shareUrl.searchParams.set('code', code);
    if (password && password.length > 0) {
        shareUrl.searchParams.set('password', password);
    }
    return shareUrl.toString();
}

export function buildFileShareQrCodeUrl(code: string, currentNode: string): string {
    const apiBase = import.meta.env.VITE_API_URL as string;
    const normalizedBase = apiBase.replace(/\/$/, '');
    const shareUrl = /^https?:\/\//i.test(normalizedBase)
        ? new URL(`${normalizedBase}/files/share/qrcode`)
        : new URL(`${normalizedBase}/files/share/qrcode`, window.location.origin);
    shareUrl.searchParams.set('operateNode', currentNode);
    shareUrl.searchParams.set('code', code);
    return shareUrl.toString();
}

export function downloadFile(filePath: string, currentNode: string) {
    window.open(buildFileDownloadUrl(filePath, currentNode), '_blank');
}

export function downloadWithContent(content: string, fileName: string) {
    const downloadUrl = window.URL.createObjectURL(new Blob([content]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = fileName;
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
}

const editorLanguages = new Map(
    Languages.flatMap((language) => language.value.map((ext) => [ext.toLowerCase(), language.label] as const)),
);

const specialEditorFileNames = ['dockerfile'];

const normalizeEditorExtension = (value: string) => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) {
        return '';
    }
    return trimmed.startsWith('.') ? trimmed.slice(1) : trimmed;
};

const isSpecialEditorFileName = (value: string) => {
    const normalized = value.trim().toLowerCase();
    if (!normalized) {
        return false;
    }
    return specialEditorFileNames.some((filename) => normalized === filename);
};

export const resolveEditorLanguage = (path: string, extension = '', name = '') => {
    if (isSpecialEditorFileName(name)) {
        return 'dockerfile';
    }

    const candidates = [extension, path.split('/').pop() || '']
        .map((value) => normalizeEditorExtension(value))
        .filter(Boolean);

    for (const ext of candidates) {
        const language = editorLanguages.get(ext);
        if (language) {
            return language;
        }
    }

    const fileName = path.split('/').pop() || '';
    if (isSpecialEditorFileName(fileName)) {
        return 'dockerfile';
    }

    return 'yaml';
};
