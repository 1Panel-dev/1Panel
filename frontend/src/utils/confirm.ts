import { ElMessageBox, type ElMessageBoxOptions } from 'element-plus';
import i18n from '@/lang';

// Keep cancellation as a rejection and leave the business operation to the caller.
export const confirmAction = (message: string, title: string, options: ElMessageBoxOptions = {}) => {
    return ElMessageBox.confirm(message, title, {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        ...options,
    });
};
