import http from '@/api';
import { HostNetwork } from '@/api/interface/host-network';

export const getHostNetworkOverview = (node: string, signal?: AbortSignal) =>
    http.get<HostNetwork.Overview>('/hosts/network', {}, { headers: { CurrentNode: node }, signal });
export const addHostNetworkAddress = (params: HostNetwork.AddressChange, node: string) =>
    http.post('/hosts/network/address', params, undefined, { CurrentNode: node });
export const deleteHostNetworkAddress = (params: HostNetwork.AddressChange, node: string) =>
    http.post('/hosts/network/address/del', params, undefined, { CurrentNode: node });
export const addHostNetworkRoute = (params: HostNetwork.RouteChange, node: string) =>
    http.post('/hosts/network/routes', params, undefined, { CurrentNode: node });
export const deleteHostNetworkRoute = (params: HostNetwork.RouteChange, node: string) =>
    http.post('/hosts/network/routes/del', params, undefined, { CurrentNode: node });
export const checkHostNetworkRoute = (params: HostNetwork.Route, node: string, signal?: AbortSignal) =>
    http.postWithConfig<HostNetwork.Reachability>('/hosts/network/routes/check', params, {
        headers: { CurrentNode: node },
        signal,
    });
export const checkHostNetworkRoutes = (routes: HostNetwork.Route[], node: string, signal?: AbortSignal) =>
    http.postWithConfig<HostNetwork.Reachability[]>(
        '/hosts/network/routes/check/batch',
        { routes },
        {
            headers: { CurrentNode: node },
            signal,
            timeout: 120000,
        },
    );
