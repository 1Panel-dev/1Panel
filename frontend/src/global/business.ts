import router from '@/routers';

export function canEditPort(app: any): boolean {
    if (app.key == 'openresty') {
        return false;
    }
    if (app.type == 'php') {
        return false;
    }
    return true;
}

export function toFolder(folder: string) {
    router.push({ path: '/hosts/files', query: { path: folder } });
}
