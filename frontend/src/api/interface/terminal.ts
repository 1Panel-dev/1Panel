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
    title: string;
    hostId: number;
    attached: boolean;
    createdAt: string;
    detachedAt: string;
}
