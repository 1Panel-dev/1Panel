export namespace Command {
    export interface CommandInfo {
        id: number;
        name: string;
        groupID: number;
        command: string;
    }
    export interface CommandOperate {
        id: number;
        name: string;
        groupID: number;
        command: string;
    }
    export interface RedisCommand {
        id: number;
        name: string;
        command: string;
    }
}
