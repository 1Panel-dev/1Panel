import http from '@/api';
import { ResPage } from '../interface';
import { Mail } from '../interface/mail';

// Domains
export const listMailDomains = () => {
    return http.get<Mail.DomainRes[]>('/mail/domains');
};

export const createMailDomain = (req: Mail.DomainCreate) => {
    return http.post<any>('/mail/domains', req);
};

export const deleteMailDomain = (req: Mail.DomainDelete) => {
    return http.post<any>('/mail/domains/del', req);
};

// Accounts
export const searchMailAccounts = (req: Mail.AccountSearch) => {
    return http.post<ResPage<Mail.AccountRes>>('/mail/accounts/search', req);
};

export const createMailAccount = (req: Mail.AccountCreate) => {
    return http.post<any>('/mail/accounts', req);
};

export const updateMailAccount = (req: Mail.AccountUpdate) => {
    return http.post<any>('/mail/accounts/update', req);
};

export const deleteMailAccount = (params: { id: number }) => {
    return http.post<any>('/mail/accounts/del', params);
};

// Aliases
export const searchMailAliases = (req: Mail.AliasSearch) => {
    return http.post<ResPage<Mail.AliasRes>>('/mail/aliases/search', req);
};

export const createMailAlias = (req: Mail.AliasCreate) => {
    return http.post<any>('/mail/aliases', req);
};

export const deleteMailAlias = (params: { id: number }) => {
    return http.post<any>('/mail/aliases/del', params);
};

// Setup & Status
export const setupMail = (params: { enable: boolean }) => {
    return http.post<any>('/mail/setup', params);
};

export const getMailStatus = () => {
    return http.get<Mail.Status>('/mail/status');
};
