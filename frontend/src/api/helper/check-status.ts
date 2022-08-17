import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import router from '@/routers';

/**
 * @description: 校验网络请求状态码
 * @param {Number} status
 * @return void
 */
export const checkStatus = (status: number): void => {
    switch (status) {
        case 400:
            ElMessage.error(i18n.global.t('commons.res.paramError'));
            break;
        case 404:
            ElMessage.error(i18n.global.t('commons.res.notFound'));
            break;
        case 403:
            router.replace({ path: '/login' });
            ElMessage.error(i18n.global.t('commons.res.forbidden'));
            break;
        case 500:
            ElMessage.error(i18n.global.t('commons.res.serverError'));
            break;
        default:
            ElMessage.error(i18n.global.t('commons.res.commonError'));
    }
};
