import http from '@/api';
import { User } from '../interface/user';
import UserDataList from '@/assets/json/user.json';

/**
 * @name 用户管理模块
 */
// * 获取用户列表
export const getUserList = (params: User.ReqGetUserParams) => {
    console.log(params);

    return UserDataList;
    // return http.post<ResPage<User.User>>(`/users/list`, params);
};

// * 新增用户
export const addUser = (params: User.UserCreate) => {
    return http.post(`/users/add`, params);
};

export const getUserById = (id: string) => {
    return http.get(`/users/detail/${id}`);
};
// // * 批量添加用户
// export const BatchAddUser = (params: FormData) => {
//     return http.post(`/users/import`, params);
// };

// * 编辑用户
export const editUser = (params: User.User) => {
    return http.post(`/users/edit`, params);
};

// * 批量删除用户
export const deleteUser = (params: { ids: number[] }) => {
    return http.post(`/users/delete`, params);
};

// * 切换用户状态
// export const changeUserStatus = (params: { id: string; status: number }) => {
//     return http.post(`/users/change`, params);

// };

// // * 重置用户密码
// export const resetUserPassWord = (params: { id: string }) => {
//     return http.post(`/users/rest_password`, params);
// };

// // * 导出用户数据
// export const exportUserInfo = (params: User.ReqGetUserParams) => {
//     return http.post<BlobPart>(`/user/export`, params, {
//         responseType: 'blob',
//     });
// };

// // * 获取用户状态
// export const getUserStatus = () => {
//     return http.get<User.ResStatus>(`/user/status`);
// };

// // * 获取用户性别字典
// export const getUserGender = () => {
//     return http.get<User.ResGender>(`/user/gender`);
// };
