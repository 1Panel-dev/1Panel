export namespace APIKey {
    export type Status = 'Enable' | 'Disable' | 'Revoked' | 'Expired';
    export interface Editable {
        name: string;
        description: string;
        ipWhiteList: string;
        apiTrustedProxies: string;
        apiKeyValidityTime: number;
        expiresAt: string | null;
        allowAppBinding: boolean;
    }
    export interface Item extends Editable {
        id: string;
        kind: 'legacy' | 'apiKey';
        keyHint: string;
        status: Status;
        revision: number;
        createdAt: string | null;
    }
    export interface Search {
        items: Item[];
        total: number;
    }
    export interface Created {
        item: Item;
        apiKey: string;
        alreadyCreated?: boolean;
    }
    export interface Reference {
        id: string;
        revision: number;
    }
}
