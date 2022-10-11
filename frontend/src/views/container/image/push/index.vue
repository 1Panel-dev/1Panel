<template>
    <el-dialog v-model="pushVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.imagePush') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-width="80px">
            <el-form-item label="Tag" :rules="Rules.requiredSelect" prop="tagName">
                <el-select filterable v-model="form.tagName">
                    <el-option v-for="item in form.tags" :key="item" :value="item" :label="item" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.repoName')" :rules="Rules.requiredSelect" prop="repoID">
                <el-select style="width: 100%" filterable v-model="form.repoID">
                    <el-option v-for="item in dialogData.repos" :key="item.id" :value="item.id" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.label')" :rules="Rules.requiredInput" prop="name">
                <el-input v-model="form.name">
                    <template #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="onSubmit(formRef)">
                    {{ $t('container.push') }}
                </el-button>
                <el-button @click="pushVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { imagePush } from '@/api/modules/container';
import { Container } from '@/api/interface/container';

const pushVisiable = ref(false);
const form = reactive({
    tags: [] as Array<string>,
    tagName: '',
    repoID: 1,
    name: '',
});

interface DialogProps {
    repos: Array<Container.RepoOptions>;
    tags: Array<string>;
}
const dialogData = ref<DialogProps>({
    repos: [] as Array<Container.RepoOptions>,
    tags: [] as Array<string>,
});

const acceptParams = async (params: DialogProps): Promise<void> => {
    pushVisiable.value = true;
    form.tags = params.tags;
    form.repoID = 1;
    form.tagName = '';
    form.name = '';
    dialogData.value.repos = params.repos;
};

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            pushVisiable.value = false;
            await imagePush(form);
            emit('search');
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            emit('search');
        }
    });
};

function loadDetailInfo(id: number) {
    for (const item of dialogData.value.repos) {
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
