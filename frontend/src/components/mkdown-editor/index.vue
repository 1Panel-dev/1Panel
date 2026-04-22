<template>
    <MdEditor
        previewOnly
        :model-value="props.content"
        :sanitize="sanitizeHtml"
        :theme="isDarkTheme ? 'dark' : 'light'"
    />
</template>

<script lang="ts" setup>
import MdEditor from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import DOMPurify from 'dompurify';

import { useGlobalStore } from '@/composables/useGlobalStore';
const { isDarkTheme } = useGlobalStore();

const props = defineProps({
    content: {
        type: String,
        default: '',
    },
});

const sanitizeHtml = (html: string): string => {
    return DOMPurify.sanitize(html);
};
</script>
