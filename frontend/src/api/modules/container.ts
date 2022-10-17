import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Container } from '../interface/container';

export const searchContainer = (params: ReqPage) => {
    return http.post<ResPage<Container.ContainerInfo>>(`/containers/search`, params);
};
export const createContainer = (params: Container.ContainerCreate) => {
    return http.post(`/containers`, params);
};
export const logContainer = (params: Container.ContainerLogSearch) => {
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
export const searchImage = (params: ReqPage) => {
    return http.post<ResPage<Container.ImageInfo>>(`/containers/image/search`, params);
};
export const listImage = () => {
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
export const searchNetwork = (params: ReqPage) => {
    return http.post<ResPage<Container.NetworkInfo>>(`/containers/network/search`, params);
};
export const deleteNetwork = (params: Container.BatchDelete) => {
    return http.post(`/containers/network/del`, params);
};
export const createNetwork = (params: Container.NetworkCreate) => {
    return http.post(`/containers/network`, params);
};

// volume
export const searchVolume = (params: ReqPage) => {
    return http.post<ResPage<Container.VolumeInfo>>(`/containers/volume/search`, params);
};
export const listVolume = () => {
    return http.get<Array<Container.Options>>(`/containers/volume`);
};
export const deleteVolume = (params: Container.BatchDelete) => {
    return http.post(`/containers/volume/del`, params);
};
export const createVolume = (params: Container.VolumeCreate) => {
    return http.post(`/containers/volume`, params);
};

// repo
export const searchImageRepo = (params: ReqPage) => {
    return http.post<ResPage<Container.RepoInfo>>(`/containers/repo/search`, params);
};
export const listImageRepo = () => {
    return http.get<Container.RepoOptions>(`/containers/repo`);
};
export const createImageRepo = (params: Container.RepoCreate) => {
    return http.post(`/containers/repo`, params);
};
export const updateImageRepo = (params: Container.RepoUpdate) => {
    return http.put(`/containers/repo/${params.id}`, params);
};
export const deleteImageRepo = (params: { ids: number[] }) => {
    return http.post(`/containers/repo/del`, params);
};

// composeTemplate
export const searchComposeTemplate = (params: ReqPage) => {
    return http.post<ResPage<Container.TemplateInfo>>(`/containers/compose/search`, params);
};
export const listComposeTemplate = (params: ReqPage) => {
    return http.post<ResPage<Container.TemplateInfo>>(`/containers/compose/search`, params);
};
export const deleteComposeTemplate = (params: { ids: number[] }) => {
    return http.post(`/containers/compose/del`, params);
};
export const createComposeTemplate = (params: Container.TemplateCreate) => {
    return http.post(`/containers/compose`, params);
};
export const updateComposeTemplate = (params: Container.TemplateUpdate) => {
    return http.put(`/containers/compose/${params.id}`, params);
};
