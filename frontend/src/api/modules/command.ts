import http from '@/api';
import { ResPage, SearchWithPage } from '../interface';
import { Command } from '../interface/command';

export const getCommandList = (type: string) => {
    return http.post<Array<Command.CommandInfo>>(`/core/commands/list`, { type: type });
};
export const getCommandPage = (params: SearchWithPage) => {
    return http.post<ResPage<Command.CommandInfo>>(`/core/commands/search`, params);
};
export const getCommandTree = (type: string) => {
    return http.post<any>(`/core/commands/tree`, { type: type });
};
export const addCommand = (params: Command.CommandOperate) => {
    return http.post<Command.CommandOperate>(`/core/commands`, params);
};
export const editCommand = (params: Command.CommandOperate) => {
    return http.post(`/core/commands/update`, params);
};
export const deleteCommand = (params: { ids: number[] }) => {
    return http.post(`/core/commands/del`, params);
};
