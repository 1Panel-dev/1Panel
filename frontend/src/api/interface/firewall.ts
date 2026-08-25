import { ReqPage } from '.';

export namespace Firewall {
    export type Provider = 'iptables' | 'nftables' | 'firewalld' | 'ufw';
    export type BackendSubsystem = 'system' | 'forwarding' | 'docker';
    export type BackendOperation = 'select' | 'initialize' | 'cleanup';
    export interface BackendOption {
        name: Provider;
        installed: boolean;
        active: boolean;
        initialized: boolean;
        bound: boolean;
        supported: boolean;
        implementation?: string;
        message?: string;
        ipv4: BackendFamilyStatus;
        ipv6: BackendFamilyStatus;
    }
    export interface BackendFamilyStatus {
        available: boolean;
        initialized: boolean;
        bound: boolean;
        reason?: string;
    }
    export interface BackendGroup {
        selected: string;
        current?: string;
        options: BackendOption[];
    }
    export interface Settings {
        system: BackendGroup;
        forwarding: BackendGroup;
        docker: BackendGroup;
        pingStatus: string;
        portWhiteList: string;
    }
    export interface BackendOperateRequest {
        subsystem: BackendSubsystem;
        backend: Provider;
        operation: BackendOperation;
    }
    export interface FirewallBase {
        name: string;
        backend: string;
        conflictBackend?: string;
        isExist: boolean;
        isActive: boolean;
        isInit: boolean;
        isBind: boolean;
        version: string;
        pingStatus: string;
        syncError?: string;
        ipv4: BackendFamilyStatus;
        ipv6: BackendFamilyStatus;
    }
    export interface ForwardRuleSearch extends ReqPage {
        strategy: string;
        info: string;
    }
    export interface RuleInfo extends ReqPage {
        family: string;
        address: string;
        destination: string;
        port: string;
        srcPort: string;
        destPort: string;
        protocol: string;
        strategy: string;
        usedStatus: string;
        description: string;
        [key: string]: any;
    }
    export interface RuleForward {
        operation: string;
        family: 'ipv4' | 'ipv6';
        protocol: string;
        port: string;
        targetIP: string;
        targetPort: string;
        interface: string;
        isDesired?: boolean;
        isRuntime?: boolean;
        syncStatus?: 'converged' | 'missing' | 'runtime_only';
    }
    export type Family = 'ipv4' | 'ipv6' | 'inet';
    export type Direction = 'input';
    export type Action = 'accept' | 'drop' | 'reject';
    export type NativeKind =
        'rule' | 'zone_port' | 'rich_rule' | 'ufw_rule' | 'ufw_application' | 'opaque' | 'zone_service';
    export type ParseStatus = 'supported' | 'partial' | 'opaque';
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

    export interface ResetResponse {
        removed: number;
        disabled: boolean;
    }

    export interface ResetRequest {
        provider?: Provider;
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

    export interface NativeDetailRequest {
        provider: 'firewalld' | 'ufw';
        nativeKind: 'zone_service' | 'ufw_application';
        name: string;
        permanent: boolean;
    }

    export interface CheckItem {
        uuid?: string;
        rule: Rule;
    }

    export interface CheckRequest {
        items: CheckItem[];
    }

    export interface CheckResponse {
        items: RuleCheckResult[];
    }

    export interface CreateItem {
        rule: Rule;
        checkFlag: string;
        action: ApplicableCheckAction;
        adoptInstanceKey?: string;
        sourceKind?: 'user' | 'panel' | 'security' | 'imported';
        sourceID?: string;
    }

    export interface CreateRequest {
        items: CreateItem[];
    }

    export interface CreateResponse {
        succeeded: number;
        failed: number;
        skipped: number;
        errors?: CreateFailure[];
    }

    export interface CreateFailure {
        index: number;
        status: 'failed' | 'skipped';
        rule: Rule;
        error?: string;
    }

    export interface RuleSyncRequest {
        subsystem: BackendSubsystem;
        sourceProvider?: Provider;
        targetProvider: Provider;
        resetSource?: boolean;
        taskID?: string;
    }

