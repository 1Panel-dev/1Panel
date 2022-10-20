export namespace Database {
    export interface MysqlDBInfo {
        id: number;
        createdAt: Date;
        name: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        permissionIPs: string;
        description: string;
    }
    export interface MysqlDBCreate {
        name: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        permissionIPs: string;
        description: string;
    }
}
