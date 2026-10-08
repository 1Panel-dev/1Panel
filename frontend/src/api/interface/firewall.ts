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
        supportReason?: string;
        implementation?: string;
        message?: string;
        ipv4: BackendFamilyStatus;
        ipv6: BackendFamilyStatus;
    }
    export interface BackendFamilyStatus {
        partial?: boolean;
        available: boolean;
        initialized: boolean;
        bound: boolean;
        reason?: string;
        forwardPolicy?: 'ACCEPT' | 'DROP';
        raInterfaces?: string[];
    }
    export interface BackendGroup {
        selected: string;
        current?: string;
        options: BackendOption[];
    }
    export interface Settings {
        ipv6Enabled?: boolean;
        system: BackendGroup;
        forwarding: BackendGroup;
        docker: BackendGroup;
        pingStatus: string;
        portWhiteList: PortWhitelist[];
        panelPort: string;
        sshPort: string;
    }
    export interface PortWhitelist {
        port?: string;
        protocol?: 'tcp' | 'udp';
        type?: 'panel' | 'ssh';
        sources: string[];
    }
    export interface PortWhitelistUpdate {
        oldRule: PortWhitelist;
        rule: PortWhitelist;
    }

    export interface BackendOperateRequest {
        subsystem: BackendSubsystem;
        backend: Provider;
        operation: BackendOperation;
    }
    export interface FilterChainOperationResult {
        taskID?: string;
        queued?: boolean;
    }
    export interface FirewallBase {
        ipv6Enabled?: boolean;
        lifecycleTaskID?: string;
        name: string;
        backend: string;
        conflictBackend?: string;
        isExist: boolean;
        isActive: boolean;
        isInit: boolean;
        isBind: boolean;
        version: string;
        pingStatus: string;
        message?: string;
        reason?: string;
        ipv4: BackendFamilyStatus;
        ipv6: BackendFamilyStatus;
    }
    export interface ForwardRuleSearch extends ReqPage {
        all?: boolean;
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
        raw?: string;
        parseStatus?: ParseStatus;
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
        uncertainFields?: string[];
        raw?: string;
        protected: boolean;
        persistence?: PersistenceStatus;
    }

    export interface PositionRange {
        min: number;
        max: number;
    }

    export interface InventoryItem {
        rule: Rule;
        observed: ObservedRule;
        isWhitelist: boolean;
        descriptionID: string;
    }

    export type ScopeNoticeCode =
        | 'family_unavailable'
        | 'managed_scope_inactive'
        | 'runtime_permanent_mismatch'
        | 'default_policy'
        | 'managed_scope_missing';

    export interface ScopeNotice {
        code: ScopeNoticeCode;
        values?: string[];
    }

    export interface Inventory {
        ipv4Range: PositionRange;
        ipv6Range: PositionRange;
        total: number;
        allTotal: number;
        items: InventoryItem[];
        notices?: ScopeNotice[];
    }

    export interface ResetResponse {
        backupPath: string;
        removed: number;
        disabled: boolean;
    }

    export interface RuleBackups {
        directory: string;
        files: Array<{ name: string; provider: Provider; ruleCount: number; modifiedAt: number }>;
    }

    export interface ResetRequest {
        subsystem?: BackendSubsystem;
        backup?: boolean;
        provider?: Provider;
        withDockerRestart?: boolean;
    }

    export interface InventoryRequest extends ReqPage {
        scopes: Scope[];
        all?: boolean;
        info: string;
        families?: Array<'ipv4' | 'ipv6'>;
        actions?: Array<'accept' | 'deny'>;
        excludeChains?: string[];
    }

    export interface NativeDetailRequest {
        provider: 'firewalld' | 'ufw';
        nativeKind: 'zone_service' | 'ufw_application';
        name: string;
        permanent: boolean;
    }

    export interface CreateItem {
        raw?: string;
        parseStatus?: ParseStatus;
        rule: Rule;
        sourceKind?: 'user' | 'panel' | 'security' | 'imported';
    }

    export interface CreateRequest {
        backupFile?: string;
        initialize?: boolean;
        items: CreateItem[];
    }

    export interface CreateResponse {
        taskID?: string;
        queued?: boolean;
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

    export interface RuleTarget {
        scope: Scope;
        instanceKey: string;
    }

    export interface DeleteRequest {
        targets: (RuleTarget & { observed: ObservedRule })[];
    }

    export interface DeleteResponse {
        taskID?: string;
        queued?: boolean;
        succeeded: number;
        failed: number;
        errors?: DeleteFailure[];
    }

    export interface DeleteFailure {
        index: number;
        instanceKey: string;
        error: string;
    }

    export type UpdateRequest =
        | { rule: Rule; description?: never; orderIndex?: never; priority?: never }
        | { rule?: never; description: string; orderIndex?: never; priority?: never }
        | { rule?: never; description?: string; orderIndex: number; priority?: never }
        | { rule?: never; description?: string; orderIndex?: never; priority: number };

    export interface DockerGuardBase {
        ipv6Enabled?: boolean;
        name: string;
        version: string;
        isExist: boolean;
        initialized: boolean;
        bound: boolean;
        ipv4: DockerGuardFamilyStatus;
        ipv6: DockerGuardFamilyStatus;
        backend: string;
        message?: string;
    }
    export interface DockerGuardFamilyStatus {
        partial?: boolean;
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
        containerState?: 'created' | 'running' | 'paused' | 'restarting' | 'removing' | 'exited' | 'dead';
        containerPort?: number;
        compose?: string;
        application?: string;
        policyUUID?: string;
        mode?: 'deny_sources' | 'allow_sources' | 'deny_all' | 'accept_sources' | 'accept_all';
        sources: string[];
        effective: boolean;
        trafficPath: 'forward' | 'input' | 'unknown';
        managementTarget?: 'container_guard' | 'host_firewall' | 'needs_diagnosis';
        managementReason?: 'nat_inspect_failed' | 'nat_chain_unreachable' | 'proxy_inspect_failed' | 'no_matching_path';
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
        mode: 'deny_sources' | 'allow_sources' | 'deny_all' | 'accept_sources' | 'accept_all';
        sources: string[];
    }
    export interface DockerGuardPolicyBatch {
        policies: DockerGuardPolicy[];
        import?: boolean;
    }
    export interface DockerGuardPolicyBatchDelete {
        uuids: string[];
    }
}
