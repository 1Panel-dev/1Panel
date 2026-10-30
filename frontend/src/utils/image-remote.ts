import type { Setting } from '@/api/interface/setting';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { routerToName } from '@/utils/router';

export const showImageProxyRequired = async (messageKey: string, isCurrent: () => boolean) => {
    try {
        await ElMessageBox.confirm(i18n.global.t(messageKey), '', {
            confirmButtonText: i18n.global.t('container.imageCheckUpdateConfigureProxy'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        });
        if (isCurrent()) routerToName('Panel');
    } catch {
        // Cancel leaves the current page unchanged.
    }
};

// Basic settings expose the environment and proxy status without proxy credentials.
export const isRuntimeProxyRequired = (settings: Pick<Setting.SettingBaseInfo, 'edition' | 'proxyEnabled'>): boolean =>
    settings.edition !== 'intl' && !settings.proxyEnabled;
