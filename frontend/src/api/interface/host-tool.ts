export namespace HostTool {
    export interface HostTool {
        type: string;
        config: {};
    }

    export interface Supersivor extends HostTool {
        configPath: string;
        includeDir: string;
        logPath: string;
        isExist: boolean;
        init: boolean;
        msg: string;
        version: string;
        status: string;
        ctlExist: boolean;
        serviceName: string;
    }

    export interface SupersivorConfig {
        type: string;
        operate: string;
        content?: string;
    }

    export interface SupersivorConfigRes {
        type: string;
        content: string;
    }

    export interface SupersivorInit {
        type: string;
        configPath: string;
        serviceName: string;
    }

    export interface SupersivorProcess {
        operate: string;
        name: string;
        command: string;
        user: string;
        dir: string;
        numprocs: string;
        status?: ProcessStatus[];
    }

    export interface ProcessStatus {
        PID: string;
        status: string;
        uptime: string;
        name: string;
    }

    export interface ProcessReq {
        operate: string;
        name: string;
    }

    export interface ProcessFileReq {
        operate: string;
        name: string;
        content?: string;
        file: string;
    }
}
