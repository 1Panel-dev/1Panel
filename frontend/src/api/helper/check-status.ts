import i18n from '@/lang';
import router from '@/routers';
import { MsgError } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

export const checkStatus = (status: number, msg: string): void => {
    const { entrance, isLogin } = useGlobalStore();
    switch (status) {
        case 400:
            MsgError(msg ? msg : i18n.global.t('commons.res.paramError'));
            break;
        case 404:
            MsgError(msg ? msg : i18n.global.t('commons.res.notFound'));
            break;
        case 403:
            isLogin.value = false;
            router.replace({ name: 'entrance', params: { code: entrance.value } });
            MsgError(msg ? msg : i18n.global.t('commons.res.forbidden'));
            break;
        case 500:
            MsgError(msg ? msg : i18n.global.t('commons.res.serverError'));
            break;
        default:
            MsgError(msg ? msg : i18n.global.t('commons.res.commonError'));
    }
};
