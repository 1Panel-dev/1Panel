import { Login } from '@/api/interface/index';
import http from '@/api';

export const loginApi = (params: Login.ReqLoginForm) => {
    return http.post(`/auth/login`, params);
};

export const getCaptcha = () => {
    return http.get(`/auth/captcha`);
};
