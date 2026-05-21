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

export const getUserInfo = () => {
    return http.get<Login.AuthInfo>('/core/auth/current');
};
export const updateUserInfo = (params: Login.AuthInfoUpdate) => {
    let request = deepCopy(params) as Login.AuthInfoUpdate;
    if (request.oldPassword) {
        request.oldPassword = Base64.encode(request.oldPassword);
    }
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    return http.post<any>('/core/auth/current/update', request);
};

export const loadMFA = (params: Login.MFARequest) => {
    return http.post<Login.MFAInfo>(`/core/auth/mfa`, params);
};
export const bindMFA = (params: Login.MFABind) => {
    return http.post(`/core/auth/mfa/bind`, params);
};
export const closeMFA = () => {
    return http.post(`/core/auth/mfa/close`);
};
export const generateApiKey = () => {
    return http.post<string>(`/core/auth/api/generate`);
};
export const updateApiConfig = (param: Login.ApiConfig) => {
    return http.post(`/core/auth/api/update`, param);
};

export const passkeyRegisterBegin = (param: { name: string }) => {
    return http.post<Login.PasskeyBeginResponse>(`/core/auth/passkey/register/begin`, param);
};
export const passkeyRegisterFinish = (param: Record<string, any>, sessionId: string) => {
    return http.post(`/core/auth/passkey/register/finish`, param, undefined, { 'Passkey-Session': sessionId });
};
export const passkeyList = () => {
    return http.get<Array<{ id: string; name: string; createdAt: string; lastUsedAt: string }>>(
        `/core/auth/passkey/list`,
    );
};
export const passkeyDelete = (id: string) => {
    return http.post(`/core/auth/passkey/del`, { id });
};

export const handleExpired = (param: Login.PasswordUpdate) => {
    return http.post(`/core/auth/expired/reset`, param);
};
