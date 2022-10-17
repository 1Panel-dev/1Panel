<template>
    <el-dialog
        v-model="buildVisiable"
        :destroy-on-close="true"
        @close="onCloseLog"
        :close-on-click-modal="false"
        width="50%"
    >
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.buildImage') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-width="80px">
            <el-form-item :label="$t('container.name')" :rules="Rules.requiredInput" prop="name">
                <el-input :placeholder="$t('container.imageNameHelper')" v-model="form.name" clearable />
            </el-form-item>
            <el-form-item label="Dockerfile" :rules="Rules.requiredSelect" prop="from">
                <el-radio-group v-model="form.from">
                    <el-radio label="edit">{{ $t('container.edit') }}</el-radio>
                    <el-radio label="path">{{ $t('container.pathSelect') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.from === 'edit'" :rules="Rules.requiredInput" prop="dockerfile">
                <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 10 }" v-model="form.dockerfile" />
            </el-form-item>
            <el-form-item v-else :rules="Rules.requiredSelect" prop="dockerfile">
                <el-input clearable v-model="form.dockerfile">
                    <template #append>
                        <FileList @choose="loadBuildDir" :dir="true"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('container.tag')">
                <el-input
                    :placeholder="$t('container.tagHelper')"
                    type="textarea"
                    :autosize="{ minRows: 2, maxRows: 4 }"
                    v-model="form.tagStr"
                />
            </el-form-item>
        </el-form>

        <codemirror
            v-if="logVisiable"
            :autofocus="true"
            placeholder="Wait for build output..."
            :indent-with-tab="true"
            :tabSize="4"
            style="max-height: 500px"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="logInfo"
            :readOnly="true"
            ref="buildLogRef"
        />

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="onSubmit(formRef)">{{ $t('container.build') }}</el-button>
                <el-button @click="buildVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import FileList from '@/components/file-list/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { imageBuild } from '@/api/modules/container';
import { LoadFile } from '@/api/modules/files';

const logVisiable = ref<boolean>(false);
const logInfo = ref();
const buildLogRef = ref();
const extensions = [javascript(), oneDark];
let timer: NodeJS.Timer | null = null;

const buildVisiable = ref(false);
const form = reactive({
    from: 'path',
    dockerfile: '',
    name: '',
    tagStr: '',
    tags: [] as Array<string>,
});
const varifyPath = (rule: any, value: any, callback: any) => {
    if (value.indexOf('docker-compose.yml') === -1) {
        callback(new Error(i18n.global.t('commons.rule.selectHelper', ['Dockerfile'])));
    }
    callback();
};
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    content: [Rules.requiredInput, { validator: varifyPath, trigger: 'change', required: true }],
});
const acceptParams = async () => {
    buildVisiable.value = true;
    form.from = 'path';
    form.dockerfile = '';
    form.tagStr = '';
    form.name = '';
};

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.tagStr !== '') {
            form.tags = form.tagStr.split('\n');
        }
        const res = await imageBuild(form);
        logVisiable.value = true;
        loadLogs(res.data);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const loadLogs = async (path: string) => {
    timer = setInterval(async () => {
        if (logVisiable.value) {
            const res = await LoadFile({ path: path });
            logInfo.value = res.data;
        }
    }, 1000 * 3);
};
const onCloseLog = async () => {
    emit('search');
    clearInterval(Number(timer));
    timer = null;
};

const loadBuildDir = async (path: string) => {
    form.dockerfile = path;
};

defineExpose({
    acceptParams,
});
</script>
