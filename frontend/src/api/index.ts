import axios, {
    AxiosInstance,
    AxiosError,
    AxiosRequestConfig,
    AxiosResponse,
} from 'axios';
import {
    showFullScreenLoading,
    tryHideFullScreenLoading,
} from '@/config/serviceLoading';
import { AxiosCanceler } from './helper/axiosCancel';
import { ResultData } from '@/api/interface';
import { ResultEnum } from '@/enums/httpEnum';
import { checkStatus } from './helper/checkStatus';
import { ElMessage } from 'element-plus';
import router from '@/routers';

/**
 * pinia 错误使用说明示例
 * https://github.com/vuejs/pinia/discussions/971
 * https://github.com/vuejs/pinia/discussions/664#discussioncomment-1329898
 * https://pinia.vuejs.org/core-concepts/outside-component-usage.html#single-page-applications
 */
// const globalStore = GlobalStore();

const axiosCanceler = new AxiosCanceler();

const config = {
    // 默认地址请求地址，可在 .env 开头文件中修改
    baseURL: import.meta.env.VITE_API_URL as string,
    // 设置超时时间（10s）
    timeout: ResultEnum.TIMEOUT as number,
    // 跨域时候允许携带凭证
    withCredentials: true,
};

class RequestHttp {
    service: AxiosInstance;
    public constructor(config: AxiosRequestConfig) {
        this.service = axios.create(config);
        this.service.interceptors.request.use(
            (config: AxiosRequestConfig) => {
                axiosCanceler.addPending(config);
                config.headers!.noLoading || showFullScreenLoading();
                return {
                    ...config,
                };
            },
            (error: AxiosError) => {
                return Promise.reject(error);
            },
        );

        /**
         * @description 响应拦截器
         *  服务器换返回信息 -> [拦截统一处理] -> 客户端JS获取到信息
         */
        this.service.interceptors.response.use(
            (response: AxiosResponse) => {
                const { data, config } = response;
                // * 在请求结束后，移除本次请求，并关闭请求 loading
                axiosCanceler.removePending(config);
                tryHideFullScreenLoading();
                // * 登陆失效（code == 599）
                if (data.code == ResultEnum.OVERDUE) {
                    ElMessage.error(data.msg);
                    router.replace({
                        path: '/login',
                    });
                    return Promise.reject(data);
                }
                // * 全局错误信息拦截（防止下载文件得时候返回数据流，没有code，直接报错）
                if (data.code && data.code !== ResultEnum.SUCCESS) {
                    ElMessage.error(data.msg);
                    return Promise.reject(data);
                }
                // * 成功请求（在页面上除非特殊情况，否则不用处理失败逻辑）
                return data;
            },
            async (error: AxiosError) => {
                const { response } = error;
                tryHideFullScreenLoading();
                // 请求超时单独判断，因为请求超时没有 response
                if (error.message.indexOf('timeout') !== -1)
                    ElMessage.error('请求超时！请您稍后重试');
                // 根据响应的错误状态码，做不同的处理
                if (response) checkStatus(response.status);
                // 服务器结果都没有返回(可能服务器错误可能客户端断网)，断网处理:可以跳转到断网页面
                if (!window.navigator.onLine) router.replace({ path: '/500' });
                return Promise.reject(error);
            },
        );
    }

    // * 常用请求方法封装
    get(url: string, params?: object, _object = {}): Promise<ResultData> {
        return this.service.get(url, { params, ..._object });
    }
    post(url: string, params?: object, _object = {}): Promise<ResultData> {
        return this.service.post(url, params, _object);
    }
    put(url: string, params?: object, _object = {}): Promise<ResultData> {
        return this.service.put(url, params, _object);
    }
    delete(url: string, params?: any, _object = {}): Promise<ResultData> {
        return this.service.delete(url, { params, ..._object });
    }
}

export default new RequestHttp(config);
