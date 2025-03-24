<template>
    <highlightjs
        ref="editorRef"
        language="JavaScript"
        :style="customStyle"
        :autodetect="false"
        :code="content"
    ></highlightjs>
</template>

<script lang="ts" setup>
import { CSSProperties } from 'vue';

const editorRef = ref();
const scrollerElement = ref<HTMLElement | null>(null);

const props = defineProps({
    modelValue: {
        type: String,
        default: '',
    },
    heightDiff: {
        type: Number,
        default: 280,
    },
});

const content = ref('');

const customStyle = computed<CSSProperties>(() => ({
    width: '100%',
    overflow: 'auto',
}));

const initLog = async () => {
    await nextTick();
    if (editorRef.value && scrollerElement.value == undefined) {
        const parentElement = editorRef.value.$el as HTMLElement;
        scrollerElement.value = parentElement.querySelector('.hljs') as HTMLElement;
        scrollerElement.value.style['min-height'] = '100px';
        scrollerElement.value.style['max-height'] = 'calc(100vh - ' + props.heightDiff + 'px)';
    }
};

watch(
    () => props.modelValue,
    async (newValue) => {
        if (editorRef.value && scrollerElement.value != undefined && newValue != content.value) {
            content.value = newValue;
            scrollerElement.value.scrollTop = scrollerElement.value.scrollHeight;
        } else {
            initLog();
        }
    },
    { immediate: true },
);

onMounted(() => {
    initLog();
});
</script>
