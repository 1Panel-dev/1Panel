import { Login } from '@/api/interface/auth';
import http from '@/api';
import { deepCopy } from '@/utils/misc';
import { Base64 } from 'js-base64';

export const loginApi = (params: Login.ReqLoginForm) => {
    return http.post<Login.ResLogin>(`/core/auth/login`, params);
};

export const mfaLoginApi = (params: Login.MFALoginForm) => {
    return http.post<Login.ResLogin>(`/core/auth/mfalogin`, params);
};

export const passkeyBeginApi = () => {
    return http.post<Login.PasskeyBeginResponse>(`/core/auth/passkey/begin`);
};

export const passkeyFinishApi = (params: Record<string, any>, sessionId: string) => {
    return http.post<Login.ResLogin>(`/core/auth/passkey/finish`, params, undefined, { 'Passkey-Session': sessionId });
};

export const getCaptcha = () => {
    return http.get<Login.ResCaptcha>(`/core/auth/captcha`);
};

export const logOutApi = () => {
    return http.post<any>(`/core/auth/logout`);
};

export const getLoginSetting = () => {
    return http.get<Login.LoginSetting>('/core/auth/setting');
};

export const getWelcomePage = () => {
    return http.get<string>('/core/auth/welcome');
};

export const getAuthInfo = () => {
    return http.get<Login.AuthInfo>('/core/xpackee/users/info');
};
export const updateAuthInfo = (params: Login.AuthInfoUpdate) => {
    let request = deepCopy(params) as Login.AuthInfoUpdate;
    if (request.oldPassword) {
        request.oldPassword = Base64.encode(request.oldPassword);
    }
    if (request.newPassword) {
        request.newPassword = Base64.encode(request.newPassword);
    }
    request.retryPassword = '';
    return http.post<any>('/core/xpackee/users/info/update', request);
};
