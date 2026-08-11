export namespace Firewall {
    export type Provider = 'iptables' | 'firewalld' | 'ufw';
    export type Family = 'ipv4' | 'ipv6' | 'inet';
    export type Direction = 'input';
    export type Action = 'accept' | 'drop' | 'reject';
    export type NativeKind = 'rule' | 'zone_port' | 'rich_rule' | 'ufw_rule' | 'opaque' | 'zone_service';
    export type ParseStatus = 'supported' | 'opaque';
    export type PersistenceStatus = 'converged' | 'runtime_only' | 'permanent_only';

    export interface Scope {
        provider: Provider;
        family: Family;
        table?: string;
        zone?: string;
        chain?: string;
        direction: Direction;
    }

    export interface Rule {
        uuid?: string;
        scope: Scope;
        nativeKind?: NativeKind;
        protocol: string;
        sourceAddress?: string;
        sourcePort?: string;
        destinationAddress?: string;
        destinationPort?: string;
        interface?: string;
        connectionStates?: string[];
        action: Action;
        priority?: number;
        orderIndex?: number;
        orderBucket?: string;
        description?: string;
    }

    export interface Locator {
        provider: Provider;
        scopeKey: string;
        nativeId?: string;
        canonical?: string;
        position?: number;
    }

    export interface ObservedRule {
        rule: Rule;
        locator: Locator;
        instanceKey?: string;
        marker?: string;
        parseStatus: ParseStatus;
        raw?: string;
        protected: boolean;
        persistence?: PersistenceStatus;
    }

    export type RuleOrigin = 'created' | 'adopted';
    export type InventoryState = 'managed' | 'adopted' | 'external' | 'drifted' | 'protected';
    export type InventoryMatch = 'none' | 'exact' | 'changed' | 'missing' | 'ambiguous' | 'opaque';

    export interface DesiredRule {
        uuid: string;
        rule: Rule;
        ruleKey: string;
        origin: RuleOrigin;
        marker?: string;
        observedInstanceKey?: string;
    }

    export interface RuntimeUsage {
        used: boolean;
        usedBy?: string[];
        reason?: string;
    }

    export interface InventoryItem {
        rule: Rule;
        observed?: ObservedRule;
        desired?: DesiredRule;
        state: InventoryState;
        match: InventoryMatch;
        usage?: RuntimeUsage;
    }

    export type ScopeNoticeCode =
        | 'default_scope_mismatch'
        | 'managed_scope_inactive'
        | 'unmanaged_active_scopes'
        | 'runtime_permanent_mismatch'
        | 'default_policy'
        | 'managed_scope_missing';

    export interface ScopeNotice {
        code: ScopeNoticeCode;
        values?: string[];
    }

    export interface Inventory {
        items: InventoryItem[];
        notices?: ScopeNotice[];
    }

    export type CheckDecision = 'ready' | 'confirmation_required' | 'blocked' | 'no_change';
    export type CheckClassification =
        'none' | 'exact_managed' | 'exact_external' | 'covered' | 'conflict' | 'unsupported' | 'protected';
    export type CheckAction = 'create' | 'create_anyway' | 'adopt' | 'select_adopt' | 'cancel';
    export type ApplicableCheckAction = Exclude<CheckAction, 'cancel'>;

    export interface RuleCheckResult {
        decision: CheckDecision;
        classification: CheckClassification;
        reason: string;
        requestedRule: Rule;
        requestedRuleKey: string;
        existingRuleUUID?: string;
        candidates?: ObservedRule[];
        allowedActions?: CheckAction[];
        checkFlag: string;
    }

    export interface InventoryRequest {
        scope: Scope;
    }

    export interface CheckRequest {
        uuid?: string;
        rule: Rule;
    }

    export interface BatchCheckRequest {
        rules: Rule[];
    }

    export interface BatchCheckResponse {
        items: RuleCheckResult[];
    }

    export interface CreateRequest {
        rule: Rule;
        checkFlag: string;
        action: ApplicableCheckAction;
        adoptInstanceKey?: string;
        sourceKind?: 'user' | 'panel' | 'security' | 'imported';
        sourceID?: string;
    }

    export interface BatchCreateRequest {
        items: CreateRequest[];
    }

    export interface BatchCreateResponse {
        succeeded: number;
        failed: number;
    }

    export interface UpdateRequest {
        rule: Rule;
    }

    export interface ReorderRequest {
        targetPosition?: number;
        priority?: number;
    }
}
