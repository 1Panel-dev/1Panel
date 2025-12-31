<template>
    <MdEditor previewOnly v-model="sanitizedReadMe" :theme="isDarkTheme ? 'dark' : 'light'" />
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

const sanitizedReadMe = computed(() => {
    return DOMPurify.sanitize(props.content);
});
</script>
