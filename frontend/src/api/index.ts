import axios, { AxiosInstance, AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import { showFullScreenLoading, tryHideFullScreenLoading } from '@/config/service-loading';
import { AxiosCanceler } from './helper/axios-cancel';
import { ResultData } from '@/api/interface';
import { ResultEnum } from '@/enums/http-enum';
import { checkStatus } from './helper/check-status';
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
    baseURL: import.meta.env.VITE_API_URL as string,
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

        this.service.interceptors.response.use(
            (response: AxiosResponse) => {
                const { data, config } = response;
                axiosCanceler.removePending(config);
                tryHideFullScreenLoading();
                if (data.code == ResultEnum.OVERDUE) {
                    ElMessage.error(data.msg);
                    router.replace({
                        path: '/login',
                    });
                    return Promise.reject(data);
                }
                if (data.code && data.code !== ResultEnum.SUCCESS) {
                    ElMessage.error(data.msg);
                    return Promise.reject(data);
                }
                return data;
            },
            async (error: AxiosError) => {
                const { response } = error;
                tryHideFullScreenLoading();
                if (error.message.indexOf('timeout') !== -1) ElMessage.error('请求超时！请您稍后重试');
                if (response) checkStatus(response.status);
                if (!window.navigator.onLine) router.replace({ path: '/500' });
                return Promise.reject(error);
            },
        );
    }

    get<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
        return this.service.get(url, { params, ..._object });
    }
    post<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
        return this.service.post(url, params, _object);
    }
    put<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
        return this.service.put(url, params, _object);
    }
    delete<T>(url: string, params?: any, _object = {}): Promise<ResultData<T>> {
        return this.service.delete(url, { params, ..._object });
    }
}

export default new RequestHttp(config);
