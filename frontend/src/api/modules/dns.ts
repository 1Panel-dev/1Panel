import http from '@/api';
import { ResPage } from '../interface';
import { Dns } from '../interface/dns';

// Zones
export const searchDnsZones = (req: Dns.ZoneSearch) => {
    return http.post<ResPage<Dns.ZoneRes>>('/dns/zones/search', req);
};

export const createDnsZone = (req: Dns.ZoneCreate) => {
    return http.post<any>('/dns/zones', req);
};

export const updateDnsZone = (req: Dns.ZoneUpdate) => {
    return http.post<any>('/dns/zones/update', req);
};

export const getDnsZone = (id: number) => {
    return http.get<Dns.ZoneRes>(`/dns/zones/${id}`);
};

export const deleteDnsZone = (params: { id: number }) => {
    return http.post<any>('/dns/zones/del', params);
};

export const syncDnsZone = (params: { id: number }) => {
    return http.post<any>('/dns/zones/sync', params);
};

// Records
export const searchDnsRecords = (req: Dns.RecordSearch) => {
    return http.post<ResPage<Dns.RecordRes>>('/dns/records/search', req);
};

export const createDnsRecord = (req: Dns.RecordCreate) => {
    return http.post<any>('/dns/records', req);
};

export const updateDnsRecord = (req: Dns.RecordUpdate) => {
    return http.post<any>('/dns/records/update', req);
};

export const deleteDnsRecord = (params: { id: number }) => {
    return http.post<any>('/dns/records/del', params);
};

// Setup & Status
export const setupDns = (params: { enable: boolean }) => {
    return http.post<any>('/dns/setup', params);
};

export const getDnsStatus = () => {
    return http.get<Dns.Status>('/dns/status');
};
