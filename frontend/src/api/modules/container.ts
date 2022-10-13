import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Container } from '../interface/container';

export const getContainerPage = (params: ReqPage) => {
    return http.post<ResPage<Container.ContainerInfo>>(`/containers/search`, params);
};
export const createContainer = (params: Container.ContainerCreate) => {
    return http.post(`/containers`, params);
};

export const getContainerLog = (params: Container.ContainerLogSearch) => {
    return http.post<string>(`/containers/log`, params);
};
export const ContainerStats = (id: string) => {
    return http.get<Container.ContainerStats>(`/containers/stats/${id}`);
};

export const ContainerOperator = (params: Container.ContainerOperate) => {
    return http.post(`/containers/operate`, params);
};

export const inspect = (params: Container.ContainerInspect) => {
    return http.post<string>(`/containers/inspect`, params);
};

// image
export const getImagePage = (params: ReqPage) => {
    return http.post<ResPage<Container.ImageInfo>>(`/containers/image/search`, params);
};
export const imageOptions = () => {
    return http.get<Array<Container.Options>>(`/containers/image`);
};
export const imageBuild = (params: Container.ImageBuild) => {
    return http.post<string>(`/containers/image/build`, params);
};
export const imagePull = (params: Container.ImagePull) => {
    return http.post(`/containers/image/pull`, params);
};
export const imagePush = (params: Container.ImagePush) => {
    return http.post(`/containers/image/push`, params);
};
export const imageLoad = (params: Container.ImageLoad) => {
    return http.post(`/containers/image/load`, params);
};
export const imageSave = (params: Container.ImageSave) => {
    return http.post(`/containers/image/save`, params);
};
export const imageTag = (params: Container.ImageTag) => {
    return http.post(`/containers/image/tag`, params);
};
export const imageRemove = (params: Container.BatchDelete) => {
    return http.post(`/containers/image/remove`, params);
};

// network
export const getNetworkPage = (params: ReqPage) => {
    return http.post<ResPage<Container.NetworkInfo>>(`/containers/network/search`, params);
};
export const deleteNetwork = (params: Container.BatchDelete) => {
    return http.post(`/containers/network/del`, params);
};
export const createNetwork = (params: Container.NetworkCreate) => {
    return http.post(`/containers/network`, params);
};

// volume
export const getVolumePage = (params: ReqPage) => {
    return http.post<ResPage<Container.VolumeInfo>>(`/containers/volume/search`, params);
};
export const volumeOptions = () => {
    return http.get<Array<Container.Options>>(`/containers/volume`);
};
export const deleteVolume = (params: Container.BatchDelete) => {
    return http.post(`/containers/volume/del`, params);
};
export const createVolume = (params: Container.VolumeCreate) => {
    return http.post(`/containers/volume`, params);
};

// repo
export const getRepoPage = (params: ReqPage) => {
    return http.post<ResPage<Container.RepoInfo>>(`/containers/repo/search`, params);
};
export const getRepoOption = () => {
    return http.get<Container.RepoOptions>(`/containers/repo`);
};
export const repoCreate = (params: Container.RepoCreate) => {
    return http.post(`/containers/repo`, params);
};
export const repoUpdate = (params: Container.RepoUpdate) => {
    return http.put(`/containers/repo/${params.id}`, params);
};
export const deleteRepo = (params: { ids: number[] }) => {
    return http.post(`/containers/repo/del`, params);
};
