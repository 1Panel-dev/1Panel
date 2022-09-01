<template>
    <el-dialog v-model="open" :title="'code editor'" @opened="onOpen" :before-close="handleClose">
        <div>
            <div id="codeBox" style="height: 600px"></div>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="saveContent">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import * as monaco from 'monaco-editor';
import { reactive } from 'vue';

let editor: monaco.editor.IStandaloneCodeEditor | undefined;

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    language: {
        type: String,
        default: 'json',
    },
    content: {
        type: String,
        default: '',
    },
});

let data = reactive({
    content: '',
    language: '',
});

const em = defineEmits(['close', 'save']);

const handleClose = () => {
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
        theme: 'vs-dark', //官方自带三种主题vs, hc-black, or vs-dark
        value: data.content,
        readOnly: false,
        automaticLayout: true,
        language: data.language,
        folding: true, //代码折叠
        roundedSelection: false, // 右侧不显示编辑器预览框
    });

    editor.onDidChangeModelContent(() => {
        if (editor) {
            data.content = editor.getValue();
        }
    });
};

const saveContent = () => {
    em('save', data.content);
};

const onOpen = () => {
    data.content = props.content;
    data.language = props.language;
    initEditor();
};
</script>
