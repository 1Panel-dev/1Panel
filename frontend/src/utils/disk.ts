import i18n from '@/lang';
import { Dashboard } from '@/api/interface/dashboard';

// A mount whose usage could not be loaded keeps its row, with errorType set and zero sizes.
// Agents older than that field report such a mount as a row of zeros only.
export const isDiskFailed = (disk: Dashboard.DiskInfo) => !!disk.errorType || !disk.total;

export const diskErrorLabel = (disk: Dashboard.DiskInfo) => {
    switch (disk.errorType) {
        case 'timeout':
            return i18n.global.t('home.diskErrTimeout');
        case 'denied':
            return i18n.global.t('home.diskErrDenied');
        case 'missing':
            return i18n.global.t('home.diskErrMissing');
        default:
            return i18n.global.t('home.diskErrOther');
    }
};

// The raw error only adds something when the reason has no wording of its own.
export const diskErrorDetail = (disk: Dashboard.DiskInfo) => {
    const label = diskErrorLabel(disk);
    const known = ['timeout', 'denied', 'missing'].includes(disk.errorType || '');
    return !known && disk.error ? `${label}: ${disk.error}` : label;
};
