import http from '@/api';
import { ResPage } from '../interface';
import { User } from '../interface/user';

export const getUserList = (currentPage: number, pageSize: number) => {
    return http.get<ResPage<User.User>>(`/users?page=${currentPage}&pageSize=${pageSize}`);
};

export const addUser = (params: User.User) => {
    return http.post(`/users`, params);
};

export const getUserById = (id: number) => {
    return http.get<User.User>(`/users/${id}`);
};

export const editUser = (params: User.User) => {
    return http.post(`/users/` + params.id, params);
};

export const deleteUser = (params: { ids: number[] }) => {
    return http.post(`/users/del`, params);
};
