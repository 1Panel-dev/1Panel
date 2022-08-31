import http from '@/api';
import { Group } from '../interface/group';

export const getGroupList = (params: Group.GroupSearch) => {
    return http.post<Array<Group.GroupInfo>>(`/groups/search`, params);
};

export const addGroup = (params: Group.GroupOperate) => {
    return http.post<Group.GroupOperate>(`/groups`, params);
};

export const editGroup = (params: Group.GroupOperate) => {
    return http.put(`/groups/` + params.id, params);
};

export const deleteGroup = (id: number) => {
    return http.delete(`/groups/` + id);
};
