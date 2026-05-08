export namespace HostTool {
    export interface HostTool {
        type: string;
        config: {};
    }

    export interface Supervisor extends HostTool {
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

    export interface SupervisorConfigUpdate {
        content?: string;
    }

    export interface SupervisorConfigRes {
        type: string;
        content: string;
    }

    export interface SupervisorInit {
        type: string;
        configPath: string;
        serviceName: string;
    }

    export interface SupervisorProcess {
        operate: string;
        name: string;
        command: string;
        user: string;
        dir: string;
        numprocs: string;
        status?: ProcessStatus[];
        autoRestart: string;
        autoStart: string;
        environment: string;
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

    export interface ProcessFileGetReq {
        name: string;
        file: string;
    }

    export interface ProcessFileReq {
        operate: string;
        name: string;
        content?: string;
        file: string;
    }
}
