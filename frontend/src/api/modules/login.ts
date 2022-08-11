import { Login } from '@/api/interface/index';
import http from '@/api';

export const loginApi = (params: Login.ReqLoginForm) => {
    return http.post<Login.ResLogin>(`/auth/login`, params);
};

export const getCaptcha = () => {
    return http.get<Login.ResCaptcha>(`/auth/captcha`);
};
