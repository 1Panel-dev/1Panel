<template>
    <span v-for="(token, index) in tokens" :key="index" :class="['token', token.type]" :style="{ color: token.color }">
        <span class="whitespace-pre">{{ token.text }}</span>
    </span>
</template>
<script setup lang="ts">
import { ansiToJson } from 'anser';
interface TokenRule {
    type: string;
    pattern: RegExp;
    color: string;
}

interface Token {
    text: string;
    type: string;
    color: string;
    html?: string;
}

const props = defineProps<{
    log: string;
    type: string;
    container?: string;
}>();

const rules = ref<TokenRule[]>([]);
const nginxRules: TokenRule[] = [
    {
        type: 'log-level-warn',
        pattern: /\b(warn|warning|NOTICE)\b/g,
        color: '#F39C12',
    },
    {
        type: 'log-level-info',
        pattern: /\b(info|debug)\b/g,
        color: '#3498DB',
    },
    {
        type: 'log-level-error',
        pattern: /\[(crit|error)\]/g,
        color: '#E74C3C',
    },

    {
        type: 'path',
        pattern: /(?<=[\s"])\/[^"\s]+(?:\.\w+)?(?:\?\w+=\w+)?/g,
        color: '#B87A2B',
    },
    {
        type: 'http-method',
        pattern: /(?<=)(?:GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)(?=\s)/g,
        color: '#27AE60',
    },
    {
        type: 'status-success',
        pattern: /\s(2\d{2})\s/g,
        color: '#2ECC71',
    },
    {
        type: 'status-error',
        pattern: /\s([45]\d{2})\s/g,
        color: '#E74C3C',
    },
    {
        type: 'process-info',
        pattern: /\d+#\d+/g,
        color: '#7F8C8D',
    },
];

const systemRules: TokenRule[] = [
    {
        type: 'log-error',
        pattern: /\[(ERROR|WARN|FATAL)\]/g,
        color: '#E74C3C',
    },
    {
        type: 'log-normal',
        pattern: /\[(INFO|DEBUG)\]/g,
        color: '#8B8B8B',
    },
    {
        type: 'timestamp',
        pattern: /\[\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\]/g,
        color: '#8B8B8B',
    },
    {
        type: 'bracket-text',
        pattern: /\[([^\]]+)\]/g,
        color: '#B87A2B',
    },
    {
        type: 'referrer-ua',
        pattern: /https?:\/\/(?:[\w-]+\.)+[\w-]+(?::\d+)?(?:\/[^\s\]\)"]*)?/g,
        color: '#786C88',
    },
];

const taskRules: TokenRule[] = [
    {
        type: 'bracket-text',
        pattern: /\[(?:[^\[\]]*(?:\[[^\[\]]*\])*[^\[\]]*)*\]/g,
        color: '#B87A2B',
    },
];

const defaultRules: TokenRule[] = [
    {
        type: 'timestamp',
        pattern:
            /(?:\[\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2}\s[+-]\d{4}\]|\d{4}[-\/]\d{2}[-\/]\d{2}\s\d{2}:\d{2}:\d{2})/g,
        color: '#8B8B8B',
    },
    {
        type: 'referrer-ua',
        pattern: /"(?:https?:\/\/[^"]+|Mozilla[^"]+|curl[^"]+)"/g,
        color: '#786C88',
    },
    {
        type: 'ip',
        pattern: /\b(?<!\[)\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b(?!\])/g,
        color: '#4A90E2',
    },
    {
        type: 'ipv6',
        pattern: /\b(?:[A-Fa-f0-9]{1,4}:){7}[A-Fa-f0-9]{1,4}\b/g,
        color: '#4A90E2',
    },
    {
        type: 'server-host',
        pattern: /(?:server|host):\s*[^,\s]+/g,
        color: '#5D6D7E',
    },
];

const commonContainerRules: TokenRule[] = [
    {
        type: 'timestamp',
        pattern: /\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}/g,
        color: '#8B8B8B',
    },
    {
        type: 'path',
        pattern: /(?<=[\s"]|^)\/[^\s"]+/g,
        color: '#9B59B6',
    },
    {
        type: 'ip',
        pattern: /\b(?<!\[)\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b(?!\])/g,
        color: '#4A90E2',
    },
    {
        type: 'ipv6',
        pattern: /\b(?:[A-Fa-f0-9]{1,4}:){7}[A-Fa-f0-9]{1,4}\b/g,
        color: '#4A90E2',
    },
    {
        type: 'log-level-warn',
        pattern: /\b(WARN|WARNING|NOTICE)\b/g,
        color: '#F39C12',
    },
    {
        type: 'log-level-info',
        pattern: /\b(INFO|DEBUG)\b/g,
        color: '#3498DB',
    },
    {
        type: 'log-level-error',
        pattern: /\[(CRIT|ERROR)\]/g,
        color: '#E74C3C',
    },
    {
        type: 'url',
        pattern: /https?:\/\/(?:[\w-]+\.)+[\w-]+(?::\d+)?(?:\/[^\s\]\)"]*)?/g,
        color: '#786C88',
    },
];

