import { CommonModel, ReqPage } from '.';

export namespace Dns {
    export interface Zone extends CommonModel {
        name: string;
        type: string;
        masterIP: string;
        soaPrimary: string;
        soaEmail: string;
        soaRefresh: number;
        soaRetry: number;
        soaExpire: number;
        soaTTL: number;
        defaultTTL: number;
        status: string;
        pdnsID: string;
    }

    export interface ZoneRes extends Zone {
        recordCount: number;
    }

    export interface ZoneSearch extends ReqPage {
        info: string;
        orderBy: string;
        order: string;
    }

    export interface ZoneCreate {
        name: string;
        type: string;
        masterIP?: string;
        soaPrimary?: string;
        soaEmail: string;
        defaultTTL?: number;
    }

    export interface ZoneUpdate {
        id: number;
        soaPrimary?: string;
        soaEmail?: string;
        soaRefresh?: number;
        soaRetry?: number;
        soaExpire?: number;
        soaTTL?: number;
        defaultTTL?: number;
    }

    export interface Record extends CommonModel {
        zoneID: number;
        name: string;
        type: string;
        content: string;
        ttl: number;
        priority: number;
        disabled: boolean;
        autoGen: boolean;
    }

    export interface RecordRes extends Record {
        zoneName: string;
    }

    export interface RecordSearch extends ReqPage {
        zoneID: number;
        type?: string;
        info?: string;
    }

    export interface RecordCreate {
        zoneID: number;
        name: string;
        type: string;
        content: string;
        ttl?: number;
        priority?: number;
    }

    export interface RecordUpdate {
        id: number;
        name: string;
        type: string;
        content: string;
        ttl?: number;
        priority?: number;
        disabled?: boolean;
    }

    export interface Status {
        enabled: boolean;
        containerUp: boolean;
        containerName: string;
        version: string;
    }
}
