<template>
    <span v-for="(token, index) in tokens" :key="index" :class="['token', token.type]" :style="{ color: token.color }">
        {{ token.text }}
    </span>
</template>
<script setup lang="ts">
interface TokenRule {
    type: string;
    pattern: RegExp;
    color: string;
}

interface Token {
    text: string;
    type: string;
    color: string;
}

const props = defineProps<{
    log: string;
    type: string;
}>();

let rules = ref<TokenRule[]>([]);
const nginxRules: TokenRule[] = [
    {
        type: 'log-level',
        pattern: /\[(error|warn|notice|info|debug)\]/gi,
        color: '#E74C3C',
    },
    {
        type: 'path',
        pattern:
            /(?:(?<=GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+|(?<=open\(\s*")|(?<="\s*))(\/[^"\s]+(?:\.\w+)?(?:\?\w+=\w+)?)/g,
        color: '#B87A2B',
    },
    {
        type: 'http-method',
        pattern: /(?<=")(?:GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)(?=\s)/g,
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
        pattern: /\[([^\]]+)\]/g,
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
        pattern: /\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/g,
        color: '#4A90E2',
    },
    {
        type: 'server-host',
        pattern: /(?:server|host):\s*[^,\s]+/g,
        color: '#5D6D7E',
    },
];

const containerRules: TokenRule[] = [
    {
        type: 'timestamp',
        pattern: /\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}/g,
        color: '#8B8B8B',
    },
    {
        type: 'bracket-text',
        pattern: /\[([^\]]+)\]/g,
        color: '#B87A2B',
    },
];

function tokenizeLog(log: string): Token[] {
    const tokens: Token[] = [];
    let lastIndex = 0;
    let matches: { index: number; text: string; type: string; color: string }[] = [];

    rules.value.forEach((rule) => {
        const regex = new RegExp(rule.pattern.source, 'g');
        let match;
        while ((match = regex.exec(log)) !== null) {
            matches.push({
                index: match.index,
                text: match[0],
                type: rule.type,
                color: rule.color,
            });
        }
    });

    matches.sort((a, b) => a.index - b.index);

    matches = matches.filter((match, index) => {
        if (index === 0) return true;
        const prev = matches[index - 1];
        return match.index >= prev.index + prev.text.length;
    });

    matches.forEach((match) => {
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
    });

    if (lastIndex < log.length) {
        tokens.push({
            text: log.substring(lastIndex),
            type: 'plain',
            color: '#666666',
        });
    }

    return tokens;
}

const tokens = computed(() => tokenizeLog(props.log));

onMounted(() => {
    switch (props.type) {
        case 'nginx':
            rules.value = nginxRules.concat(defaultRules);
            break;
        case 'system':
            rules.value = systemRules.concat(defaultRules);
            break;
        case 'container':
            rules.value = containerRules.concat(defaultRules);
            break;
        case 'task':
            rules.value = taskRules.concat(defaultRules);
            break;
        default:
            rules.value = defaultRules;
            break;
    }
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
