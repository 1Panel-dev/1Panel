<template>
    <div>
        <el-form-item :label="$t('website.rewriteMode')">
            <el-select v-model="req.name" filterable @change="getRewriteConfig(req.name)" class="p-w-200">
                <el-option :label="$t('website.current')" :value="'current'"></el-option>
                <el-option
                    v-for="(rewrite, index) in Rewrites"
                    :key="index"
                    :label="rewrite"
                    :value="rewrite"
                ></el-option>
            </el-select>
        </el-form-item>
        <el-text type="warning">{{ $t('website.rewriteHelper2') }}</el-text>
        <Codemirror
            ref="codeRef"
            v-loading="loading"
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
        <div class="mt-2">
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
import { computed, nextTick, onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { oneDark } from '@codemirror/theme-one-dark';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { GetWebsite, GetRewriteConfig, UpdateRewriteConfig } from '@/api/modules/website';
import { Rewrites } from '@/global/mimetype';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const loading = ref(false);
const content = ref(' ');
const extensions = [StreamLanguage.define(nginx), oneDark];
const codeRef = ref();

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
    content: 'd',
    name: '',
});

const getRewriteConfig = async (rewrite: string) => {
    loading.value = true;
    req.name = rewrite;
    req.websiteID = id.value;
    try {
        const res = await GetRewriteConfig(req);
        content.value = res.data.content;
        if (res.data.content == '') {
            content.value = ' ';
        }

        setCursorPosition();
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const setCursorPosition = () => {
    nextTick(() => {
        const codeMirrorInstance = codeRef.value?.codemirror;
        codeMirrorInstance?.setCursor(0, 0);
    });
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
