import http from '@/api';
import type { ReqPage, ResPage } from '@/api/interface';
import type { PanelMCP } from '@/api/interface/panel-mcp';
import type { Container } from '@/api/interface/container';

const prefix = '/core/mcp';
const local = { headers: { CurrentNode: 'local' } };

export const getPanelMCPStatus = () => http.get<PanelMCP.ServiceStatus>(`${prefix}/status`, {}, local);
export const getPanelMCPDockerStatus = () =>
    http.get<Container.DockerStatus>('/containers/docker/status', {}, { ...local, timeout: 5000 });
export const getPanelMCPSettings = () => http.get<PanelMCP.Settings>(`${prefix}/settings`, {}, local);
export const getPanelMCPWebsite = () => http.get<PanelMCP.WebsiteBinding>(`${prefix}/website/status`, {}, local);
export const bindPanelMCPWebsite = (params: PanelMCP.WebsiteBind) =>
    http.postWithConfig<PanelMCP.WebsiteBinding>(`${prefix}/website/bind`, params, { ...local, timeout: 120000 });
export const updatePanelMCPWebsite = (params: PanelMCP.WebsiteUpdate) =>
    http.postWithConfig<PanelMCP.WebsiteBinding>(`${prefix}/website/update`, params, { ...local, timeout: 120000 });
export const unbindPanelMCPWebsite = (params: PanelMCP.WebsiteMutation) =>
    http.postWithConfig<PanelMCP.WebsiteBinding>(`${prefix}/website/unbind`, params, { ...local, timeout: 120000 });
export const savePanelMCPSettings = (params: PanelMCP.Settings) =>
    http.postWithConfig(`${prefix}/settings`, params, { ...local, timeout: 60000 });
export const operatePanelMCPService = (operation: PanelMCP.Operation) =>
    http.postWithConfig(`${prefix}/service/operate`, { operation }, { ...local, timeout: 60000 });
export const installPanelMCPService = (params: PanelMCP.Installation) =>
    http.postWithConfig<PanelMCP.InstallationTask>(`${prefix}/service/install`, params, local);
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
export const deletePanelMCPCredential = (params: PanelMCP.Reference) =>
    http.postWithConfig(`${prefix}/credentials/delete`, params, local);
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
