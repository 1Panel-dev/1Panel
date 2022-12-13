import http from '@/api';
import { Group } from '../interface/group';

export const getGroupList = (params: Group.GroupSearch) => {
    return http.post<Array<Group.GroupInfo>>(`/groups/search`, params);
};

export const addGroup = (params: Group.GroupOperate) => {
    return http.post<Group.GroupOperate>(`/groups`, params);
};

export const editGroup = (params: Group.GroupOperate) => {
    return http.post(`/groups/update`, params);
};

export const deleteGroup = (id: number) => {
    return http.post(`/groups/del`, { id: id });
};
