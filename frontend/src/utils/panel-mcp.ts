import type { PanelMCP } from '@/api/interface/panel-mcp';
import { isAxiosError } from 'axios';

export const DEFAULT_MCP_DOCKER_IMAGE = '1panel/1panel-mcp-server:v0.0.1';
const releaseVersionPattern =
    'v?(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?';
export const isMCPDockerImage = (image: string) =>
    image.length <= 180 &&
    new RegExp(`^1panel/1panel-mcp-server(?::${releaseVersionPattern}|@sha256:[a-f0-9]{64})$`).test(image);

export const copyMCPSettings = (settings: PanelMCP.Settings): PanelMCP.Settings => ({
    runtimeType: settings.runtimeType,
    dockerImage: settings.dockerImage,
    enabled: settings.enabled,
    listenPort: settings.listenPort,
    externalUrl: settings.externalUrl,
    revision: settings.revision,
    accessPolicy: settings.accessPolicy,
    allowedIPs: [...settings.allowedIPs],
});

export const hasMCPSettingsChanges = (settings: PanelMCP.Settings, saved: PanelMCP.Settings) =>
    (['listenPort', 'externalUrl', 'accessPolicy'] as const).some((field) => settings[field] !== saved[field]) ||
    JSON.stringify(settings.allowedIPs) !== JSON.stringify(saved.allowedIPs);

export const getMCPCredentialStatus = (
    item: Pick<PanelMCP.Credential, 'status' | 'expiresAt'>,
    now = Date.now(),
): PanelMCP.CredentialStatus => {
    if (item.status === 'Revoked') return 'Revoked';
    if (item.expiresAt && (!Number.isFinite(Date.parse(item.expiresAt)) || Date.parse(item.expiresAt) <= now)) {
        return 'Expired';
    }
    return item.status;
};

export const buildMCPClientConfig = (
    endpoint: string,
    token = '<MCP_TOKEN>',
    client: PanelMCP.Client = 'generic',
    credentialId = '',
) => {
    if (!endpoint) return '';
    const name = credentialId ? `1panel_${credentialId.replaceAll('-', '')}` : '1panel';
    if (client === 'codex') {
        return `[mcp_servers.${JSON.stringify(name)}]\nurl = ${JSON.stringify(endpoint)}\nhttp_headers = { Authorization = ${JSON.stringify(`Bearer ${token}`)} }\n`;
    }
    return JSON.stringify(
        {
            mcpServers: {
                [name]: { type: 'http', url: endpoint, headers: { Authorization: `Bearer ${token}` } },
            },
        },
        null,
        2,
    );
};

export const parseMCPAllowedIPs = (value: string) => [
    ...new Set(
        value
            .split(/[\s,]+/)
            .map((item) => item.trim())
            .filter(Boolean),
    ),
];

export const isMCPAllowedIPList = (values: string[]) =>
    values.length <= 64 &&
    values.every((value) => {
        if (!value || value.length > 64 || value.includes('%')) return false;
        const [ip, prefix, extra] = value.split('/');
        if (extra !== undefined) return false;
        const v6 = ip.includes(':');
        if (prefix !== undefined && (!/^(0|[1-9]\d*)$/.test(prefix) || Number(prefix) > (v6 ? 128 : 32))) return false;
        if (!v6) {
            const parts = ip.split('.');
            return parts.length === 4 && parts.every((part) => /^(0|[1-9]\d{0,2})$/.test(part) && Number(part) <= 255);
        }
        if (!/^[0-9a-fA-F:.]+$/.test(ip)) return false;
        try {
            return new URL(`http://[${ip}]`).hostname.startsWith('[');
        } catch {
            return false;
        }
    });

export type MCPToolFilter = 'all' | 'read' | 'write';
export const filterMCPTools = (tools: PanelMCP.Tool[], search: string, mode: MCPToolFilter) => {
    const keyword = search.trim().toLocaleLowerCase();
    return tools.filter(
        (tool) =>
            (mode === 'all' || tool.readOnly === (mode === 'read')) &&
            [tool.toolId, tool.name, tool.description].some((value) => value.toLocaleLowerCase().includes(keyword)),
    );
};

export const groupMCPTools = (tools: PanelMCP.Tool[], filtered: PanelMCP.Tool[], selected: string[]) => {
    const visibleIds = new Set(filtered.map((tool) => tool.toolId));
    const selectedIds = new Set(selected);
    const groups = new Map<
        string,
        { id: string; name: string; order: number; total: number; selected: number; tools: PanelMCP.Tool[] }
    >();
    for (const tool of tools) {
        let group = groups.get(tool.categoryId);
        if (!group) {
            group = {
                id: tool.categoryId,
                name: tool.categoryName,
                order: tool.categoryOrder,
                total: 0,
                selected: 0,
                tools: [],
            };
            groups.set(tool.categoryId, group);
        }
        group.total++;
        if (selectedIds.has(tool.toolId)) group.selected++;
        if (visibleIds.has(tool.toolId)) group.tools.push(tool);
    }
    return [...groups.values()]
        .filter((group) => group.tools.length)
        .sort((a, b) => a.order - b.order || a.id.localeCompare(b.id));
};

export const changeVisibleMCPGrants = (
    selected: string[],
    visible: PanelMCP.Tool[],
    action: 'select' | 'clear' | 'readOnly',
) => {
    const result = new Set(selected);
    for (const tool of visible) {
        if (action === 'select' || (action === 'readOnly' && tool.readOnly)) result.add(tool.toolId);
        else result.delete(tool.toolId);
    }
    return [...result];
};

