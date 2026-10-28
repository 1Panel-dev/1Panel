export namespace PanelMCP {
    export type CredentialStatus = 'Enable' | 'Disable' | 'Revoked' | 'Expired';
    export interface EditableCredential {
        name: string;
        description: string;
        expiresAt: string | null;
        tools: string[];
        toolPolicies: Record<string, string>;
        toolReadOnly: Record<string, boolean>;
        nodeIds: string[];
    }
    export interface Credential extends EditableCredential {
        id: string;
        keyHint: string;
        status: CredentialStatus;
        revision: number;
        createdAt: string;
        ownerName: string;
        staleTools: string[];
    }
    export interface Reference {
        id: string;
        revision: number;
    }
    export interface Created {
        item: Credential;
        token: string;
        alreadyCreated?: boolean;
    }
    export type Client = 'generic' | 'codex' | 'claude';
    export interface Revealed {
        id: string;
        token: string;
    }
    export interface Exported {
        filename: string;
        content: string;
    }
    export interface ToolOperation {
        id: string;
        method: string;
        path: string;
        bodySchema: Record<string, unknown>;
        responseFields: string[];
    }
    export interface Tool {
        toolId: string;
        schemaVersion: number;
        name: string;
        description: string;
        categoryId: string;
        categoryName: string;
        categoryOrder: number;
        readOnly: boolean;
        risk: string;
        inputSchema: Record<string, unknown>;
        operations: ToolOperation[];
        policyHash: string;
    }
    export interface Catalog {
        contractVersion: number;
        tools: Tool[];
        scope: 'nodes';
        toolNodeIds: Record<string, string[]>;
    }
    export interface Node {
        id: string;
        name: string;
        status: string;
        isLocal: boolean;
    }
    export type RuntimeType = 'docker';
    export interface Settings {
        runtimeType: RuntimeType;
        dockerImage: string;
        enabled: boolean;
        listenPort: number;
        externalUrl: string;
        revision: number;
        accessPolicy: 'inherit' | 'restricted' | 'all';
        allowedIPs: string[];
    }
    export interface ServiceStatus {
        runtimeType: RuntimeType;
        dockerImage: string;
        containerName: string;
        imageId: string;
        enabled: boolean;
        installed: boolean;
        status: 'not_installed' | 'unsupported' | 'stopped' | 'running' | 'unhealthy' | 'incompatible';
        ready: boolean;
        reason: string;
        adapterVersion: string;
        contractVersion: number;
        protocol: 'streamable-http';
        endpoint: string;
        internalEndpoint: string;
        canManageService: boolean;
        installing: boolean;
        installTaskID: string;
        website?: WebsiteBinding;
    }
    export interface WebsiteDraft {
        requested: boolean;
        domain: string;
        enableSSL: boolean;
        sslID: number;
    }
    export interface WebsitePendingRequest {
        operation: 'bind' | 'update' | 'unbind';
        requestId: string;
        domain?: string;
        enableSSL?: boolean;
        sslID?: number;
    }
    export interface WebsiteBinding {
        state: 'unbound' | 'pending' | 'bound' | 'unavailable';
        websiteId: number;
        domain: string;
        endpoint: string;
        sslID: number;
        enableSSL: boolean;
        websiteStatus: string;
        requestId: string;
        revision: number;
        reason?: string;
        pendingRequest?: WebsitePendingRequest;
    }
    export interface WebsiteMutation {
        requestId: string;
        revision: number;
    }
    export interface WebsiteBind extends WebsiteMutation {
        domain: string;
        enableSSL: boolean;
        sslID: number;
    }
    export interface WebsiteUpdate extends WebsiteMutation {
        enableSSL: boolean;
        sslID: number;
    }
    export interface Installation {
        dockerImage: string;
        listenPort: number;
        externalUrl: string;
        revision: number;
        accessPolicy: Settings['accessPolicy'];
        allowedIPs: string[];
        website?: Pick<WebsiteDraft, 'domain' | 'enableSSL' | 'sslID'>;
    }
    export interface InstallationTask {
        taskID: string;
    }
    export type Operation = 'start' | 'stop' | 'restart';
    export interface Invocation {
        id: string;
        requestId: string;
        credentialId: string;
        credentialName: string;
        ownerName: string;
        toolId: string;
        nodeId: string;
        revision: number;
        status: 'running' | 'success' | 'denied' | 'failed';
        code: string;
        startedAt: string;
        finishedAt: string | null;
        durationMs: number;
    }
}
