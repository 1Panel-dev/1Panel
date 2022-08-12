import http from '@/api';
import { ResPage } from '../interface';
import { User } from '../interface/user';
// import UserDataList from '@/assets/json/user.json';

/**
 * @name 用户管理模块
 */
// * 获取用户列表
export const getUserList = (params: User.ReqGetUserParams) => {
    // return UserDataList;
    return http.post<ResPage<User.User>>(`/users/search`, params);
};

// * 新增用户
export const addUser = (params: User.User) => {
    return http.post(`/users`, params);
};

export const getUserById = (id: number) => {
    return http.get<User.User>(`/users/${id}`);
};
// * 编辑用户
export const editUser = (params: User.User) => {
    return http.post(`/users/` + params.id, params);
};

// * 批量删除用户
export const deleteUser = (params: { ids: number[] }) => {
    return http.post(`/users/delete`, params);
};
