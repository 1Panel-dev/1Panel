import i18n from '@/lang';

export const typeOptions = [
    { label: i18n.global.t('cronjob.shell'), value: 'shell' },
    { label: i18n.global.t('cronjob.website'), value: 'website' },
    { label: i18n.global.t('cronjob.database'), value: 'database' },
    { label: i18n.global.t('cronjob.directory'), value: 'directory' },
    { label: i18n.global.t('cronjob.syncDate'), value: 'sync' },
    // { label: i18n.global.t('cronjob.releaseMemory'), value: 'release' },
    { label: i18n.global.t('cronjob.curl') + ' URL', value: 'curl' },
];

export const specOptions = [
    { label: i18n.global.t('cronjob.perMonth'), value: 'perMonth' },
    { label: i18n.global.t('cronjob.perWeek'), value: 'perWeek' },
    { label: i18n.global.t('cronjob.perNDay'), value: 'perNDay' },
    { label: i18n.global.t('cronjob.perNHour'), value: 'perNHour' },
    { label: i18n.global.t('cronjob.perHour'), value: 'perHour' },
    { label: i18n.global.t('cronjob.perNMinute'), value: 'perNMinute' },
];

export const weekOptions = [
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 7 },
];
export const loadWeek = (i: number) => {
    for (const week of weekOptions) {
        if (week.value === i) {
            return week.label;
        }
    }
    return '';
};
export const loadZero = (i: number) => {
    return i < 10 ? '0' + i : '' + i;
};
