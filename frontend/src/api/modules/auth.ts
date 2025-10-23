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

export const getLoginSetting = () => {
    return http.get<Login.LoginSetting>('/core/auth/setting');
};

export const getWelcomePage = () => {
    return http.get<string>('/core/auth/welcome');
};
