<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.system')">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button class="tag-button no-active" @click="onChangeRoute('OperationLog')">
                            {{ $t('logs.operation') }}
                        </el-button>
                        <el-button class="tag-button no-active" @click="onChangeRoute('LoginLog')">
                            {{ $t('logs.login') }}
                        </el-button>
                        <el-button class="tag-button" type="primary" @click="onChangeRoute('SystemLog')">
                            {{ $t('logs.system') }}
                        </el-button>
                    </el-col>
                </el-row>
            </template>
            <template #search>
                <el-select class="float-left" v-model="currentFile" @change="search()">
                    <template #prefix>{{ $t('commons.button.log') }}</template>
                    <el-option v-for="(item, index) in fileList" :key="index" :label="item" :value="item" />
                </el-select>
                <div class="watchCheckbox">
                    <el-checkbox border @change="changeWatch" v-model="isWatch">
                        {{ $t('commons.button.watch') }}
                    </el-checkbox>
                </div>
            </template>
            <template #main>
                <codemirror
                    :autofocus="true"
                    :placeholder="$t('commons.msg.noneData')"
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="height: calc(100vh - 370px)"
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
import { nextTick, onBeforeUnmount, onMounted, ref, shallowRef } from 'vue';
import { useRouter } from 'vue-router';
import { getSystemFiles, getSystemLogs } from '@/api/modules/log';
const router = useRouter();

const loading = ref();
const isWatch = ref();
const currentFile = ref();
const fileList = ref();

const extensions = [javascript(), oneDark];
const logs = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

let timer: NodeJS.Timer | null = null;

const changeWatch = async () => {
    if (isWatch.value) {
        timer = setInterval(() => {
            search();
        }, 1000 * 3);
    } else {
        if (timer) {
            clearInterval(Number(timer));
            timer = null;
        }
    }
};

const loadFiles = async () => {
    const res = await getSystemFiles();
    fileList.value = res.data || [];
    if (fileList.value) {
        currentFile.value = fileList.value[0];
        search();
    }
};

const search = async () => {
    await getSystemLogs(currentFile.value)
        .then((res) => {
            loading.value = false;
            logs.value = res.data.replace(/\u0000/g, '');
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

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});

onMounted(() => {
    loadFiles();
});
</script>

<style scoped lang="scss">
.watchCheckbox {
    margin-top: 2px;
    margin-bottom: 10px;
    float: left;
    margin-left: 20px;
}
</style>
