import http from '@/api';
import { ResPage } from '../interface';
import { Container } from '../interface/container';

export const getContainerPage = (params: Container.ContainerSearch) => {
    return http.post<ResPage<Container.ContainerInfo>>(`/containers/search`, params);
};

export const getContainerLog = (params: Container.ContainerLogSearch) => {
    return http.post<string>(`/containers/log`, params);
};

export const ContainerOperator = (params: Container.ContainerOperate) => {
    return http.post(`/containers/operate`, params);
};

export const getContainerInspect = (containerID: string) => {
    return http.get<string>(`/containers/detail/${containerID}`);
};
