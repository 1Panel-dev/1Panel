export interface ReqTerminal {
    name: string;
    ip: string;
    port: number;
    user: string;
    authType: string;
    password: string;
    key: string;
}

export interface TerminalSession {
    id: string;
    kind: 'local' | 'ssh';
    hostId: number;
    title: string;
    pinned: boolean;
    attached: boolean;
    createdAt: string;
    lastActiveAt: string;
    detachedAt?: string;
    expiresAt?: string;
}
