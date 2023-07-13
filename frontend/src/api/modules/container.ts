import http from '@/api';
import { ResPage, SearchWithPage } from '../interface';
import { Container } from '../interface/container';

export const searchContainer = (params: Container.ContainerSearch) => {
    return http.post<ResPage<Container.ContainerInfo>>(`/containers/search`, params, 400000);
};
export const listContainer = () => {
    return http.post<Array<string>>(`/containers/list`, {});
};
export const loadResourceLimit = () => {
    return http.get<Container.ResourceLimit>(`/containers/limit`);
};
export const createContainer = (params: Container.ContainerHelper) => {
    return http.post(`/containers`, params, 3000000);
};
export const updateContainer = (params: Container.ContainerHelper) => {
    return http.post(`/containers/update`, params, 3000000);
};
export const upgradeContainer = (name: string, image: string, forcePull: boolean) => {
    return http.post(`/containers/upgrade`, { name: name, image: image, forcePull: forcePull }, 3000000);
};
export const loadContainerInfo = (name: string) => {
    return http.post<Container.ContainerHelper>(`/containers/info`, { name: name });
};
export const cleanContainerLog = (containerName: string) => {
    return http.post(`/containers/clean/log`, { name: containerName });
};
export const containerListStats = () => {
    return http.get<Array<Container.ContainerListStats>>(`/containers/list/stats`);
};
export const containerStats = (id: string) => {
    return http.get<Container.ContainerStats>(`/containers/stats/${id}`);
};
export const containerOperator = (params: Container.ContainerOperate) => {
    return http.post(`/containers/operate`, params);
};
export const containerPrune = (params: Container.ContainerPrune) => {
    return http.post<Container.ContainerPruneReport>(`/containers/prune`, params);
};
export const inspect = (params: Container.ContainerInspect) => {
    return http.post<string>(`/containers/inspect`, params);
};

// image
export const searchImage = (params: SearchWithPage) => {
    return http.post<ResPage<Container.ImageInfo>>(`/containers/image/search`, params);
};
export const listImage = () => {
    return http.get<Array<Container.Options>>(`/containers/image`);
};
export const imageBuild = (params: Container.ImageBuild) => {
    return http.post<string>(`/containers/image/build`, params);
};
export const imagePull = (params: Container.ImagePull) => {
    return http.post<string>(`/containers/image/pull`, params);
};
export const imagePush = (params: Container.ImagePush) => {
    return http.post<string>(`/containers/image/push`, params);
};
export const imageLoad = (params: Container.ImageLoad) => {
    return http.post(`/containers/image/load`, params, 1200000);
};
export const imageSave = (params: Container.ImageSave) => {
    return http.post(`/containers/image/save`, params, 1200000);
};
export const imageTag = (params: Container.ImageTag) => {
    return http.post(`/containers/image/tag`, params);
};
export const imageRemove = (params: Container.BatchDelete) => {
    return http.post(`/containers/image/remove`, params);
};

// network
export const searchNetwork = (params: SearchWithPage) => {
    return http.post<ResPage<Container.NetworkInfo>>(`/containers/network/search`, params);
};
export const listNetwork = () => {
    return http.get<Array<Container.Options>>(`/containers/network`);
};
export const deleteNetwork = (params: Container.BatchDelete) => {
    return http.post(`/containers/network/del`, params);
};
export const createNetwork = (params: Container.NetworkCreate) => {
    return http.post(`/containers/network`, params);
};

// volume
export const searchVolume = (params: SearchWithPage) => {
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
export const checkRepoStatus = (id: number) => {
    return http.post(`/containers/repo/status`, { id: id }, 40000);
};
export const searchImageRepo = (params: SearchWithPage) => {
    return http.post<ResPage<Container.RepoInfo>>(`/containers/repo/search`, params);
};
export const listImageRepo = () => {
    return http.get<Container.RepoOptions>(`/containers/repo`);
};
export const createImageRepo = (params: Container.RepoCreate) => {
    return http.post(`/containers/repo`, params, 40000);
};
export const updateImageRepo = (params: Container.RepoUpdate) => {
    return http.post(`/containers/repo/update`, params, 40000);
};
export const deleteImageRepo = (params: Container.RepoDelete) => {
    return http.post(`/containers/repo/del`, params, 40000);
};

// composeTemplate
export const searchComposeTemplate = (params: SearchWithPage) => {
    return http.post<ResPage<Container.TemplateInfo>>(`/containers/template/search`, params);
};
export const listComposeTemplate = () => {
    return http.get<Container.TemplateInfo>(`/containers/template`);
};
export const deleteComposeTemplate = (params: { ids: number[] }) => {
    return http.post(`/containers/template/del`, params);
};
export const createComposeTemplate = (params: Container.TemplateCreate) => {
    return http.post(`/containers/template`, params);
};
export const updateComposeTemplate = (params: Container.TemplateUpdate) => {
    return http.post(`/containers/template/update`, params);
};

// compose
export const searchCompose = (params: SearchWithPage) => {
    return http.post<ResPage<Container.ComposeInfo>>(`/containers/compose/search`, params);
};
export const upCompose = (params: Container.ComposeCreate) => {
    return http.post<string>(`/containers/compose`, params);
};
export const testCompose = (params: Container.ComposeCreate) => {
    return http.post<boolean>(`/containers/compose/test`, params);
};
export const composeOperator = (params: Container.ComposeOpration) => {
    return http.post(`/containers/compose/operate`, params);
};
export const composeUpdate = (params: Container.ComposeUpdate) => {
    return http.post(`/containers/compose/update`, params, 600000);
};

// docker
export const dockerOperate = (operation: string) => {
    return http.post(`/containers/docker/operate`, { operation: operation });
};
export const loadDaemonJson = () => {
    return http.get<Container.DaemonJsonConf>(`/containers/daemonjson`);
};
export const loadDaemonJsonFile = () => {
    return http.get<string>(`/containers/daemonjson/file`);
};
export const loadDockerStatus = () => {
    return http.get<string>(`/containers/docker/status`);
};
export const updateDaemonJson = (key: string, value: string) => {
    return http.post(`/containers/daemonjson/update`, { key: key, value: value }, 60000);
};
export const updateLogOption = (maxSize: string, maxFile: string) => {
    return http.post(`/containers/logoption/update`, { logMaxSize: maxSize, logMaxFile: maxFile }, 60000);
};
export const updateDaemonJsonByfile = (params: Container.DaemonJsonUpdateByFile) => {
    return http.post(`/containers/daemonjson/update/byfile`, params);
};
