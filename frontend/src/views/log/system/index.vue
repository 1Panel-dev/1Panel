<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.system')">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button class="no-active-button" @click="onChangeRoute('OperationLog')">
                            {{ $t('logs.operation') }}
                        </el-button>
                        <el-button class="no-active-button" @click="onChangeRoute('LoginLog')">
                            {{ $t('logs.login') }}
                        </el-button>
                        <el-button type="primary" @click="onChangeRoute('SystemLog')">
                            {{ $t('logs.system') }}
                        </el-button>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <codemirror
                    :autofocus="true"
                    :placeholder="$t('commons.msg.noneData')"
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="height: calc(100vh - 290px)"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    @ready="handleReady"
                    v-model="logs"
                    :disabled="true"
                />
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { nextTick, onMounted, ref, shallowRef } from 'vue';
import { LoadFile } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';
import { useRouter } from 'vue-router';
const router = useRouter();

const loading = ref();
const extensions = [javascript(), oneDark];
const logs = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const loadSystemlogs = async () => {
    const pathRes = await loadBaseDir();
    let logPath = pathRes.data + '/log';
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

const onChangeRoute = async (addr: string) => {
    router.push({ name: addr });
};

onMounted(() => {
    loadSystemlogs();
});
</script>
