<template>
    <el-dialog v-model="composeVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.compose') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
            <el-form-item :label="$t('container.name')" prop="name">
                <el-input v-model="form.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.from')">
                <el-radio-group v-model="form.from">
                    <el-radio label="edit">{{ $t('container.edit') }}</el-radio>
                    <el-radio label="path">{{ $t('container.pathSelect') }}</el-radio>
                    <el-radio label="template">{{ $t('container.composeTemplate') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.from === 'path'" prop="path">
                <el-input
                    clearable
                    :placeholder="$t('commons.example') + '/tmp/docker-compose.yml'"
                    v-model="form.path"
                >
                    <template #append>
                        <FileList @choose="loadDir" :dir="false"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item v-if="form.from === 'template'" prop="template">
                <el-select v-model="form.template">
                    <el-option v-for="item in templateOptions" :key="item.id" :value="item.id" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item v-if="form.from === 'edit'">
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
                    v-model="form.file"
                    :readOnly="true"
                />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="composeVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { listComposeTemplate, upCompose } from '@/api/modules/container';

const extensions = [javascript(), oneDark];
const composeVisiable = ref(false);
const templateOptions = ref();

const varifyPath = (rule: any, value: any, callback: any) => {
    if (value.indexOf('docker-compose.yml') === -1) {
        callback(new Error(i18n.global.t('commons.rule.selectHelper', ['docker-compose.yml'])));
    }
    callback();
};
const form = reactive({
    name: '',
    from: 'edit',
    path: '',
    file: '',
    template: 0,
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    path: [Rules.requiredSelect, { validator: varifyPath, trigger: 'change', required: true }],
});

const loadTemplates = async () => {
    const res = await listComposeTemplate();
    templateOptions.value = res.data;
    if (templateOptions.value && templateOptions.value.length !== 0) {
        form.template = templateOptions.value[0].id;
    }
};

const acceptParams = (): void => {
    composeVisiable.value = true;
    form.name = '';
    form.from = 'edit';
    form.path = '';
    form.file = '';
    loadTemplates();
};
const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await upCompose(form);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
        composeVisiable.value = false;
    });
};

const loadDir = async (path: string) => {
    form.path = path;
};

defineExpose({
    acceptParams,
});
</script>
