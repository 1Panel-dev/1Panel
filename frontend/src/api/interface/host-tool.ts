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
    }

    export interface SupersivorProcess {
        operate: string;
        name: string;
        command: string;
        user: string;
        dir: string;
        numprocs: string;
    }
}