export const isMCPRevisionConflict = (error: unknown) => {
    const candidate = error as {
        code?: number;
        errorCode?: string;
        response?: { status?: number; data?: { code?: number; errorCode?: string } };
    };
    return (
        candidate?.code === 409 ||
        candidate?.errorCode === 'MCP_CONFLICT' ||
        candidate?.response?.data?.code === 409 ||
        candidate?.response?.data?.errorCode === 'MCP_CONFLICT' ||
        candidate?.response?.status === 409
    );
};

export const isMCPPolicyChanged = (error: unknown) => {
    const candidate = error as { errorCode?: string; response?: { data?: { errorCode?: string } } };
    return (
        candidate?.errorCode === 'MCP_POLICY_CHANGED' || candidate?.response?.data?.errorCode === 'MCP_POLICY_CHANGED'
    );
};

export const isMCPMutationUncertain = (error: unknown) => {
    const isServerError = (code: unknown) => typeof code === 'number' && code >= 500 && code < 600;
    if (isAxiosError(error)) return !error.response || isServerError(error.response.status);
    return isServerError((error as { code?: unknown })?.code);
};

const isObject = (value: unknown): value is Record<string, unknown> =>
    typeof value === 'object' && value !== null && !Array.isArray(value);

const isMCPNodeID = (id: unknown): id is string =>
    typeof id === 'string' &&
    (id === 'local' || (/^[1-9][0-9]{0,19}$/.test(id) && BigInt(id) <= 18446744073709551615n));

export const isUsableMCPToolNodes = (tools: PanelMCP.Tool[], mapping: unknown): mapping is Record<string, string[]> =>
    isObject(mapping) &&
    Object.keys(mapping).length === tools.length &&
    new Set(tools.map((tool) => tool.toolId)).size === tools.length &&
    tools.every((tool) => {
        const nodes = mapping[tool.toolId];
        return (
            Object.hasOwn(mapping, tool.toolId) &&
            Array.isArray(nodes) &&
            nodes.every(isMCPNodeID) &&
            new Set(nodes).size === nodes.length
        );
    });

export const filterMCPToolsByNodes = (
    tools: PanelMCP.Tool[],
    snapshot: Record<string, string[]>,
    nodeIds: string[],
    current = snapshot,
) =>
    tools.filter((tool) =>
        nodeIds.some((id) => snapshot[tool.toolId]?.includes(id) && current[tool.toolId]?.includes(id)),
    );

const isToolOperation = (operation: PanelMCP.ToolOperation) =>
    isObject(operation) &&
    typeof operation.id === 'string' &&
    !!operation.id &&
    typeof operation.method === 'string' &&
    !!operation.method &&
    typeof operation.path === 'string' &&
    operation.path.startsWith('/') &&
    !/[?#\s]/.test(operation.path) &&
    isObject(operation.bodySchema) &&
    Array.isArray(operation.responseFields) &&
    operation.responseFields.every((field) => typeof field === 'string');

export const isUsableMCPCatalog = (catalog: PanelMCP.Catalog) =>
    catalog?.contractVersion === 5 &&
    catalog.scope === 'nodes' &&
    Array.isArray(catalog.tools) &&
    catalog.tools.every(
        (tool) =>
            isObject(tool) &&
            typeof tool.toolId === 'string' &&
            !!tool.toolId &&
            typeof tool.name === 'string' &&
            typeof tool.description === 'string' &&
            typeof tool.categoryId === 'string' &&
            !!tool.categoryId.trim() &&
            typeof tool.categoryName === 'string' &&
            !!tool.categoryName.trim() &&
            Number.isInteger(tool.categoryOrder) &&
            tool.categoryOrder >= 0 &&
            typeof tool.policyHash === 'string' &&
            !!tool.policyHash &&
            Array.isArray(tool.operations) &&
            tool.operations.length > 0 &&
            tool.operations.every(isToolOperation) &&
            new Set(tool.operations.map((operation) => operation.id)).size === tool.operations.length &&
            typeof tool.readOnly === 'boolean',
    ) &&
    new Set(catalog.tools.map((tool) => tool.toolId)).size === catalog.tools.length &&
    isUsableMCPToolNodes(catalog.tools, catalog.toolNodeIds);

export const getMCPStaleTools = (credential: PanelMCP.Credential, tools: PanelMCP.Tool[]) =>
    credential.tools.filter((id) => {
        const current = tools.find((tool) => tool.toolId === id);
        if (
            !current?.policyHash ||
            !Array.isArray(credential.staleTools) ||
            credential.staleTools.some((tool) => typeof tool !== 'string') ||
            !isObject(credential.toolReadOnly) ||
            !Object.hasOwn(credential.toolReadOnly, id)
        )
            return true;
        const reviewedReadOnly = credential.toolReadOnly[id];
        // Core returns the effective contract for compatible grants. Bind that
        // decision to this catalog so responses from different upgrades cannot mix.
        return (
            credential.staleTools.includes(id) ||
            credential.toolPolicies?.[id] !== current.policyHash ||
            typeof reviewedReadOnly !== 'boolean' ||
            reviewedReadOnly !== current.readOnly
        );
    });

export const isUsableMCPNodes = (nodes: unknown): nodes is PanelMCP.Node[] =>
    Array.isArray(nodes) &&
    nodes.every(
        (node) =>
            node &&
            isMCPNodeID(node.id) &&
            typeof node.name === 'string' &&
            node.name.length > 0 &&
            typeof node.status === 'string' &&
            node.isLocal === (node.id === 'local'),
    ) &&
    new Set(nodes.map((node) => node.id)).size === nodes.length;
