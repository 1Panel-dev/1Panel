<template>
    <MdEditor previewOnly v-model="sanitizedReadMe" :theme="isDarkTheme ? 'dark' : 'light'" />
</template>

<script lang="ts" setup>
import MdEditor from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import DOMPurify from 'dompurify';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import { computed } from 'vue';

const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);
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
