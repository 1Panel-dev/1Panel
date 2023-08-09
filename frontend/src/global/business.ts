import router from '@/routers';

export function canEditPort(appKey: string): boolean {
    const apps = ['openresty', 'php', 'frpc', 'frps', 'ddns-go', 'home-assistant'];
    return !apps.includes(appKey);
}

export function toFolder(folder: string) {
    router.push({ path: '/hosts/files', query: { path: folder } });
}
