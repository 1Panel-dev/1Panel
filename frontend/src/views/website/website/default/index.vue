<template>
    <DrawerPro v-model="open" :header="$t('website.advancedSettings')" size="normal" @close="handleClose">
        <template #content>
            <el-tabs v-model="activeTab" tab-position="left" @tab-change="changeTab">
                <el-tab-pane :label="$t('website.defaultServer')" name="site">
                    <el-form @submit.prevent label-position="top" v-loading="loading">
                        <el-form-item :label="$t('website.defaultServer')">
                            <el-select v-model="defaultId">
                                <el-option :value="0" :key="-1" :label="$t('website.noDefaultServer')"></el-option>
                                <el-option
                                    v-for="(website, key) in websites"
                                    :key="key"
                                    :value="website.id"
                                    :label="website.primaryDomain"
                                ></el-option>
                            </el-select>
                        </el-form-item>
                    </el-form>
                    <el-alert :closable="false">
                        <template #default>
                            <span class="whitespace-pre-line">{{ $t('website.defaultServerHelper') }}</span>
                        </template>
                    </el-alert>
                </el-tab-pane>
                <el-tab-pane :label="$t('website.defaultHtml')" name="html">
                    <el-select v-model="req.type" class="w-full" @change="getHtml()" v-loading="loading">
                        <el-option :value="'404'" :label="$t('website.website404')"></el-option>
                        <el-option :value="'domain404'" :label="$t('website.domain404')"></el-option>
                        <el-option :value="'index'" :label="$t('website.indexHtml')"></el-option>
                        <el-option :value="'php'" :label="$t('website.indexPHP')"></el-option>
                        <el-option :value="'stop'" :label="$t('website.stopHtml')"></el-option>
                    </el-select>
                    <div class="mt-1.5">
                        <el-text v-if="req.type == '404'" type="info">
                            {{ $t('website.website404Helper') }}
                        </el-text>
                    </div>
                    <div class="mt-1.5">
                        <el-checkbox v-model="req.sync">{{ $t('website.syncHtmlHelper') }}</el-checkbox>
                    </div>
                    <div ref="htmlRef" class="default-html"></div>
                </el-tab-pane>
            </el-tabs>
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission type="primary" @click="submit()" :disabled="loading">
                    {{ $t(activeTab === 'site' ? 'commons.button.confirm' : 'commons.button.save') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { changeDefaultServer, getDefaultHtml, listWebsites, updateDefaultHtml } from '@/api/modules/website';
import i18n from '@/lang';
import { nextTick, reactive, ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { EditorState } from '@codemirror/state';
import { basicSetup, EditorView } from 'codemirror';
import { html } from '@codemirror/lang-html';
import { php } from '@codemirror/lang-php';
import { oneDark } from '@codemirror/theme-one-dark';

let open = ref(false);
let websites = ref<any>();
let defaultId = ref(-1);
let loading = ref(false);
const activeTab = ref('site');
const content = ref('');
const view = ref();
const htmlRef = ref();
const htmlLoaded = ref(false);
const req = reactive({
    type: '404',
    sync: false,
});

const acceptParams = () => {
    activeTab.value = 'site';
    defaultId.value = 0;
    req.type = '404';
    htmlLoaded.value = false;
    destroyEditor();
    get();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
    destroyEditor();
};

const get = async () => {
    const res = await listWebsites();
    websites.value = res.data;
    websites.value.forEach((website: Website.WebsiteDTO) => {
        if (website.defaultServer) {
            defaultId.value = website.id;
        }
    });
};

const changeTab = async () => {
    if (activeTab.value === 'html' && !htmlLoaded.value) {
        await getHtml();
    }
};

const getHtml = async () => {
    loading.value = true;
    try {
        const res = await getDefaultHtml(req.type);
        content.value = res.data.content;
        htmlLoaded.value = true;
        await nextTick();
        initEditor();
    } finally {
        loading.value = false;
    }
};

const destroyEditor = () => {
    if (view.value) {
        view.value.destroy();
        view.value = undefined;
    }
};

const initEditor = () => {
    destroyEditor();
    let extensions = [basicSetup, oneDark];
    if (req.type === 'php') {
        extensions.push(php());
    } else {
        extensions.push(html());
    }
    const startState = EditorState.create({
        doc: content.value,
        extensions: extensions,
    });
    if (htmlRef.value) {
        view.value = new EditorView({
            state: startState,
            parent: htmlRef.value,
        });
    }
};

const submit = () => {
    if (activeTab.value === 'html') {
        submitHtml();
        return;
    }
    submitDefaultServer();
};

const submitDefaultServer = () => {
    loading.value = true;
    changeDefaultServer({ id: defaultId.value })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            handleClose();
        })
        .finally(() => {
            loading.value = false;
        });
};

const submitHtml = async () => {
    loading.value = true;
    try {
        const content = view.value.state.doc.toString();
        await updateDefaultHtml({ type: req.type, content: content, sync: req.sync });
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } finally {
        loading.value = false;
    }
};
defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.default-html {
    width: 100%;
    min-height: 300px;
}
</style>
