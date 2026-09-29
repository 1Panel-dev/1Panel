import http from '@/api';
import type { APIKey } from '@/api/interface/api-key';

const prefix = '/core/auth/api/keys';

export const searchAPIKeys = (params: { page: number; pageSize: number; excludeRevoked?: boolean }) => {
    return http.post<APIKey.Search>(`${prefix}/search`, params);
};
export const createAPIKey = (params: APIKey.Editable & { requestID: string }) => {
    return http.post<APIKey.Created>(`${prefix}/create`, params);
};
export const updateAPIKey = (params: APIKey.Editable & APIKey.Reference) => {
    return http.post<{ terminalClosePending: boolean }>(`${prefix}/update`, params);
};
export const setAPIKeyStatus = (params: APIKey.Reference & { status: 'Enable' | 'Disable' }) => {
    return http.post<{ terminalClosePending: boolean }>(`${prefix}/status`, params);
};
export const revokeAPIKey = (params: APIKey.Reference) => {
    return http.post<{ terminalClosePending: boolean }>(`${prefix}/revoke`, params);
};
