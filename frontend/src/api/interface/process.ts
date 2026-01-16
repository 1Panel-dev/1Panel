export namespace Process {
    export interface StopReq {
        PID: number;
    }

    export interface PsProcessData {
        PID: number;
        name: string;
        PPID: number;
        username: string;
        status: string;
        startTime: string;
        numThreads: number;
        numConnections: number;
        cpuPercent: string;

        diskRead: string;
        diskWrite: string;
        cmdLine: string;

        rss: string;
        vms: string;
        hwm: string;
        data: string;
        stack: string;
        locked: string;
        swap: string;
        dirty: string;
        pss: string;
        uss: string;
        shared: string;
        text: string;

        cpuValue: number;
        rssValue: number;

        envs: string[];

        openFiles: OpenFilesStat[];
        connects: ProcessConnect[];
    }

    export interface ProcessConnect {
        type: string;
        status: string;
        localaddr: string;
        remoteaddr: string;
        PID: number;
        name: string;
    }

    export type ProcessConnects = ProcessConnect[];

    export interface OpenFilesStat {
        path: string;
        fd: number;
    }

    export interface ListeningProcess {
        PID: number;
        Port: { [key: string]: {} };
        Protocol: number;
        Name: string;
    }
}
