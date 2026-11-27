export namespace HostNetwork {
    export interface Overview {
        persistence: Persistence;
        interfaces: Interface[];
        routes: Route[];
    }
    export interface Persistence {
        supported: boolean;
        os: string;
        reason: string;
    }
    export interface Address {
        editable: boolean;
        reason: string;
        address: string;
        family: string;
        scope: string;
        managed: boolean;
        applied: boolean;
    }
    export interface Interface {
        backend: string;
        reason: string;
        index: number;
        name: string;
        state: string;
        linkState: string;
        mac: string;
        mtu: number;
        flags: string[];
        addresses: Address[];
    }
    export interface AddressRequest {
        interface: string;
        address: string;
    }
    export interface AddressChange extends AddressRequest {
        previousAddress?: string;
        confirmation: string;
    }
    export interface RouteChange extends Route {
        previous?: Route;
        confirmation: string;
    }
    export interface NextHop {
        gateway: string;
        interface: string;
        weight: number;
    }
    export interface Route {
        editable: boolean;
        reason: string;
        destination: string;
        gateway: string;
        interface: string;
        metric: number;
        table: string;
        protocol: string;
        source?: string;
        scope: string;
        family: string;
        readOnly: boolean;
        nextHops: NextHop[];
        managed: boolean;
        applied: boolean;
    }
    export interface Reachability {
        skipped?: boolean;
        failed?: boolean;
        reachable: boolean;
        message: string;
    }
}
