<template>
    <div v-loading="loading">
        <el-form-item :label="$t('website.rewriteMode')">
            <el-select v-model="req.name" filterable @change="getRewriteConfig(req.name)">
                <el-option :label="$t('website.current')" :value="'current'"></el-option>
                <el-option
                    v-for="(rewrite, index) in Rewrites"
                    :key="index"
                    :label="rewrite"
                    :value="rewrite"
                ></el-option>
            </el-select>
        </el-form-item>

        <codemirror
            :autofocus="true"
            placeholder=""
            :indent-with-tab="true"
            :tabSize="4"
            style="margin-top: 10px; height: 300px"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="content"
        />
        <div style="margin-top: 10px">
            <el-form-item>
                <el-alert :title="$t('website.rewriteHelper')" type="info" :closable="false" />
            </el-form-item>
            <el-button type="primary" @click="submit()">
                {{ $t('nginx.saveAndReload') }}
            </el-button>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { oneDark } from '@codemirror/theme-one-dark';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { GetWebsite, GetRewriteConfig, UpdateRewriteConfig } from '@/api/modules/website';
import { Rewrites } from '@/global/mimetype';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const loading = ref(false);
const content = ref('');
const extensions = [StreamLanguage.define(nginx), oneDark];

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

const req = reactive({
    websiteID: id.value,
    name: 'default',
});

const update = reactive({
    websiteID: id.value,
    content: '',
    name: '',
});

const getRewriteConfig = async (rewrite: string) => {
    loading.value = true;
    req.name = rewrite;
    req.websiteID = id.value;
    try {
        const res = await GetRewriteConfig(req);
        content.value = res.data.content;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const submit = async () => {
    update.name = req.name;
    update.websiteID = id.value;
    update.content = content.value;
    loading.value = true;
    try {
        await UpdateRewriteConfig(update);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    GetWebsite(id.value).then((res) => {
        const name = res.data.rewrite == '' ? 'default' : 'current';
        if (name === 'current') {
            req.name = 'current';
        }
        getRewriteConfig(name);
    });
});
</script>
