import { Firewall } from '@/api/interface/firewall';

export const buildHostRuleExport = (items: Firewall.InventoryItem[]) =>
    items
        .filter(
            ({ rule }) =>
                !['iptables', 'nftables'].includes(rule.scope.provider) || rule.scope.chain === '1PANEL_BASIC',
        )
        .map(({ rule, observed }) => ({
            ...rule,
            uuid: undefined,
            orderIndex: undefined,
            raw: observed.raw,
            parseStatus: observed.parseStatus,
            scope: { ...rule.scope },
        }));
