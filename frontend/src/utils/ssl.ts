import { AcmeAccountTypes, DNSTypes, KeyTypes } from '@/global/mimetype';
import i18n from '@/lang';

export function getAccountName(type: string) {
    for (const item of AcmeAccountTypes) {
        if (item.value === type) {
            return item.label;
        }
    }
    return '';
}

export function getKeyName(type: string) {
    for (const item of KeyTypes) {
        if (item.value === type) {
            return item.label;
        }
    }
    return '';
}

export function getDNSName(type: string) {
    for (const item of DNSTypes) {
        if (item.value === type) {
            return item.label;
        }
    }
    return '';
}

export function getProvider(provider: string): string {
    switch (provider) {
        case 'dnsAccount':
            return i18n.global.t('website.dnsAccount');
        case 'dnsManual':
            return i18n.global.t('website.dnsManual');
        case 'http':
            return 'HTTP';
        case 'selfSigned':
            return i18n.global.t('ssl.selfSigned');
        case 'fromMaster':
            return i18n.global.t('ssl.fromMaster');
        default:
            return i18n.global.t('ssl.manualCreate');
    }
}
