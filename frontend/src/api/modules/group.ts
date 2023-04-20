import { Group } from '../interface/group';
import http from '@/api';

export const GetGroupList = (params: Group.GroupSearch) => {
    return http.post<Array<Group.GroupInfo>>(`/groups/search`, params);
};
export const CreateGroup = (params: Group.GroupCreate) => {
    return http.post<Group.GroupCreate>(`/groups`, params);
};
export const UpdateGroup = (params: Group.GroupUpdate) => {
    return http.post(`/groups/update`, params);
};
export const DeleteGroup = (id: number) => {
    return http.post(`/groups/del`, { id: id });
};
