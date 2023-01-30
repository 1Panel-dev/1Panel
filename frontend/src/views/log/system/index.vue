<template>
    <div>
        <Submenu activeName="system" />
        <el-card style="margin-top: 20px">
            <LayoutContent :header="$t('logs.system')">
                <codemirror
                    :autofocus="true"
                    placeholder="None data"
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="height: calc(100vh - 180px)"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    @ready="handleReady"
                    v-model="logs"
                    :readOnly="true"
                />
            </LayoutContent>
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { Codemirror } from 'vue-codemirror';
import LayoutContent from '@/layout/layout-content.vue';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { nextTick, onMounted, ref, shallowRef } from 'vue';
import Submenu from '@/views/log/index.vue';
import { LoadFile } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';

const extensions = [javascript(), oneDark];
const logs = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const loadSystemlogs = async () => {
    const pathRes = await loadBaseDir();
    let logPath = pathRes.data.replaceAll('/data', '/log');
    const res = await LoadFile({ path: `${logPath}/1Panel.log` });
    logs.value = res.data;
    nextTick(() => {
        const state = view.value.state;
        view.value.dispatch({
            selection: { anchor: state.doc.length, head: state.doc.length },
            scrollIntoView: true,
        });
    });
};

onMounted(() => {
    loadSystemlogs();
});
</script>
