<template>
    <el-dialog
        v-model="pullVisiable"
        @close="onCloseLog"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="50%"
    >
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.imagePull') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-width="80px">
            <el-form-item :label="$t('container.from')">
                <el-checkbox v-model="form.fromRepo">{{ $t('container.imageRepo') }}</el-checkbox>
            </el-form-item>
            <el-form-item
                v-if="form.fromRepo"
                :label="$t('container.repoName')"
                :rules="Rules.requiredSelect"
                prop="repoID"
            >
                <el-select style="width: 100%" filterable v-model="form.repoID">
                    <el-option v-for="item in repos" :key="item.id" :value="item.id" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.imageName')" :rules="Rules.requiredInput" prop="imageName">
                <el-input v-model="form.imageName">
                    <template v-if="form.fromRepo" #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                </el-input>
            </el-form-item>
        </el-form>

        <codemirror
            v-if="logVisiable"
            :autofocus="true"
            placeholder="Wait for pull output..."
            :indent-with-tab="true"
            :tabSize="4"
            style="max-height: 300px"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="logInfo"
            :readOnly="true"
        />
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="buttonDisabled" @click="pullVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="buttonDisabled" type="primary" @click="onSubmit(formRef)">
                    {{ $t('container.pull') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { imagePull } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';

const pullVisiable = ref(false);
const form = reactive({
    fromRepo: true,
    repoID: null as number,
    imageName: '',
});

const buttonDisabled = ref(false);

const logVisiable = ref(false);
const logInfo = ref();
const extensions = [javascript(), oneDark];
let timer: NodeJS.Timer | null = null;

interface DialogProps {
    repos: Array<Container.RepoOptions>;
}
const repos = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    pullVisiable.value = true;
    form.fromRepo = true;
    form.imageName = '';
    repos.value = params.repos;
    buttonDisabled.value = false;
    logInfo.value = '';
};

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!form.fromRepo) {
            form.repoID = 0;
        }
        const res = await imagePull(form);
        logVisiable.value = true;
        buttonDisabled.value = true;
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

function loadDetailInfo(id: number) {
    for (const item of repos.value) {
        if (item.id === id) {
            return item.downloadUrl;
        }
    }
    return '';
}

defineExpose({
    acceptParams,
});
</script>
