export namespace Group {
    export interface GroupInfo {
        id: number;
        name: string;
        type: string;
        isDefault: boolean;
    }
    export interface GroupCreate {
        id: number;
        name: string;
        type: string;
    }
    export interface GroupUpdate {
        id: number;
        name: string;
        isDefault: boolean;
    }
    export interface GroupSearch {
        type: string;
    }
}
