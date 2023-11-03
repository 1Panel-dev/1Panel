import router from '@/routers';

export function toFolder(folder: string) {
    router.push({ path: '/hosts/files', query: { path: folder } });
}
