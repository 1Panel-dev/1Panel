import http from '@/api';
import { ResPage } from '../interface';
import { Command } from '../interface/command';
import { Group } from '../interface/group';
import { Host } from '../interface/host';

export const searchHosts = (params: Host.SearchWithPage) => {
    return http.post<ResPage<Host.Host>>(`/hosts/search`, params);
};
export const getHostTree = (params: Host.ReqSearch) => {
    return http.post<Array<Host.HostTree>>(`/hosts/tree`, params);
};
export const getHostInfo = (id: number) => {
    return http.get<Host.Host>(`/hosts/` + id);
};
export const addHost = (params: Host.HostOperate) => {
    return http.post<Host.HostOperate>(`/hosts`, params);
};
export const testByInfo = (params: Host.HostConnTest) => {
    return http.post<boolean>(`/hosts/test/byinfo`, params);
};
export const testByID = (id: number) => {
    return http.post<boolean>(`/hosts/test/byid/${id}`);
};
export const editHost = (params: Host.HostOperate) => {
    return http.post(`/hosts/update`, params);
};
export const editHostGroup = (params: Host.GroupChange) => {
    return http.post(`/hosts/update/group`, params);
};
export const deleteHost = (params: { ids: number[] }) => {
    return http.post(`/hosts/del`, params);
};

// group
export const getGroupList = (params: Group.GroupSearch) => {
    return http.post<Array<Group.GroupInfo>>(`/hosts/group/search`, params);
};
export const addGroup = (params: Group.GroupOperate) => {
    return http.post<Group.GroupOperate>(`/hosts/group`, params);
};
export const editGroup = (params: Group.GroupOperate) => {
    return http.post(`/hosts/group/update`, params);
};
export const deleteGroup = (id: number) => {
    return http.post(`/hosts/group/del`, { id: id });
};

// command
export const getCommandList = () => {
    return http.get<Array<Command.CommandInfo>>(`/hosts/command`, {});
};
export const getCommandPage = (params: Command.CommandSearch) => {
    return http.post<ResPage<Command.CommandInfo>>(`/hosts/command/search`, params);
};
export const addCommand = (params: Command.CommandOperate) => {
    return http.post<Command.CommandOperate>(`/hosts/command`, params);
};
export const editCommand = (params: Command.CommandOperate) => {
    return http.post(`/hosts/command/update`, params);
};
export const deleteCommand = (params: { ids: number[] }) => {
    return http.post(`/hosts/command/del`, params);
};
