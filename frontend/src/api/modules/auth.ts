import { Login } from '@/api/interface/auth';
import http from '@/api';

export const loginApi = (params: Login.ReqLoginForm) => {
    return http.post<Login.ResLogin>(`/auth/login`, params);
};

export const mfaLoginApi = (params: Login.MFALoginForm) => {
    return http.post<Login.ResLogin>(`/auth/mfalogin`, params);
};

export const getCaptcha = () => {
    return http.get<Login.ResCaptcha>(`/auth/captcha`);
};

export const logOutApi = () => {
    return http.post<any>(`/auth/logout`);
};

export const entrance = (code: string) => {
    return http.get<any>(`/${code}`);
};

export const loginStatus = () => {
    return http.get<any>('/info');
};

export const checkIsFirst = () => {
    return http.get<boolean>('/auth/status');
};
export const initUser = (params: Login.InitUser) => {
    return http.post(`/auth/init`, params);
};
