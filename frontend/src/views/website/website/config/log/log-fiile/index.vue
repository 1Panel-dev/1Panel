<template>
    <div v-loading="loading">
        <codemirror
            :autofocus="true"
            placeholder="None data"
            :indent-with-tab="true"
            :tabSize="4"
            style="margin-top: 10px; max-height: 500px"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="content"
            :readOnly="true"
        />
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { GetFileContent } from '@/api/modules/files';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { computed, onMounted, ref } from 'vue';

const extensions = [javascript(), oneDark];
const props = defineProps({
    path: {
        type: String,
        default: '',
    },
});
const path = computed(() => {
    return props.path;
});
let loading = ref(false);
let content = ref('');

const getContent = () => {
    loading.value = true;
    GetFileContent({ path: path.value, expand: false, page: 1, pageSize: 1 })
        .then((res) => {
            content.value = res.data.content;
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    getContent();
});
</script>
