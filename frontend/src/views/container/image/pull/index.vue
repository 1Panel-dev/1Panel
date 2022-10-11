<template>
    <el-dialog v-model="pullVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
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
                    <el-option v-for="item in dialogData.repos" :key="item.id" :value="item.id" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.imageName')" :rules="Rules.requiredInput" prop="imageName">
                <el-input v-model="form.imageName">
                    <template v-if="form.fromRepo" #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="onSubmit(formRef)">
                    {{ $t('container.pull') }}
                </el-button>
                <el-button @click="pullVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
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

const pullVisiable = ref(false);
const form = reactive({
    fromRepo: true,
    repoID: 1,
    imageName: '',
});

interface DialogProps {
    repos: Array<Container.RepoOptions>;
}
const dialogData = ref<DialogProps>({
    repos: [] as Array<Container.RepoOptions>,
});

const acceptParams = async (params: DialogProps): Promise<void> => {
    pullVisiable.value = true;
    form.fromRepo = true;
    form.repoID = 1;
    form.imageName = '';
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
            if (!form.fromRepo) {
                form.repoID = 0;
            }
            pullVisiable.value = false;
            await imagePull(form);
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
