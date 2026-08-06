export interface ExternalLoginTickets {
    oidcTicket: string;
    saml2Ticket: string;
}

const oidcTicketParam = 'oidc_ticket';
const saml2TicketParam = 'saml2_ticket';

const getFragmentParams = (url: URL) => {
    return new URLSearchParams(url.hash.startsWith('#') ? url.hash.slice(1) : url.hash);
};

export const parseExternalLoginTickets = (href: string): ExternalLoginTickets => {
    const url = new URL(href);
    const fragmentParams = getFragmentParams(url);
    return {
        oidcTicket: fragmentParams.get(oidcTicketParam) || url.searchParams.get(oidcTicketParam) || '',
        saml2Ticket: fragmentParams.get(saml2TicketParam) || url.searchParams.get(saml2TicketParam) || '',
    };
};

export const hasExternalLoginTicket = () => {
    const { oidcTicket, saml2Ticket } = parseExternalLoginTickets(window.location.href);
    return Boolean(oidcTicket || saml2Ticket);
};

export const takeExternalTicketsFromURL = (): ExternalLoginTickets => {
    const tickets = parseExternalLoginTickets(window.location.href);
    if (!tickets.oidcTicket && !tickets.saml2Ticket) return tickets;

    const url = new URL(window.location.href);
    const fragmentParams = getFragmentParams(url);
    fragmentParams.delete(oidcTicketParam);
    fragmentParams.delete(saml2TicketParam);
    url.searchParams.delete(oidcTicketParam);
    url.searchParams.delete(saml2TicketParam);
    const sanitizedFragment = fragmentParams.toString();
    url.hash = sanitizedFragment ? `#${sanitizedFragment}` : '';
    const sanitizedURL = `${url.pathname}${url.search}${url.hash}`;
    window.history.replaceState(window.history.state, '', sanitizedURL);
    return tickets;
};
