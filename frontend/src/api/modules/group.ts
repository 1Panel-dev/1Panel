import { Group } from '../interface/group';
import http from '@/api';

export const getGroupList = (type: string) => {
    return http.post<Array<Group.GroupInfo>>(`/core/groups/search`, { type: type });
};
export const createGroup = (params: Group.GroupCreate) => {
    return http.post<Group.GroupCreate>(`/core/groups`, params);
};
export const updateGroup = (params: Group.GroupUpdate) => {
    return http.post(`/core/groups/update`, params);
};
export const deleteGroup = (id: number) => {
    return http.post(`/core/groups/del`, { id: id });
};

export const getAgentGroupList = (type: string) => {
    return http.post<Array<Group.GroupInfo>>(`/groups/search`, { type: type });
};
export const createAgentGroup = (params: Group.GroupCreate) => {
    return http.post<Group.GroupCreate>(`/groups`, params);
};
export const updateAgentGroup = (params: Group.GroupUpdate) => {
    return http.post(`/groups/update`, params);
};
export const deleteAgentGroup = (id: number) => {
    return http.post(`/groups/del`, { id: id });
};
