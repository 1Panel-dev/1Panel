import { ResultEnum } from '@/enums/http-enum';
import i18n from '@/lang';
import router from '@/routers';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgError } from '@/utils/message';

export type AuthResponseAction = 'reject' | 'return';

export interface AuthResponseResult {
    handled: boolean;
    action?: AuthResponseAction;
    message: string;
}

interface AuthResponseOptions {
    showRBACMessage?: boolean;
}

const forbiddenMessage = () => i18n.global.t('commons.res.forbidden');

export const redirectToEntrance = () => {
    const { entrance, isLogin } = useGlobalStore();
    isLogin.value = false;
    router.push({
        name: 'entrance',
        params: { code: entrance.value },
    });
};

export const redirectToExpired = () => {
    router.push({ name: 'Expired' });
};

export const handleAuthResponseCode = (data: any, options: AuthResponseOptions = {}): AuthResponseResult => {
    const message = data?.message || forbiddenMessage();
    switch (data?.code) {
        case ResultEnum.OVERDUE:
        case ResultEnum.FORBIDDEN:
            redirectToEntrance();
            return { handled: true, action: 'reject', message };
        case ResultEnum.ERR_RBAC:
            if (options.showRBACMessage) {
                MsgError(message);
            }
            return { handled: true, action: 'reject', message };
        case ResultEnum.EXPIRED:
            redirectToExpired();
            return { handled: true, action: 'return', message };
        default:
            return { handled: false, message: '' };
    }
};

export const handleAuthResponseStatus = (status: number): AuthResponseResult => {
    switch (status) {
        case ResultEnum.OVERDUE:
        case ResultEnum.FORBIDDEN:
            redirectToEntrance();
            return { handled: true, action: 'reject', message: forbiddenMessage() };
        case ResultEnum.EXPIRED:
            redirectToExpired();
            return { handled: true, action: 'return', message: forbiddenMessage() };
        default:
            return { handled: false, message: '' };
    }
};
