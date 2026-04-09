import type { RouteRecordRaw } from 'vue-router';

export interface EditionFrontendProvider {
    name: string;
    routes: RouteRecordRaw[];
    loadStyles?: () => Promise<void>;
}
