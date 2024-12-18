<template>
    <el-drawer
        v-model="drawerVisible"
        @close="onCloseLog"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.imagePull')" :back="onCloseLog" />
        </template>
        <el-row type="flex" justify="center">
            <el-col :span="22">
                <el-form ref="formRef" label-position="top" :model="form">
                    <el-form-item :label="$t('container.from')">
                        <el-checkbox @change="onEdit()" v-model="form.fromRepo">
                            {{ $t('container.imageRepo') }}
                        </el-checkbox>
                    </el-form-item>
                    <el-form-item
                        v-if="form.fromRepo"
                        :label="$t('container.repoName')"
                        :rules="Rules.requiredSelect"
                        prop="repoID"
                    >
                        <el-select @change="onEdit()" clearable style="width: 100%" filterable v-model="form.repoID">
                            <el-option v-for="item in repos" :key="item.id" :value="item.id" :label="item.name" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.imageName')" :rules="Rules.imageName" prop="imageName">
                        <el-input @change="onEdit()" v-model.trim="form.imageName">
                            <template v-if="form.fromRepo" #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                        </el-input>
                    </el-form-item>
                </el-form>
                <LogFile
                    ref="logRef"
                    :config="logConfig"
                    :default-button="false"
                    v-model:is-reading="isReading"
                    v-if="showLog"
                    :style="'height: calc(100vh - 397px);min-height: 200px'"
                />
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="isStartReading || isReading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('container.pull') }}
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
import { imagePull } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref(false);
const form = reactive({
    fromRepo: true,
    repoID: null as number,
    imageName: '',
});
const logConfig = reactive({
    type: 'image-pull',
    name: '',
});
const showLog = ref(false);
const logRef = ref();
const logVisible = ref(false);
const logInfo = ref();
const isStartReading = ref(false);
const isReading = ref(false);

interface DialogProps {
    repos: Array<Container.RepoOptions>;
}
const repos = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    logVisible.value = false;
    drawerVisible.value = true;
    form.fromRepo = true;
    form.imageName = '';
    repos.value = params.repos;
    form.repoID = 1;
    for (const item of repos.value) {
        if (item.name === 'Docker Hub' && item.downloadUrl === 'docker.io') {
            form.repoID = item.id;
            break;
        }
    }
    isStartReading.value = false;
    logInfo.value = '';
    showLog.value = false;
};
const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onEdit = () => {
    if (!isReading.value && isStartReading.value) {
        isStartReading.value = false;
    }
};
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!form.fromRepo) {
            form.repoID = 0;
        }
        const res = await imagePull(form);
        logVisible.value = true;
        isStartReading.value = true;
        logConfig.name = res.data;
        search();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const search = () => {
    showLog.value = false;
    nextTick(() => {
        showLog.value = true;
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
