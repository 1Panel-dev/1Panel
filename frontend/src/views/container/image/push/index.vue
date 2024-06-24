<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        @close="onCloseLog"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.imagePush')" :back="onCloseLog" />
        </template>
        <el-row type="flex" justify="center">
            <el-col :span="22">
                <el-form ref="formRef" label-position="top" :model="form" label-width="80px">
                    <el-form-item :label="$t('container.tag')" :rules="Rules.requiredSelect" prop="tagName">
                        <el-select @change="onEdit(true)" filterable v-model="form.tagName">
                            <el-option v-for="item in form.tags" :key="item" :value="item" :label="item" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.repoName')" :rules="Rules.requiredSelect" prop="repoID">
                        <el-select @change="onEdit()" clearable style="width: 100%" filterable v-model="form.repoID">
                            <el-option
                                v-for="item in dialogData.repos"
                                :key="item.id"
                                :value="item.id"
                                :label="item.name"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.image')" :rules="Rules.imageName" prop="name">
                        <el-input @change="onEdit()" v-model.trim="form.name">
                            <template #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                        </el-input>
                    </el-form-item>
                </el-form>

                <LogFile
                    ref="logRef"
                    :config="logConfig"
                    :default-button="false"
                    v-model:is-reading="isReading"
                    v-if="logVisible"
                    :style="'height: calc(100vh - 370px);min-height: 200px'"
                    v-model:loading="loading"
                />
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="isStartReading || isReading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('container.push') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { nextTick, reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { imagePush } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref(false);
const form = reactive({
    tags: [] as Array<string>,
    tagName: '',
    repoID: 1,
    name: '',
});

const logVisible = ref(false);
const loading = ref(false);
const isStartReading = ref(false);
const isReading = ref(false);

const logRef = ref();
const logConfig = reactive({
    type: 'image-push',
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
    logVisible.value = false;
    loading.value = false;
    drawerVisible.value = true;
    form.tags = params.tags;
    form.repoID = 1;
    form.tagName = form.tags.length !== 0 ? form.tags[0] : '';
    form.name = form.tags.length !== 0 ? form.tags[0] : '';
    dialogData.value.repos = params.repos;
    isStartReading.value = false;
};
const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onEdit = (isName?: boolean) => {
    if (!isReading.value && isStartReading.value) {
        isStartReading.value = false;
    }
    if (isName) {
        form.name = form.tagName;
    }
};
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        const res = await imagePush(form);
        logVisible.value = true;
        isStartReading.value = true;
        logConfig.name = res.data;
        loadLogs();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const loadLogs = () => {
    logVisible.value = false;
    nextTick(() => {
        logVisible.value = true;
        nextTick(() => {
            logRef.value.changeTail(true);
        });
    });
};

const onCloseLog = async () => {
    emit('search');
    drawerVisible.value = false;
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
