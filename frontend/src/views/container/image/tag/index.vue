<template>
    <el-drawer v-model="tagVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <div class="card-header">
                <span>Tag {{ $t('container.image') }}</span>
            </div>
        </template>
        <el-form v-loading="loading" ref="formRef" :model="form" label-width="80px">
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
            <el-form-item :label="$t('container.imageName')" :rules="Rules.requiredInput" prop="targetName">
                <el-input v-model="form.targetName">
                    <template v-if="form.fromRepo" #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabeld="loading" @click="tagVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabeld="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.save') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { imageTag } from '@/api/modules/container';
import { Container } from '@/api/interface/container';

const loading = ref(false);

const tagVisiable = ref(false);
const repos = ref();
const form = reactive({
    sourceID: '',
    fromRepo: true,
    repoID: 1,
    targetName: '',
});

interface DialogProps {
    repos: Array<Container.RepoOptions>;
    sourceID: string;
}

const acceptParams = async (params: DialogProps): Promise<void> => {
    tagVisiable.value = true;
    form.repoID = 1;
    form.sourceID = params.sourceID;
    form.targetName = '';
    form.fromRepo = true;
    repos.value = params.repos;
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
        loading.value = true;
        await imageTag(form)
            .then(() => {
                loading.value = false;
                tagVisiable.value = false;
                emit('search');
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
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
