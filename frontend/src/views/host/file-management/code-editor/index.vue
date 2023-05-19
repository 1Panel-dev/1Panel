<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.edit') + ' - ' + fileName"
        :before-close="handleClose"
        destroy-on-close
        width="70%"
        @opened="onOpen"
        :top="'5vh'"
    >
        <el-form :inline="true" :model="config">
            <el-form-item :label="$t('file.theme')">
                <el-select v-model="config.theme" @change="changeTheme()">
                    <el-option v-for="item in themes" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('file.language')">
                <el-select v-model="config.language" @change="changeLanguage()">
                    <el-option v-for="lang in Languages" :key="lang.label" :value="lang.label" :label="lang.label" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('file.eol')">
                <el-select v-model="config.eol" @change="changeEOL()">
                    <el-option v-for="eol in eols" :key="eol.label" :value="eol.value" :label="eol.label" />
                </el-select>
            </el-form-item>
        </el-form>
        <div v-loading="loading">
            <div id="codeBox" style="height: 55vh"></div>
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
import { nextTick, onBeforeUnmount, reactive, ref } from 'vue';
import { Languages } from '@/global/mimetype';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';

let editor: monaco.editor.IStandaloneCodeEditor | undefined;

self.MonacoEnvironment = {
    getWorker(workerId, label) {
        if (label === 'json') {
            return new jsonWorker();
        }
        if (label === 'css' || label === 'scss' || label === 'less') {
            return new cssWorker();
        }
        if (label === 'html' || label === 'handlebars' || label === 'razor') {
            return new htmlWorker();
        }
        if (['typescript', 'javascript'].includes(label)) {
            return new tsWorker();
        }
        return new EditorWorker();
    },
};

interface EditProps {
    language: string;
    content: string;
    path: string;
    name: string;
}

interface EditorConfig {
    theme: string;
    language: string;
    eol: number;
}

const open = ref(false);
const loading = ref(false);
const fileName = ref('');

const config = reactive<EditorConfig>({
    theme: 'vs-dark',
    language: 'plaintext',
    eol: monaco.editor.EndOfLineSequence.LF,
});

const eols = [
    {
        label: 'LF (Linux)',
        value: monaco.editor.EndOfLineSequence.LF,
    },
    {
        label: 'CRLF (Windows)',
        value: monaco.editor.EndOfLineSequence.CRLF,
    },
];

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
    em('close', open.value);
};
const changeLanguage = () => {
    monaco.editor.setModelLanguage(editor.getModel(), config.language);
};

const changeTheme = () => {
    monaco.editor.setTheme(config.theme);
};

const changeEOL = () => {
    editor.getModel().pushEOL(config.eol);
};

const initEditor = () => {
    if (editor) {
        editor.dispose();
    }
    nextTick(() => {
        const codeBox = document.getElementById('codeBox');
        editor = monaco.editor.create(codeBox as HTMLElement, {
            theme: config.theme,
            value: form.value.content,
            readOnly: false,
            automaticLayout: true,
            language: config.language,
            folding: true,
            roundedSelection: false,
            overviewRulerBorder: false,
        });

        editor.onDidChangeModelContent(() => {
            if (editor) {
                form.value.content = editor.getValue();
            }
        });

        // After onDidChangeModelContent
        editor.getModel().pushEOL(config.eol);

        editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, quickSave);
    });
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
        if (closePage) {
            handleClose();
        }
    });
};

const acceptParams = (props: EditProps) => {
    form.value.content = props.content;
    form.value.path = props.path;
    config.language = props.language;
    fileName.value = props.name;
    // TODO Now,1panel only support liunux,so we can use LF.
    // better,We should rely on the actual line feed character of the file returned from the background
    config.eol = monaco.editor.EndOfLineSequence.LF;
    open.value = true;
};

const onOpen = () => {
    initEditor();
};

onBeforeUnmount(() => {
    if (editor) {
        editor.dispose();
    }
});

defineExpose({ acceptParams });
</script>

<style>
.dialog-top {
    top: 0;
}
</style>
