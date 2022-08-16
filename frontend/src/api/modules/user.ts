import http from '@/api';
import { ResPage } from '../interface';
import { User } from '../interface/user';

export const getUserList = (params: User.ReqGetUserParams) => {
    return http.post<ResPage<User.User>>(`/users/search`, params);
};

export const addUser = (params: User.User) => {
    return http.post(`/users`, params);
};

export const getUserById = (id: number) => {
    return http.get<User.User>(`/users/${id}`);
};

export const editUser = (params: User.User) => {
    return http.put(`/users/` + params.id, params);
};

export const deleteUser = (params: { ids: number[] }) => {
    return http.post(`/users/del`, params);
};
