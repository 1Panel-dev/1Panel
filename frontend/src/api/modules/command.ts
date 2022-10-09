import http from '@/api';
import { ResPage } from '../interface';
import { Command } from '../interface/command';

export const getCommandList = () => {
    return http.get<Array<Command.CommandInfo>>(`/commands`, {});
};

export const getCommandPage = (params: Command.CommandSearch) => {
    return http.post<ResPage<Command.CommandInfo>>(`/commands/search`, params);
};

export const addCommand = (params: Command.CommandOperate) => {
    return http.post<Command.CommandOperate>(`/commands`, params);
};

export const editCommand = (params: Command.CommandOperate) => {
    return http.put(`/commands/${params.id}`, params);
};

export const deleteCommand = (params: { ids: number[] }) => {
    return http.post(`/commands/del`, params);
};
