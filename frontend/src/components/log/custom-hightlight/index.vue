<template>
    <code ref="codeRef" class="log-highlight whitespace-pre" v-html="highlightedLog" />
</template>
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import anser from 'anser';
const props = defineProps<{
    log: string;
    type: string;
}>();

const codeRef = ref<HTMLElement | null>(null);
const highlightedLog = ref('');

interface HighlightRule {
    regex: RegExp;
    className: string;
}

const highlightRules: HighlightRule[] = [
    { regex: /\b(INFO|TRACE|System|Note|notice)\b/gi, className: 'hljs-keyword' },
    { regex: /\b(ERROR|DEBUG|FATAL)\b/gi, className: 'hljs-error' },
    { regex: /\b(WARN|WARNING)\b/gi, className: 'hljs-warn' },
    { regex: /\b(true|false|null)\b/g, className: 'hljs-literal' },
    { regex: /\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?\b/g, className: 'hljs-number' },
    { regex: /\b\d+:[A-Z]\b/g, className: 'hljs-symbol' },
    { regex: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
    { regex: /([/~]?[A-Za-z0-9._-]{1,255}(?:\/[A-Za-z0-9._-]{1,255})+)/g, className: 'hljs-string' },
    { regex: /\b\d{1,3}(?:\.\d{1,3}){3}\b/g, className: 'hljs-attr' },
    { regex: /\b(SELECT|INSERT|UPDATE|DELETE|FROM|WHERE|JOIN|ON|CREATE|DROP|ALTER)\b/gi, className: 'hljs-built_in' },
    { regex: /\[?(Thread|PID)[-:]?\d+\]?/gi, className: 'hljs-symbol' },
    { regex: /\b([A-Za-z0-9_\-./]+\.([a-z]+)):(\d+)\b/g, className: 'hljs-title' },
    { regex: /\bhttps?:\/\/[^\s]+/gi, className: 'hljs-link' },
    { regex: /\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b/g, className: 'hljs-link' },
    { regex: /\b[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\b/gi, className: 'hljs-meta' },
    { regex: /\b\d+(\.\d+)?(%|ms|s|GB|MB|KB)?\b/g, className: 'hljs-number' },
    { regex: /(['])(?:\\.|[^\1\\])*?\1/g, className: 'hljs-string' },
    { regex: /[{}\[\]()|+*%&^~!]/g, className: 'hljs-symbol' },
];

function extraHighlightPlugin(html: string, rules: HighlightRule[] = highlightRules): string {
    let result = html;
    for (const rule of rules) {
        result = result.replace(rule.regex, `<span class="${rule.className}">$&</span>`);
    }
    return result;
}

function hasANSICodes(text: string) {
    const ansiRegex = /\x1b\[[0-9;]*[mK]/;
    return ansiRegex.test(text);
}

const highlightContent = (): string => {
    if (!props.log) return '';
    if (hasANSICodes(props.log)) {
        return anser.ansiToHtml(props.log);
    } else {
        return extraHighlightPlugin(props.log);
    }
};

watch(
    () => [props.log, props.type],
    () => {
        highlightedLog.value = highlightContent();
    },
    { immediate: true },
);

onMounted(() => {
    highlightedLog.value = highlightContent();
});
</script>

<style scoped>
.log-highlight {
    color: #e06c75;
    font-size: 14px;
    line-height: inherit;
    white-space: inherit;
    text-align: left;
}

:deep(.hljs-attr) {
    color: #56b6c2;
}
:deep(.hljs-built_in, .hljs-meta) {
    color: #e6c07b;
}
:deep(.hljs-title) {
    color: #d19a66;
}
:deep(.hljs-symbol) {
    color: #61afef;
}
:deep(.hljs-link) {
    color: #56b6c2;
}
:deep(.hljs-debug) {
    color: #61aeee !important;
}
:deep(.hljs-info) {
    color: #00bb00 !important;
}
:deep(.hljs-warn) {
    color: #bbbb00 !important;
}
:deep(.hljs-error) {
    color: #f91306 !important;
    font-weight: bold;
}
</style>
