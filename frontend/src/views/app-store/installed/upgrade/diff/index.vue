<template>
    <el-dialog
        v-model="open"
        :title="$t('app.composeDiff')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="90%"
    >
        <div>
            <el-text type="warning">{{ $t('app.diffHelper') }}</el-text>
            <div ref="container" class="compose-diff"></div>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="success" @click="confirm(true)">
                    {{ $t('app.useNew') }}
                </el-button>
                <el-button type="primary" @click="confirm(false)">
                    {{ $t('app.useDefault') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import * as monaco from 'monaco-editor';

const open = ref(false);
const newContent = ref('');
const oldContent = ref('');
const em = defineEmits(['confirm']);

let originalModel = null;
let modifiedModel = null;
let editor: monaco.editor.IStandaloneDiffEditor = null;

const container = ref();

const handleClose = () => {
    open.value = false;
    if (editor) {
        editor.dispose();
    }
};

const acceptParams = (oldCompose: string, newCompose: string) => {
    oldContent.value = oldCompose;
    newContent.value = newCompose;
    open.value = true;
    initEditor();
};

const confirm = (useEditor: boolean) => {
    let content = '';
    if (useEditor) {
        content = editor.getModifiedEditor().getValue();
    } else {
        content = '';
    }
    em('confirm', content);
    handleClose();
};

const initEditor = () => {
    nextTick(() => {
        originalModel = monaco.editor.createModel(oldContent.value, 'yaml');
        modifiedModel = monaco.editor.createModel(newContent.value, 'yaml');

        editor = monaco.editor.createDiffEditor(container.value, {
            theme: 'vs-dark',
            readOnly: false,
            automaticLayout: true,
            folding: true,
            roundedSelection: false,
            overviewRulerBorder: false,
        });

        editor.setModel({
            original: originalModel,
            modified: modifiedModel,
        });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.compose-diff {
    width: 100%;
    height: calc(100vh - 350px);
}
</style>
