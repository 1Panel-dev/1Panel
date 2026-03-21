<template>
    <DrawerPro v-model="open" :header="$t('website.defaultServer')" size="normal" @close="handleClose">
        <el-form @submit.prevent label-position="top" v-loading="loading">
            <el-form-item :label="$t('website.defaultServer')">
                <el-select v-model="defaultMode" @change="onModeChange">
                    <el-option value="none" :label="$t('website.noDefaultServer')" />
                    <el-option value="defaultPage" :label="$t('website.defaultPage')" />
                    <el-option
                        v-for="(website, key) in websites"
                        :key="key"
                        :value="'site:' + website.id"
                        :label="website.primaryDomain"
                    />
                </el-select>
            </el-form-item>

            <template v-if="defaultMode === 'defaultPage'">
                <el-form-item :label="$t('website.defaultPageType')">
                    <el-select v-model="pageType">
                        <el-option value="html" :label="$t('website.customHtml')" />
                        <el-option value="noResponse" :label="$t('website.noResponse')" />
                        <el-option value="redirect" :label="$t('website.redirectTo')" />
                    </el-select>
                </el-form-item>

                <el-form-item v-if="pageType === 'redirect'" :label="$t('website.redirectUrl')">
                    <el-input v-model="redirectUrl" placeholder="https://example.com" />
                </el-form-item>

                <template v-if="pageType === 'html'">
                    <el-form-item :label="$t('website.defaultHtml')">
                        <el-select v-model="htmlType" @change="loadHtml" class="w-full">
                            <el-option value="404" :label="$t('website.website404')" />
                            <el-option value="index" :label="$t('website.indexHtml')" />
                            <el-option value="stop" :label="$t('website.stopHtml')" />
                        </el-select>
                    </el-form-item>
                    <div ref="htmlRef" class="default-html"></div>
                </template>
            </template>
        </el-form>

        <el-alert :closable="false" class="mt-4">
            <template #default>
                <span class="whitespace-pre-line">{{ $t('website.defaultServerHelper') }}</span>
                <div v-if="pageType === 'noResponse'" class="mt-2">
                    <span>{{ $t('website.noResponseHelper') }}</span>
                </div>
            </template>
        </el-alert>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref, nextTick } from 'vue';
import { Website } from '@/api/interface/website';
import { changeDefaultServer, listWebsites, getDefaultHtml, updateDefaultHtml } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { EditorState } from '@codemirror/state';
import { basicSetup, EditorView } from 'codemirror';
import { html } from '@codemirror/lang-html';
import { oneDark } from '@codemirror/theme-one-dark';

let open = ref(false);
let websites = ref<any>([]);
let loading = ref(false);
let defaultMode = ref('none');
let pageType = ref('html');
let htmlType = ref('404');
let redirectUrl = ref('');

const htmlRef = ref();
const view = ref();
const content = ref('');

const acceptParams = () => {
    defaultMode.value = 'none';
    pageType.value = 'html';
    htmlType.value = '404';
    redirectUrl.value = '';
    get();
    open.value = true;
};

const handleClose = () => {
    if (view.value) {
        view.value.destroy();
        view.value = null;
    }
    open.value = false;
};

const get = async () => {
    const res = await listWebsites();
    websites.value = res.data;
    defaultMode.value = 'none';
    websites.value.forEach((website: Website.WebsiteDTO) => {
        if (website.defaultServer) {
            defaultMode.value = 'site:' + website.id;
        }
    });
    if (defaultMode.value === 'none') {
        // Check if default page config exists (default_server on 00.default.conf)
        // Keep as 'none' - user can switch to defaultPage if desired
    }
};

const onModeChange = () => {
    if (defaultMode.value === 'defaultPage') {
        nextTick(() => {
            if (pageType.value === 'html') {
                loadHtml();
            }
        });
    }
};

const loadHtml = async () => {
    try {
        const res = await getDefaultHtml(htmlType.value);
        content.value = res.data.content;
        initEditor();
    } catch (e) {
        content.value = '';
    }
};

const initEditor = () => {
    if (view.value) {
        view.value.destroy();
        view.value = null;
    }
    nextTick(() => {
        if (!htmlRef.value) return;
        const startState = EditorState.create({
            doc: content.value,
            extensions: [basicSetup, oneDark, html()],
        });
        view.value = new EditorView({
            state: startState,
            parent: htmlRef.value,
        });
    });
};

const submit = async () => {
    loading.value = true;
    try {
        if (defaultMode.value === 'none') {
            await changeDefaultServer({ id: 0 });
        } else if (defaultMode.value === 'defaultPage') {
            // Set no specific website as default (use default config)
            await changeDefaultServer({ id: 0 });

            if (pageType.value === 'html' && view.value) {
                const htmlContent = view.value.state.doc.toString();
                await updateDefaultHtml({ type: htmlType.value, content: htmlContent, sync: false });
            } else if (pageType.value === 'noResponse') {
                // Write a 444 return config as the default page
                const noResponseHtml = '<html><head><meta http-equiv="refresh" content="0;url=about:blank"></head><body></body></html>';
                await updateDefaultHtml({ type: 'index', content: noResponseHtml, sync: false });
            } else if (pageType.value === 'redirect' && redirectUrl.value) {
                const redirectHtml = `<html><head><meta http-equiv="refresh" content="0;url=${redirectUrl.value}"></head><body>Redirecting...</body></html>`;
                await updateDefaultHtml({ type: 'index', content: redirectHtml, sync: false });
            }
        } else if (defaultMode.value.startsWith('site:')) {
            const siteId = parseInt(defaultMode.value.replace('site:', ''));
            await changeDefaultServer({ id: siteId });
        }
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        handleClose();
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
    margin-top: 10px;
}
</style>
