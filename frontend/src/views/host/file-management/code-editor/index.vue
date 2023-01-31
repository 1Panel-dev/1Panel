<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.edit')"
        :before-close="handleClose"
        destroy-on-close
        width="70%"
        draggable
        @opened="onOpen"
    >
        <div>
            <div v-loading="loading">
                <div id="codeBox" style="height: 60vh"></div>
            </div>
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
import { ElMessage } from 'element-plus';
import * as monaco from 'monaco-editor';
import { ref } from 'vue';

let editor: monaco.editor.IStandaloneCodeEditor | undefined;

interface EditProps {
    language: string;
    content: string;
    path: string;
    name: string;
}

let open = ref(false);
let loading = ref(false);
let language = ref('json');

let form = ref({
    content: '',
    path: '',
});

const em = defineEmits(['close']);

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
        value: form.value.content,
        readOnly: false,
        automaticLayout: true,
        language: language.value,
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
        ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
    });
};

const acceptParams = (props: EditProps) => {
    form.value.content = props.content;
    form.value.path = props.path;
    open.value = true;
};

const onOpen = () => {
    initEditor();
};

defineExpose({ acceptParams });
</script>
