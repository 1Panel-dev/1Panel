import i18n from '@/lang';
import router from '@/routers';
import { MsgError } from '@/utils/message';

/**
 * @description: 校验网络请求状态码
 * @param {Number} status
 * @return void
 */
export const checkStatus = (status: number): void => {
    switch (status) {
        case 400:
            MsgError(i18n.global.t('commons.res.paramError'));
            break;
        case 404:
            MsgError(i18n.global.t('commons.res.notFound'));
            break;
        case 403:
            router.replace({ path: '/login' });
            MsgError(i18n.global.t('commons.res.forbidden'));
            break;
        case 500:
            MsgError(i18n.global.t('commons.res.serverError'));
            break;
        default:
            MsgError(i18n.global.t('commons.res.commonError'));
    }
};
