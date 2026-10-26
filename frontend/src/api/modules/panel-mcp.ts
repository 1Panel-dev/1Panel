import http from '@/api';
import type { ReqPage, ResPage } from '@/api/interface';
import type { PanelMCP } from '@/api/interface/panel-mcp';

const prefix = '/core/mcp';
const local = { headers: { CurrentNode: 'local' } };

export const getPanelMCPStatus = () => http.get<PanelMCP.ServiceStatus>(`${prefix}/status`, {}, local);
export const getPanelMCPSettings = () => http.get<PanelMCP.Settings>(`${prefix}/settings`, {}, local);
export const savePanelMCPSettings = (params: PanelMCP.Settings) =>
    http.postWithConfig(`${prefix}/settings`, params, { ...local, timeout: 60000 });
export const operatePanelMCPService = (operation: PanelMCP.Operation) =>
    http.postWithConfig(`${prefix}/service/operate`, { operation }, { ...local, timeout: 60000 });
export const installPanelMCPService = (params: PanelMCP.Installation) =>
    http.postWithConfig<PanelMCP.ServiceStatus>(`${prefix}/service/install`, params, { ...local, timeout: 420000 });
export const getPanelMCPLogs = () => http.get<{ content: string }>(`${prefix}/service/logs`, {}, local);
export const getPanelMCPCatalog = () => http.get<PanelMCP.Catalog>(`${prefix}/catalog`, {}, local);
export const getPanelMCPNodes = () => http.get<PanelMCP.Node[]>(`${prefix}/nodes`, {}, local);
export const searchPanelMCPCredentials = (params: ReqPage & { excludeRevoked?: boolean }) =>
    http.postWithConfig<ResPage<PanelMCP.Credential>>(`${prefix}/credentials/search`, params, local);
export const createPanelMCPCredential = (params: PanelMCP.EditableCredential & { requestID: string }) =>
    http.postWithConfig<PanelMCP.Created>(`${prefix}/credentials/create`, params, local);
export const updatePanelMCPCredential = (params: PanelMCP.EditableCredential & PanelMCP.Reference) =>
    http.postWithConfig(`${prefix}/credentials/update`, params, local);
export const setPanelMCPCredentialStatus = (params: PanelMCP.Reference & { status: 'Enable' | 'Disable' }) =>
    http.postWithConfig(`${prefix}/credentials/status`, params, local);
export const revokePanelMCPCredential = (params: PanelMCP.Reference) =>
    http.postWithConfig(`${prefix}/credentials/revoke`, params, local);
export const revealPanelMCPCredential = (params: PanelMCP.Reference, signal?: AbortSignal) =>
    http.postWithConfig<PanelMCP.Revealed>(`${prefix}/credentials/reveal`, params, { ...local, signal });
export const rotatePanelMCPCredential = (params: PanelMCP.Reference, signal?: AbortSignal) =>
    http.postWithConfig<PanelMCP.Created>(`${prefix}/credentials/rotate`, params, { ...local, signal });
export const exportPanelMCPCredential = (
    params: PanelMCP.Reference & { client: PanelMCP.Client },
    signal?: AbortSignal,
) => http.postWithConfig<PanelMCP.Exported>(`${prefix}/credentials/export`, params, { ...local, signal });
export const searchPanelMCPInvocations = (
    params: ReqPage & { credentialId?: string; toolId?: string; status?: string },
) => http.postWithConfig<ResPage<PanelMCP.Invocation>>(`${prefix}/invocations/search`, params, local);