    export type RuleSyncStatus = 'ready' | 'existing' | 'remove' | 'blocked';

    export interface RuleSyncItem {
        sourceUUID: string;
        rule?: Rule;
        forwardRule?: RuleForward;
        dockerRule?: DockerGuardEndpoint;
        status: RuleSyncStatus;
        reason?: string;
    }

    export interface RuleSyncPreview {
        subsystem: BackendSubsystem;
        sourceProvider?: Provider;
        targetProvider: Provider;
        total: number;
        ready: number;
        existing: number;
        removed: number;
        blocked: number;
        items: RuleSyncItem[];
    }

    export interface RuleSyncResult {
        subsystem: BackendSubsystem;
        sourceProvider?: Provider;
        targetProvider: Provider;
        total: number;
        succeeded: number;
        skipped: number;
        removed: number;
        failed: number;
        errors?: RuleSyncFailure[];
        taskID?: string;
        queued?: boolean;
    }

    export interface RuleSyncTask {
        taskID?: string;
        executing: boolean;
    }

    export interface RuleSyncFailure {
        sourceUUID: string;
        rule?: Rule;
        forwardRule?: RuleForward;
        dockerRule?: DockerGuardEndpoint;
        error: string;
    }

    export interface DeleteRequest {
        uuids: string[];
    }

    export interface DeleteResponse {
        succeeded: number;
        failed: number;
        errors?: DeleteFailure[];
    }

    export interface DeleteFailure {
        index: number;
        uuid: string;
        error: string;
    }

    export interface UpdateRequest {
        rule: Rule;
    }

    export interface ReorderRequest {
        targetPosition?: number;
        priority?: number;
    }

    export interface DockerGuardBase {
        name: string;
        version: string;
        initialized: boolean;
        bound: boolean;
        ipv4: DockerGuardFamilyStatus;
        ipv6: DockerGuardFamilyStatus;
        backend: string;
        message?: string;
    }
    export interface DockerGuardFamilyStatus {
        state: 'effective' | 'disabled' | 'not_effective';
        reason?:
            | 'command_missing'
            | 'docker_chain_missing'
            | 'guard_chain_missing'
            | 'jump_missing'
            | 'jump_not_first'
            | 'jump_duplicate'
            | 'inspect_failed';
        initialized: boolean;
        bound: boolean;
        effective: boolean;
    }
    export interface DockerGuardEndpoint {
        family: 'ipv4' | 'ipv6';
        hostIP: string;
        hostPort: number;
        protocol: 'tcp' | 'udp';
        containerID?: string;
        containerName?: string;
        containerPort?: number;
        compose?: string;
        application?: string;
        policyUUID?: string;
        mode?: 'deny_sources' | 'allow_sources' | 'deny_all';
        sources: string[];
        effective: boolean;
        description?: string;
    }
    export interface DockerGuardPortGroup {
        key: string;
        label: string;
        endpoint: DockerGuardEndpoint;
        endpoints: DockerGuardEndpoint[];
    }
    export interface DockerGuardContainer {
        key: string;
        name: string;
        compose?: string;
        application?: string;
        endpoints: DockerGuardEndpoint[];
        portGroups: DockerGuardPortGroup[];
    }
    export interface DockerGuardList {
        base: DockerGuardBase;
        containers: DockerGuardContainer[];
        orphanPolicies: DockerGuardEndpoint[];
    }
    export interface DockerGuardEndpointIdentity {
        family: 'ipv4' | 'ipv6';
        hostIP: string;
        hostPort: number;
        protocol: 'tcp' | 'udp';
    }
    export interface DockerGuardPolicy extends DockerGuardEndpointIdentity {
        mode: 'deny_sources' | 'allow_sources' | 'deny_all';
        sources: string[];
        description: string;
    }
    export interface DockerGuardPolicyBatch {
        endpoints: DockerGuardEndpointIdentity[];
        mode: 'deny_sources' | 'allow_sources' | 'deny_all';
        sources: string[];
        description: string;
    }
    export interface DockerGuardPolicyBatchDelete {
        uuids: string[];
    }
}
