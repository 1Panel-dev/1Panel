import axios, { AxiosInstance, AxiosError, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import { ResultData } from '@/api/interface';
import { ResultEnum } from '@/enums/http-enum';
import { checkStatus } from './helper/check-status';
import router from '@/routers';
import { GlobalStore } from '@/store';
import { MsgError } from '@/utils/message';
import { Base64 } from 'js-base64';

const globalStore = GlobalStore();

const config = {
    baseURL: import.meta.env.VITE_API_URL as string,
    timeout: ResultEnum.TIMEOUT as number,
    withCredentials: true,
};

class RequestHttp {
    service: AxiosInstance;
    public constructor(config: AxiosRequestConfig) {
        this.service = axios.create(config);
        this.service.interceptors.request.use(
            (config: AxiosRequestConfig) => {
                let language = globalStore.language === 'tw' ? 'zh-Hant' : globalStore.language;
                config.headers = {
                    'Accept-Language': language,
                    ...config.headers,
                };
                if (config.url === '/auth/login' || config.url === '/auth/mfalogin') {
                    let entrance = Base64.encode(globalStore.entrance);
                    config.headers.EntranceCode = entrance;
                }
                return {
                    ...config,
                } as InternalAxiosRequestConfig<any>;
            },
            (error: AxiosError) => {
                return Promise.reject(error);
            },
        );

        this.service.interceptors.response.use(
            (response: AxiosResponse) => {
                globalStore.errStatus = '';
                const { data } = response;
                if (data.code == ResultEnum.OVERDUE || data.code == ResultEnum.FORBIDDEN) {
                    globalStore.setLogStatus(false);
                    router.push({
                        name: 'entrance',
                        params: { code: globalStore.entrance },
                    });
                    return Promise.reject(data);
                }
                if (data.code == ResultEnum.NOTFOUND) {
                    globalStore.errStatus = 'err-found';
                    return;
                }
                if (data.code == ResultEnum.ERRIP) {
                    globalStore.errStatus = 'err-ip';
                    return;
                }
                if (data.code == ResultEnum.ERRDOMAIN) {
                    globalStore.errStatus = 'err-domain';
                    return;
                }
                if (data.code == ResultEnum.UNSAFETY) {
                    globalStore.errStatus = 'err-unsafe';
                    return;
                }
                if (data.code == ResultEnum.EXPIRED) {
                    router.push({ name: 'Expired' });
                    return;
                }
                if (data.code == ResultEnum.ERRXPACK) {
                    globalStore.isProductPro = false;
                    window.location.reload();
                    return Promise.reject(data);
                }
                if (data.code == ResultEnum.ERRGLOBALLOADDING) {
                    globalStore.setGlobalLoading(true);
                    globalStore.setLoadingText(data.message);
                    return;
                } else {
                    if (globalStore.isLoading) {
                        globalStore.setGlobalLoading(false);
                    }
                }
                if (data.code == ResultEnum.ERRAUTH) {
                    return data;
                }
                if (data.code && data.code !== ResultEnum.SUCCESS) {
                    MsgError(data.message);
                    return Promise.reject(data);
                }
                return data;
            },
            async (error: AxiosError) => {
                globalStore.errStatus = '';
                const { response } = error;
                if (error.message.indexOf('timeout') !== -1) MsgError('请求超时！请您稍后重试');
                if (response) {
                    switch (response.status) {
                        case 310:
                            globalStore.errStatus = 'err-ip';
                            router.push({
                                name: 'entrance',
                                params: { code: globalStore.entrance },
                            });
                            return;
                        case 311:
                            globalStore.errStatus = 'err-domain';
                            router.push({
                                name: 'entrance',
                                params: { code: globalStore.entrance },
                            });
                            return;
                        case 312:
                            globalStore.errStatus = 'err-entrance';
                            router.push({
                                name: 'entrance',
                                params: { code: globalStore.entrance },
                            });
                            return;
                        case 313:
                            router.push({ name: 'Expired' });
                            return;
                        case 500:
                        case 407:
                            checkStatus(
                                response.status,
                                response.data && response.data['message'] ? response.data['message'] : '',
                            );
                            return Promise.reject(error);
                        default:
                            globalStore.isLogin = false;
                            globalStore.errStatus = 'code-' + response.status;
                            router.push({
                                name: 'entrance',
                                params: { code: globalStore.entrance },
                            });
                            return Promise.reject(error);
                    }
                }
                if (!window.navigator.onLine) router.replace({ path: '/500' });
                return Promise.reject(error);
            },
        );
    }

    get<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
        return this.service.get(url, { params, ..._object });
    }
    post<T>(url: string, params?: object, timeout?: number): Promise<ResultData<T>> {
        return this.service.post(url, params, {
            baseURL: import.meta.env.VITE_API_URL as string,
            timeout: timeout ? timeout : (ResultEnum.TIMEOUT as number),
            withCredentials: true,
        });
    }
    put<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
        return this.service.put(url, params, _object);
    }
    delete<T>(url: string, params?: any, _object = {}): Promise<ResultData<T>> {
        return this.service.delete(url, { params, ..._object });
    }
    download<BlobPart>(url: string, params?: object, _object = {}): Promise<BlobPart> {
        return this.service.post(url, params, _object);
    }
    upload<T>(url: string, params: object = {}, config?: AxiosRequestConfig): Promise<T> {
        return this.service.post(url, params, config);
    }
}

export default new RequestHttp(config);
