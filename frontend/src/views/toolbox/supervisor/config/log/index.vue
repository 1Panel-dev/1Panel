<template>
    <div v-loading="loading">
        <codemirror
            :autofocus="true"
            :placeholder="$t('commons.msg.noneData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="height: calc(100vh - 375px)"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="content"
        />
    </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { GetSupervisorLog } from '@/api/modules/host-tool';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';

const extensions = [javascript(), oneDark];

let content = ref('');
let loading = ref(false);

const getConfig = async () => {
    const res = await GetSupervisorLog();
    content.value = res.data;
};

onMounted(() => {
    getConfig();
});
</script>
