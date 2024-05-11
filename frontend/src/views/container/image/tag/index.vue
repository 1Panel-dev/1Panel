<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.imageTag')" :back="handleClose" />
        </template>
        <el-form v-loading="loading" label-position="top" ref="formRef" :model="form" label-width="80px">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('container.from')">
                        <el-checkbox v-model="form.fromRepo">{{ $t('container.imageRepo') }}</el-checkbox>
                    </el-form-item>
                    <el-form-item
                        v-if="form.fromRepo"
                        :label="$t('container.repoName')"
                        :rules="Rules.requiredSelect"
                        prop="repo"
                    >
                        <el-select style="width: 100%" clearable filterable v-model="form.repo" @change="changeRepo">
                            <el-option v-for="item in repos" :key="item.id" :value="item.name" :label="item.name" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.imageTag')" :rules="Rules.imageName" prop="targetName">
                        <el-input v-model="form.targetName" />
                    </el-form-item>

                    <el-form-item>
                        <el-checkbox style="width: 100%" v-model="form.deleteTag">
                            {{ $t('container.imageTagDeleteHelper') }}
                        </el-checkbox>
                        <el-checkbox-group class="ml-5" v-if="form.deleteTag" v-model="form.deleteTags">
                            <el-checkbox
                                style="width: 100%"
                                v-for="item in tags"
                                :key="item"
                                :value="item"
                                :label="item"
                            />
                        </el-checkbox-group>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
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
import { ElForm } from 'element-plus';
import { imageRemove, imageTag } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const drawerVisible = ref(false);
const repos = ref();
const tags = ref();
const form = reactive({
    imageID: '',
    fromRepo: false,
    repo: '',
    originName: '',
    targetName: '',

    deleteTag: false,
    deleteTags: [],
});

interface DialogProps {
    repos: Array<Container.RepoOptions>;
    imageID: string;
    tags: Array<string>;
}

const acceptParams = async (params: DialogProps): Promise<void> => {
    drawerVisible.value = true;
    form.imageID = params.imageID;
    form.originName = params.tags?.length !== 0 ? params.tags[0] : '';
    form.targetName = params.tags?.length !== 0 ? params.tags[0] : '';
    form.fromRepo = false;
    form.repo = '';
    form.deleteTag = false;
    form.deleteTags = [];
    repos.value = params.repos;
    tags.value = params.tags;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            sourceID: form.imageID,
            targetName: form.targetName,
        };
        loading.value = true;
        await imageTag(params)
            .then(async () => {
                loading.value = false;
                if (form.deleteTag && form.deleteTags.length !== 0) {
                    await imageRemove({ names: form.deleteTags });
                }
                drawerVisible.value = false;
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const changeRepo = (val) => {
    if (val === 'Docker Hub') {
        form.targetName = form.originName;
        return;
    }
    for (const item of repos.value) {
        if (item.name == val) {
            form.targetName = item.downloadUrl + '/' + form.originName;
            return;
        }
    }
};

defineExpose({
    acceptParams,
});
</script>
