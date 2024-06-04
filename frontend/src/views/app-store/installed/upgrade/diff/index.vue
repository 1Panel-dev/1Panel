<template>
    <el-dialog
        v-model="open"
        :title="$t('app.composeDiff')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="60%"
    >
        <el-row :gutter="10">
            <el-col :span="11" :offset="1" class="mt-2">
                <el-text type="info">{{ $t('app.oldVersion') }}</el-text>
                <codemirror
                    placeholder=""
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="width: 100%; height: calc(100vh - 500px)"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    v-model="oldContent"
                    :readOnly="true"
                />
            </el-col>
            <el-col :span="11" class="mt-2">
                <el-text type="success">{{ $t('app.newVersion') }}</el-text>
                <el-text type="warning" class="!ml-5">编辑之后点击使用自定义版本保存</el-text>
                <codemirror
                    :autofocus="true"
                    placeholder=""
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="width: 100%; height: calc(100vh - 500px)"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    v-model="newContent"
                />
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="success" @click="confirm(newContent)">
                    {{ $t('app.useNew') }}
                </el-button>
                <el-button type="primary" @click="confirm('')">
                    {{ $t('app.useDefault') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
const extensions = [javascript(), oneDark];

const open = ref(false);
const newContent = ref('');
const oldContent = ref('');
const em = defineEmits(['confirm']);

const handleClose = () => {
    open.value = false;
};

const acceptParams = (oldCompose: string, newCompose: string) => {
    oldContent.value = oldCompose;
    newContent.value = newCompose;
    open.value = true;
};

const confirm = (content: string) => {
    em('confirm', content);
    handleClose();
};

defineExpose({
    acceptParams,
});
</script>
