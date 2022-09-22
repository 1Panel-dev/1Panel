import i18n from '@/lang';

export const loadBackupName = (type: string) => {
    switch (type) {
        case 'OSS':
            return i18n.global.t('setting.OSS');
            break;
        case 'S3':
            return i18n.global.t('setting.S3');
            break;
        case 'LOCAL':
            return i18n.global.t('setting.serverDisk');
            break;
        default:
            return type;
    }
};
