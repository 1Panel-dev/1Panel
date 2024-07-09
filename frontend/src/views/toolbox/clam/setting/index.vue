<template>
    <div v-loading="loading">
        <LayoutContent>
            <template #app>
                <ClamStatus v-model:loading="loading" />
            </template>
            <template #title>
                <back-button name="Clam" header="ClamAV">
                    <template #buttons>
                        <el-button type="primary" :plain="activeName !== 'clamd'" @click="search('clamd')">
                            {{ $t('toolbox.clam.clamConf') }}
                        </el-button>
                        <el-button type="primary" :plain="activeName !== 'freshclam'" @click="search('freshclam')">
                            {{ $t('toolbox.clam.freshClam') }}
                        </el-button>
                        <el-button type="primary" :plain="activeName !== 'clamd-log'" @click="search('clamd-log')">
                            {{ $t('toolbox.clam.clamLog') }}
                        </el-button>
                        <el-button
                            type="primary"
                            :plain="activeName !== 'freshclam-log'"
                            @click="search('freshclam-log')"
                        >
                            {{ $t('toolbox.clam.freshClamLog') }}
                        </el-button>
                    </template>
                </back-button>
            </template>

            <template #main>
                <div>
                    <el-select v-if="!canUpdate()" style="width: 20%" @change="search()" v-model.number="tail">
                        <template #prefix>{{ $t('toolbox.clam.tail') }}</template>
                        <el-option :value="0" :label="$t('commons.table.all')" />
                        <el-option :value="10" :label="10" />
                        <el-option :value="100" :label="100" />
                        <el-option :value="200" :label="200" />
                        <el-option :value="500" :label="500" />
                        <el-option :value="1000" :label="1000" />
                    </el-select>
                    <codemirror
                        :autofocus="true"
                        :placeholder="$t('commons.msg.noneData')"
                        :indent-with-tab="true"
                        :tabSize="4"
                        :style="{ height: `calc(100vh - ${loadHeight()})`, 'margin-top': '10px' }"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        @ready="handleReady"
                        :extensions="extensions"
                        v-model="content"
                        :disabled="!canUpdate()"
                    />
                    <el-button type="primary" style="margin-top: 10px" v-if="canUpdate()" @click="onSave">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </template>
        </LayoutContent>

        <ConfirmDialog ref="confirmRef" @confirm="onSubmit"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { nextTick, onMounted, ref, shallowRef } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import ClamStatus from '@/views/toolbox/clam/status/index.vue';
import { searchClamFile, updateClamFile } from '@/api/modules/toolbox';
import { oneDark } from '@codemirror/theme-one-dark';
import { GlobalStore } from '@/store';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
const globalStore = GlobalStore();

const loading = ref(false);
const extensions = [javascript(), oneDark];
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const activeName = ref('clamd');
const tail = ref(200);
const content = ref();
const confirmRef = ref();

const loadHeight = () => {
    let height = globalStore.openMenuTabs ? '405px' : '375px';
    if (!canUpdate()) {
        height = globalStore.openMenuTabs ? '383px' : '353px';
    }
    return height;
};

const canUpdate = () => {
    return activeName.value.indexOf('-log') === -1;
};

const search = async (itemName?: string) => {
    if (itemName) {
        tail.value = itemName.indexOf('-log') === -1 ? 0 : 200;
        activeName.value = itemName;
    }
    loading.value = true;
    await searchClamFile(activeName.value, tail.value + '')
        .then((res) => {
            loading.value = false;
            content.value = res.data;
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

const onSave = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmRef.value!.acceptParams(params);
};

const onSubmit = async () => {
    loading.value = true;
    await updateClamFile(activeName.value, content.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search(activeName.value);
});
</script>
