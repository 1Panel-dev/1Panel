<template>
    <el-dialog v-model="composeVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="70%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('commons.button.edit') }}</span>
            </div>
        </template>
        <div v-loading="loading">
            <codemirror
                ref="mymirror"
                :autofocus="true"
                placeholder="None data"
                :indent-with-tab="true"
                :tabSize="4"
                style="max-height: 500px"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="content"
                :readOnly="true"
            />
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="composeVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmitEdit()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { composeUpdate } from '@/api/modules/container';
import i18n from '@/lang';

const loading = ref(false);
const composeVisiable = ref(false);
const extensions = [javascript(), oneDark];
const path = ref();
const content = ref();
const name = ref();

const onSubmitEdit = async () => {
    const param = {
        name: name.value,
        path: path.value,
        content: content.value,
    };
    loading.value = true;
    await composeUpdate(param)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            composeVisiable.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

interface DialogProps {
    name: string;
    path: string;
    content: string;
}

const acceptParams = (props: DialogProps): void => {
    composeVisiable.value = true;
    path.value = props.path;
    name.value = props.name;
    content.value = props.content;
};

defineExpose({
    acceptParams,
});
</script>
