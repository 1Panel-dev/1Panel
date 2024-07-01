<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="40%">
        <template #header>
            <DrawerHeader :header="$t('website.defaultHtml')" :back="handleClose"></DrawerHeader>
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-select v-model="type" class="w-full" @change="get()">
                    <el-option :value="'404'" :label="$t('website.website404')"></el-option>
                    <el-option :value="'domain404'" :label="$t('website.domain404')"></el-option>
                    <el-option :value="'index'" :label="$t('website.indexHtml')"></el-option>
                    <el-option :value="'php'" :label="$t('website.indexPHP')"></el-option>
                    <el-option :value="'stop'" :label="$t('website.stopHtml')"></el-option>
                </el-select>
                <div class="mt-1.5">
                    <el-text v-if="type == '404'" type="info">
                        {{ $t('website.website404Helper') }}
                    </el-text>
                </div>
                <div ref="htmlRef" class="default-html"></div>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.save') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { UpdateDefaultHtml, GetDefaultHtml } from '@/api/modules/website';
import i18n from '@/lang';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { EditorState } from '@codemirror/state';
import { basicSetup, EditorView } from 'codemirror';
import { html } from '@codemirror/lang-html';
import { php } from '@codemirror/lang-php';
import { oneDark } from '@codemirror/theme-one-dark';

let open = ref(false);
let loading = ref(false);
const content = ref('');
const type = ref('404');
const view = ref();
const htmlRef = ref();

const acceptParams = () => {
    type.value = '404';
    get();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
};

const get = async () => {
    const res = await GetDefaultHtml(type.value);
    content.value = res.data.content;
    initEditor();
};

const initEditor = () => {
    if (view.value) {
        view.value.destroy();
    }
    let extensions = [basicSetup, oneDark];
    if (type.value === 'php') {
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

const submit = async () => {
    loading.value = true;
    try {
        const content = view.value.state.doc.toString();
        await UpdateDefaultHtml({ type: type.value, content: content });
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } catch (error) {
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
