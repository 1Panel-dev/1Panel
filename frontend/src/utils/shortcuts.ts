import i18n from '@/lang';

export const shortcuts = [
    {
        text: i18n.global.t('monitor.today'),
        value: () => {
            const end = new Date();
            const start = new Date(new Date().setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.yesterday'),
        value: () => {
            const yesterday = new Date(new Date().getTime() - 3600 * 1000 * 24 * 1);
            const end = new Date(yesterday.setHours(23, 59, 59, 999));
            const start = new Date(yesterday.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [3]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 3);
            const end = new Date();
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [7]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 7);
            const end = new Date();
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [30]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 30);
            const end = new Date();
            return [start, end];
        },
    },
];
