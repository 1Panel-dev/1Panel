import i18n from '@/lang';

export const getAgentProviderDisplayName = (provider: string, displayName?: string): string => {
    if (provider === 'custom' || displayName === 'Custom') {
        return i18n.global.t('container.custom');
    }
    return displayName || provider;
};
