import { CommonModel, ReqPage } from '.';

export namespace Mail {
    export interface Domain extends CommonModel {
        name: string;
        status: string;
        dkimKey: string;
        dnsAutoGen: boolean;
    }

    export interface DomainRes extends Domain {
        accountCount: number;
        aliasCount: number;
    }

    export interface DomainCreate {
        name: string;
        dnsAutoGen: boolean;
    }

    export interface DomainDelete {
        id: number;
        cleanupDns: boolean;
    }

    export interface Account extends CommonModel {
        domainID: number;
        username: string;
        email: string;
        quota: number;
        status: string;
    }

    export interface AccountRes extends Account {
        domainName: string;
    }

    export interface AccountSearch extends ReqPage {
        domainID?: number;
        info?: string;
    }

    export interface AccountCreate {
        domainID: number;
        username: string;
        password: string;
        quota?: number;
    }

    export interface AccountUpdate {
        id: number;
        password?: string;
        quota?: number;
    }

    export interface Alias extends CommonModel {
        domainID: number;
        source: string;
        target: string;
    }

    export interface AliasRes extends Alias {
        domainName: string;
    }

    export interface AliasSearch extends ReqPage {
        domainID?: number;
        info?: string;
    }

    export interface AliasCreate {
        domainID: number;
        source: string;
        target: string;
    }

    export interface Status {
        enabled: boolean;
        mailContainerUp: boolean;
        mailContainerName: string;
        roundcubeUp: boolean;
        roundcubeURL: string;
        version: string;
    }
}