const mysqlContainerRules: TokenRule[] = [
    {
        type: 'mysql-timestamp',
        pattern: /\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z/g,
        color: '#8B8B8D',
    },
    {
        type: 'mysql-thread-id',
        pattern: /(?<=T\d{2}:\d{2}:\d{2}\.\d{6}Z\s)\d+(?=\s)/g,
        color: '#7F8C8D',
    },
    {
        type: 'mysql-log-error',
        pattern: /\[(Error)\]/g,
        color: '#E74C3C',
    },
    {
        type: 'mysql-log-warn',
        pattern: /\[(Warning)\]/g,
        color: '#F39C12',
    },
    {
        type: 'mysql-log-info',
        pattern: /\[(System|Note|Entrypoint)\]/g,
        color: '#3498DB',
    },
    {
        type: 'mysql-error-code',
        pattern: /\[MY-\d{6}\]/g,
        color: '#9B59B6',
    },
    {
        type: 'mysql-component',
        pattern: /\[(Server|Repl|InnoDB|Audit|Query)\]/g,
        color: '#3498DB',
    },
];

const redisContainerRules: TokenRule[] = [
    {
        type: 'redis-role',
        pattern: /\b(Master|Slave|Replica)\b/g,
        color: '#E67E22',
    },
    {
        type: 'redis-signal',
        pattern: /\b(SIGTERM|SIGINT|SIGHUP)\b/g,
        color: '#9B59B6',
    },
    {
        type: 'redis-memory',
        pattern: /\d+\s*[KMG]B?\b/g,
        color: '#3498DB',
    },
];

const postgresContainerRules: TokenRule[] = [
    {
        type: 'postgres-severity',
        pattern: /\b(LOG|INFO|NOTICE|WARNING|ERROR|FATAL|PANIC)\b/g,
        color: '#E74C3C',
    },
    {
        type: 'postgres-process',
        pattern: /\[\d+\]/g,
        color: '#7F8C8D',
    },
    {
        type: 'postgres-statement',
        pattern: /\b(SELECT|INSERT|UPDATE|DELETE|CREATE|DROP|ALTER)\b/gi,
        color: '#27AE60',
    },
];

const phpContainerRules: TokenRule[] = [];

const getContainerRules = (containerName?: string): TokenRule[] => {
    if (!containerName) {
        return commonContainerRules;
    }

    const name = containerName.toLowerCase();
    let specificRules: TokenRule[] = [];

    if (name.includes('openresty') || name.includes('nginx')) {
        specificRules = nginxRules;
    } else if (name.includes('mysql') || name.includes('mariadb')) {
        specificRules = mysqlContainerRules;
    } else if (name.includes('redis')) {
        specificRules = redisContainerRules;
    } else if (name.includes('postgres') || name.includes('postgresql')) {
        specificRules = postgresContainerRules;
    } else if (name.includes('php')) {
        specificRules = phpContainerRules;
    }

    return [...specificRules, ...commonContainerRules];
};

function tokenizeLog(log: string): Token[] {
    if (
        log.indexOf('<div') !== -1 ||
        log.indexOf('<span') !== -1 ||
        log.indexOf('<p>') !== -1 ||
        log.indexOf('<img') !== -1 ||
        log.indexOf('</div>') !== -1 ||
        log.indexOf('</span>') !== -1 ||
        log.indexOf('<script>') !== -1
    ) {
        return [
            {
                text: log,
                type: 'plain',
                color: '#666666',
            },
        ];
    }

    if (log.indexOf('\x1b[') !== -1) {
        try {
            const parsed = ansiToJson(log, { json: true, remove_empty: true });
            return parsed.map((item: any) => ({
                text: item.content,
                type: 'ansi',
                color: item.fg ? `rgb(${item.fg})` : '#666666',
            }));
        } catch (error) {
            return [
                {
                    text: log,
                    type: 'plain',
                    color: '#666666',
                },
            ];
        }
    }

    const textLength = log.length;
    if (textLength > 5000) {
        return [
            {
                text: log,
                type: 'plain',
                color: '#666666',
            },
        ];
    }

    const tokens: Token[] = [];
    let lastIndex = 0;
    const matches: { index: number; text: string; type: string; color: string }[] = [];

    for (const rule of rules.value) {
        const regex = new RegExp(rule.pattern.source, rule.pattern.flags);
        let match;
        while ((match = regex.exec(log)) !== null) {
            matches.push({
                index: match.index,
                text: match[0],
                type: rule.type,
                color: rule.color,
            });
        }
    }

    matches.sort((a, b) => a.index - b.index);

    const filteredMatches = matches.filter((match, index) => {
        if (index === 0) return true;
        const prev = matches[index - 1];
        return match.index >= prev.index + prev.text.length;
    });

    for (const match of filteredMatches) {
        if (match.index > lastIndex) {
            tokens.push({
                text: log.substring(lastIndex, match.index),
                type: 'plain',
                color: '#666666',
            });
        }
        tokens.push({
            text: match.text,
            type: match.type,
            color: match.color,
        });
        lastIndex = match.index + match.text.length;
    }

    if (lastIndex < log.length) {
        const rest = log.substring(lastIndex).replace(/\n?$/, '\n');
        tokens.push({
            text: rest,
            type: 'plain',
            color: '#666666',
        });
    }

    return tokens;
}

const tokens = computed(() => tokenizeLog(props.log));

const initRules = () => {
    switch (props.type) {
        case 'nginx':
            return [...nginxRules, ...defaultRules];
        case 'system':
            return [...systemRules, ...defaultRules];
        case 'container':
            return [...getContainerRules(props.container), ...defaultRules];
        case 'task':
            return [...taskRules, ...defaultRules];
        default:
            return defaultRules;
    }
};

watchEffect(() => {
    rules.value = initRules();
});
</script>

<style scoped>
.token {
    font-family: 'JetBrains Mono', Monaco, Menlo, Consolas, 'Courier New', monospace;
    font-size: 14px;
    font-weight: 500;
}

.ip {
    text-decoration: underline;
    text-decoration-style: dotted;
    text-decoration-thickness: 1px;
}
</style>
