<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.system')">
            <template #main>
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
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { Codemirror } from 'vue-codemirror';
import LayoutContent from '@/layout/layout-content.vue';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { nextTick, onMounted, ref, shallowRef } from 'vue';
import { LoadFile } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';

const loading = ref();
const extensions = [javascript(), oneDark];
const logs = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const loadSystemlogs = async () => {
    const pathRes = await loadBaseDir();
    let logPath = pathRes.data.replaceAll('/data', '/log');
    await LoadFile({ path: `${logPath}/1Panel.log` })
        .then((res) => {
            loading.value = false;
            logs.value = res.data;
            nextTick(() => {
                const state = view.value.state;
                view.value.dispatch({
                    selection: { anchor: state.doc.length, head: state.doc.length },
                    scrollIntoView: true,
                });
            });
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    loadSystemlogs();
});
</script>
