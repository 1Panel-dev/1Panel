<template>
    <div>
        <Submenu activeName="system" />
        <el-card style="margin-top: 20px">
            <codemirror
                :autofocus="true"
                placeholder="None data"
                :indent-with-tab="true"
                :tabSize="4"
                style="height: calc(100vh - 150px)"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="logs"
                :readOnly="true"
            />
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { onMounted, ref } from 'vue';
import Submenu from '@/views/log/index.vue';
import { LoadFile } from '@/api/modules/files';

const extensions = [javascript(), oneDark];
const logs = ref();

const loadSystemlogs = async () => {
    const res = await LoadFile({ path: '/opt/1Panel/log/1Panel.log' });
    logs.value = res.data;
};

onMounted(() => {
    loadSystemlogs();
});
</script>
