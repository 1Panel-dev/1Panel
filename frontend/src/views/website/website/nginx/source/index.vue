<template>
    <div v-loading="loading">
        <codemirror
            :autofocus="true"
            :placeholder="$t('commons.msg.noneData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="width: 100%; height: calc(100vh - 375px)"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            :mode="'text/x-nginx-conf'"
            v-model="content"
        />
        <div style="margin-top: 10px">
            <el-button @click="getDefaultConfig()" :disabled="loading">
                {{ $t('app.defaultConfig') }}
            </el-button>
            <el-button type="primary" @click="submit()" :disabled="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </div>
        <el-row>
            <el-col :span="4">
                <el-alert
                    v-if="useOld"
                    style="margin-top: 10px"
                    :title="$t('app.defaultConfigHelper')"
                    type="info"
                    :closable="false"
                ></el-alert>
            </el-col>
        </el-row>
    </div>
</template>
<script lang="ts" setup>
import { GetNginx, UpdateNginxConfigFile } from '@/api/modules/nginx';
import { onMounted, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { oneDark } from '@codemirror/theme-one-dark';
import i18n from '@/lang';
import { GetAppDefaultConfig } from '@/api/modules/app';
import { MsgSuccess } from '@/utils/message';

const extensions = [StreamLanguage.define(nginx), oneDark];

let content = ref('');
let loading = ref(false);
let useOld = ref(false);

const submit = () => {
    loading.value = true;
    UpdateNginxConfigFile({
        content: content.value,
        backup: useOld.value,
    })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            getNginx();
        })
        .finally(() => {
            loading.value = false;
        });
};

const getNginx = async () => {
    try {
        const res = await GetNginx();
        content.value = res.data.content;
        useOld.value = false;
    } catch (error) {}
};

const getDefaultConfig = async () => {
    loading.value = true;
    try {
        const res = await GetAppDefaultConfig('openresty', '');
        content.value = res.data;
        useOld.value = true;
    } catch (error) {}
    loading.value = false;
};

onMounted(() => {
    getNginx();
});
</script>
