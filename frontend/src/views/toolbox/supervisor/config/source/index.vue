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
            :mode="'text/x-ini'"
            v-model="content"
        />
        <div style="margin-top: 10px">
            <el-button type="primary" @click="submit()" :disabled="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { StreamLanguage } from '@codemirror/language';
import { properties } from '@codemirror/legacy-modes/mode/properties';
import { oneDark } from '@codemirror/theme-one-dark';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { OperateSupervisorConfig } from '@/api/modules/host-tool';

const extensions = [StreamLanguage.define(properties), oneDark];

let data = ref();
let content = ref('');
let loading = ref(false);

const submit = () => {
    loading.value = true;
    OperateSupervisorConfig({ type: 'supervisord', operate: 'set', content: content.value })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            getConfig();
        })
        .finally(() => {
            loading.value = false;
        });
};

const getConfig = async () => {
    const res = await OperateSupervisorConfig({ type: 'supervisord', operate: 'get' });
    data.value = res.data;
    content.value = data.value.content;
};

onMounted(() => {
    getConfig();
});
</script>
