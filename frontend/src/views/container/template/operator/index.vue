<template>
    <el-dialog v-model="templateVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>{{ title }}{{ $t('container.composeTemplate') }}</span>
            </div>
        </template>
        <el-form v-loading="loading" ref="formRef" :model="dialogData.rowData" :rules="rules" label-width="80px">
            <el-form-item :label="$t('container.name')" prop="name">
                <el-input :disabled="dialogData.title === 'edit'" v-model="dialogData.rowData!.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.description')">
                <el-input v-model="dialogData.rowData!.description"></el-input>
            </el-form-item>
            <el-form-item>
                <codemirror
                    :autofocus="true"
                    placeholder="#Define or paste the content of your docker-compose file here"
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="max-height: 500px; width: 100%; min-height: 200px"
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
                <el-button :disabled="loading" @click="templateVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { Container } from '@/api/interface/container';
import { createComposeTemplate, updateComposeTemplate } from '@/api/modules/container';

const loading = ref(false);

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

const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    content: [Rules.requiredInput],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await createComposeTemplate(dialogData.value.rowData!)
                .then(() => {
                    loading.value = false;
                    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    templateVisiable.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await updateComposeTemplate(dialogData.value.rowData!)
            .then(() => {
                loading.value = false;
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                templateVisiable.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
