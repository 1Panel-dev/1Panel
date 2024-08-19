import { Group } from '../interface/group';
import http from '@/api';

export const GetGroupList = (type: string) => {
    return http.post<Array<Group.GroupInfo>>(`/core/groups/search`, { type: type });
};
export const CreateGroup = (params: Group.GroupCreate) => {
    return http.post<Group.GroupCreate>(`/core/groups`, params);
};
export const UpdateGroup = (params: Group.GroupUpdate) => {
    return http.post(`/core/groups/update`, params);
};
export const DeleteGroup = (id: number) => {
    return http.post(`/core/groups/del`, { id: id });
};
