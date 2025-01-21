import { Login } from '@/api/interface/auth';
import http from '@/api';

export const loginApi = (params: Login.ReqLoginForm) => {
    return http.post<Login.ResLogin>(`/core/auth/login`, params);
};

export const mfaLoginApi = (params: Login.MFALoginForm) => {
    return http.post<Login.ResLogin>(`/core/auth/mfalogin`, params);
};

export const getCaptcha = () => {
    return http.get<Login.ResCaptcha>(`/core/auth/captcha`);
};

export const logOutApi = () => {
    return http.post<any>(`/core/auth/logout`);
};

export const checkIsSafety = (code: string) => {
    return http.get<string>(`/core/auth/issafety?code=${code}`);
};

export const checkIsDemo = () => {
    return http.get<boolean>('/core/auth/demo');
};

export const getLanguage = () => {
    return http.get<string>(`/core/auth/language`);
};

export const checkIsIntl = () => {
    return http.get<boolean>('/core/auth/intl');
};
