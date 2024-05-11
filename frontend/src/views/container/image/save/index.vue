<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="30%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.exportImage')" :back="handleClose" />
        </template>
        <el-form v-loading="loading" label-position="top" ref="formRef" :model="form" label-width="80px">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('container.tag')" :rules="Rules.requiredSelect" prop="tagName">
                        <el-select filterable v-model="form.tagName">
                            <el-option
                                :disabled="item.indexOf(':<none>') !== -1"
                                v-for="item in form.tags"
                                :key="item"
                                :value="item"
                                :label="item"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.path')" :rules="Rules.requiredInput" prop="path">
                        <el-input v-model="form.path">
                            <template #prepend>
                                <FileList @choose="loadSaveDir" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('container.fileName')" :rules="Rules.requiredInput" prop="name">
                        <el-input v-model="form.name">
                            <template #append>.tar</template>
                        </el-input>
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
                    {{ $t('container.export') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import FileList from '@/components/file-list/index.vue';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { imageSave } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const drawerVisible = ref(false);
const form = reactive({
    tags: [] as Array<string>,
    tagName: '',
    path: '',
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
    drawerVisible.value = true;
    form.tags = params.tags;
    form.tagName = form.tags.length !== 0 ? form.tags[0] : '';
    form.path = '';
    form.name = '';
    dialogData.value.repos = params.repos;
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
        loading.value = true;
        await imageSave(form)
            .then(() => {
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const loadSaveDir = async (path: string) => {
    form.path = path;
};

defineExpose({
    acceptParams,
});
</script>
