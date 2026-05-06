<template>
    <DialogPro v-model="open" :title="$t('app.composeDiff')" @close="handleClose" size="w-90">
        <div>
            <el-text type="warning">{{ $t('app.diffHelper') }}</el-text>
            <div ref="container" class="compose-diff"></div>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="confirm(true)">
                    {{ $t('app.useNew') }}
                </el-button>
                <el-button @click="confirm(false)">
                    {{ $t('app.useDefault') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { loadMonacoLanguageSupport, setupMonacoEnvironment } from '@/utils/monaco';

type MonacoEditorApi = typeof import('monaco-editor/esm/vs/editor/editor.api');

const open = ref(false);
const newContent = ref('');
const oldContent = ref('');
const em = defineEmits(['confirm']);

let monaco: MonacoEditorApi | null = null;
let originalModel: MonacoEditorApi['editor']['ITextModel'] | null = null;
let modifiedModel: MonacoEditorApi['editor']['ITextModel'] | null = null;
let editor: MonacoEditorApi['editor']['IStandaloneDiffEditor'] | null = null;

const container = ref();

const ensureMonaco = async () => {
    if (monaco) {
        return monaco;
    }
    setupMonacoEnvironment();
    const [monacoModule] = await Promise.all([
        import('monaco-editor/esm/vs/editor/editor.api'),
        loadMonacoLanguageSupport(),
    ]);
    monaco = monacoModule;
    return monaco;
};

const disposeModels = () => {
    originalModel?.dispose();
    modifiedModel?.dispose();
    originalModel = null;
    modifiedModel = null;
};

const handleClose = () => {
    open.value = false;
    if (editor) {
        editor.dispose();
        editor = null;
    }
    disposeModels();
};

const acceptParams = async (oldCompose: string, newCompose: string) => {
    oldContent.value = oldCompose;
    newContent.value = newCompose;
    open.value = true;
    await initEditor();
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

const initEditor = async () => {
    const monacoApi = await ensureMonaco();
    await nextTick();

    if (!container.value) {
        return;
    }

    if (editor) {
        editor.dispose();
        editor = null;
    }
    disposeModels();

    originalModel = monacoApi.editor.createModel(oldContent.value, 'yaml');
    modifiedModel = monacoApi.editor.createModel(newContent.value, 'yaml');

    editor = monacoApi.editor.createDiffEditor(container.value, {
        theme: 'vs-dark',
        readOnly: false,
        automaticLayout: true,
        folding: true,
        glyphMargin: true,
        roundedSelection: false,
        overviewRulerBorder: false,
        renderMarginRevertIcon: true,
        renderGutterMenu: false,
    });

    editor.setModel({
        original: originalModel,
        modified: modifiedModel,
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
