<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.edit')"
        :before-close="handleClose"
        destroy-on-close
        width="70%"
        @opened="onOpen"
    >
        <el-form :inline="true" :model="config">
            <el-form-item :label="$t('file.theme')">
                <el-select v-model="config.theme" @change="initEditor()">
                    <el-option v-for="item in themes" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('file.language')">
                <el-select v-model="config.language" @change="initEditor()">
                    <el-option v-for="lang in Languages" :key="lang.label" :value="lang.value" :label="lang.label" />
                </el-select>
            </el-form-item>
        </el-form>
        <div class="coder-editor" v-loading="loading">
            <div id="codeBox" style="height: 60vh"></div>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="saveContent(true)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { SaveFileContent } from '@/api/modules/files';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import * as monaco from 'monaco-editor';
import { reactive, ref } from 'vue';
import { Languages } from '@/global/mimetype';

let editor: monaco.editor.IStandaloneCodeEditor | undefined;

interface EditProps {
    language: string;
    content: string;
    path: string;
    name: string;
}

interface EditorConfig {
    theme: string;
    language: string;
}

let open = ref(false);
let loading = ref(false);

let config = reactive<EditorConfig>({
    theme: 'vs-dark',
    language: 'json',
});

const themes = [
    {
        label: 'Visual Studio',
        value: 'vs',
    },
    {
        label: 'Visual Studio Dark',
        value: 'vs-dark',
    },
    {
        label: 'High Contrast Dark',
        value: 'hc-black',
    },
];

let form = ref({
    content: '',
    path: '',
});

const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    if (editor) {
        editor.dispose();
    }
    em('close', false);
};

const initEditor = () => {
    if (editor) {
        editor.dispose();
    }
    const codeBox = document.getElementById('codeBox');
    editor = monaco.editor.create(codeBox as HTMLElement, {
        theme: config.theme, //官方自带三种主题vs, hc-black, or vs-dark
        value: form.value.content,
        readOnly: false,
        automaticLayout: true,
        language: config.language,
        folding: true, //代码折叠
        roundedSelection: false, // 右侧不显示编辑器预览框
    });

    editor.onDidChangeModelContent(() => {
        if (editor) {
            form.value.content = editor.getValue();
        }
    });

    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, quickSave);
};

const quickSave = () => {
    saveContent(false);
};

const saveContent = (closePage: boolean) => {
    loading.value = true;
    SaveFileContent(form.value).finally(() => {
        loading.value = false;
        open.value = !closePage;
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    });
};

const acceptParams = (props: EditProps) => {
    form.value.content = props.content;
    form.value.path = props.path;
    config.language = props.language;
    open.value = true;
};

const onOpen = () => {
    initEditor();
};

defineExpose({ acceptParams });
</script>

<style lang="scss">
.coder-editor {
    margin-top: 10px;
}
</style>
