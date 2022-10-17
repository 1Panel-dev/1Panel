<template>
    <el-dialog v-model="templateVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>{{ title }}{{ $t('container.composeTemplate') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="dialogData.rowData" :rules="rules" label-width="80px">
            <el-form-item :label="$t('container.name')" prop="name">
                <el-input :disabled="dialogData.title === 'edit'" v-model="dialogData.rowData!.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.description')">
                <el-input v-model="dialogData.rowData!.description"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.from')">
                <el-radio-group v-model="dialogData.rowData!.from">
                    <el-radio label="edit">{{ $t('container.edit') }}</el-radio>
                    <el-radio label="path">{{ $t('container.pathSelect') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="dialogData.rowData!.from === 'path'" prop="path">
                <el-input
                    clearable
                    :placeholder="$t('commons.example') + '/tmp/docker-compose.yml'"
                    v-model="dialogData.rowData!.path"
                >
                    <template #append>
                        <FileList @choose="loadDir" :dir="false"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item v-if="dialogData.rowData!.from === 'edit'">
                <codemirror
                    :autofocus="true"
                    placeholder="None data"
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="max-height: 500px; width: 100%"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    v-model="dialogData.rowData!.content"
                    :readOnly="true"
                />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="templateVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import FileList from '@/components/file-list/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { Container } from '@/api/interface/container';
import { createComposeTemplate, updateComposeTemplate } from '@/api/modules/container';

interface DialogProps {
    title: string;
    rowData?: Container.TemplateInfo;
    getTableList?: () => Promise<any>;
}
const extensions = [javascript(), oneDark];
const title = ref<string>('');
const templateVisiable = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    templateVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const varifyPath = (rule: any, value: any, callback: any) => {
    console.log(value, value.indexOf('docker-compose.yml'));
    if (value.indexOf('docker-compose.yml') === -1) {
        callback(new Error(i18n.global.t('commons.rule.selectHelper', ['docker-compose.yml'])));
    }
    callback();
};
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    path: [Rules.requiredInput, { validator: varifyPath, trigger: 'change', required: true }],
    content: [Rules.requiredInput],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

function restForm() {
    if (formRef.value) {
        formRef.value.resetFields();
    }
}
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (dialogData.value.title === 'create') {
            await createComposeTemplate(dialogData.value.rowData!);
        }
        if (dialogData.value.title === 'edit') {
            await updateComposeTemplate(dialogData.value.rowData!);
        }
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        restForm();
        emit('search');
        templateVisiable.value = false;
    });
};

const loadDir = async (path: string) => {
    dialogData.value.rowData!.path = path;
};

defineExpose({
    acceptParams,
});
</script>
