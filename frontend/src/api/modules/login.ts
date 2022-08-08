import { Login } from '@/api/interface/index';
import { PORT1 } from '@/api/config/servicePort';
// import qs from 'qs';

import http from '@/api';

/**
 * @name 登录模块
 */
// * 用户登录接口
export const loginApi = (params: Login.ReqLoginForm) => {
    console.log(params);
    return http.post<Login.ResLogin>(PORT1 + `base/login`, params); // 正常 post json 请求  ==>  application/json
    // return http.post<Login.ResLogin>(PORT1 + `/login`, {}, { params }); // post 请求携带 query 参数  ==>  ?username=admin&password=123456
    // return http.post<Login.ResLogin>(PORT1 + `/login`, qs.stringify(params)); // post 请求携带 表单 参数  ==>  application/x-www-form-urlencoded
    // return http.post<Login.ResLogin>(PORT1 + `/login`, params, {
    //     headers: { noLoading: true },
    // }); // 控制当前请求不显示 loading

    // return { data: { access_token: '565656565' } };
};

// // * 获取按钮权限
// export const getAuthButtons = () => {
//     return http.get<Login.ResAuthButtons>(PORT1 + `/auth/buttons`);
// };

// * 获取验证码
export const getCaptcha = () => {
    return http.post<Login.ResCaptcha>(PORT1 + `base/captcha`);
};
