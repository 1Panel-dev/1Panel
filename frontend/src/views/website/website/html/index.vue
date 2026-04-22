<template>
    <DrawerPro v-model="open" :header="$t('website.defaultHtml')" @close="handleClose" size="normal">
        <el-select v-model="req.type" class="w-full" @change="get()" v-loading="loading">
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

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.save') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { updateDefaultHtml, getDefaultHtml } from '@/api/modules/website';
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
const view = ref();
const htmlRef = ref();
const req = reactive({
    type: '404',
    sync: false,
});

const acceptParams = () => {
    req.type = '404';
    get();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
};

const get = async () => {
    const res = await getDefaultHtml(req.type);
    content.value = res.data.content;
    initEditor();
};

const initEditor = () => {
    if (view.value) {
        view.value.destroy();
    }
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

const submit = async () => {
    loading.value = true;
    try {
        const content = view.value.state.doc.toString();
        await updateDefaultHtml({ type: req.type, content: content, sync: req.sync });
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
}
</style>
