<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t('commons.button.upgrade')"
        @close="handleClose"
        :resource="form.name"
        size="large"
    >
        <el-alert
            :title="$t('container.appHelper')"
            v-if="form.fromApp"
            class="common-prompt"
            :closable="false"
            type="error"
        />
        <el-form @submit.prevent ref="formRef" :model="form" label-position="top">
            <el-form-item :label="$t('container.oldImage')" prop="oldImage">
                <el-tooltip placement="top-start" :content="form.oldImageName" v-if="form.oldImageName.length > 50">
                    <el-tag>{{ form.oldImageName.substring(0, 50) }}...</el-tag>
                </el-tooltip>
                <el-tag v-else>{{ form.oldImageName }}</el-tag>
            </el-form-item>
            <el-form-item prop="newImageName" :rules="Rules.imageName">
                <template #label>
                    {{ $t('container.targetImage') }}
                    <span v-if="!form.hasName">
                        {{ ' (' + $t('container.imageLoadErr') + ')' }}
                    </span>
                </template>
                <el-input v-model="form.newImageName" />
                <span class="input-help">{{ $t('container.upgradeHelper') }}</span>
            </el-form-item>
            <el-form-item prop="forcePull">
                <el-checkbox v-model="form.forcePull">
                    {{ $t('container.forcePull') }}
                </el-checkbox>
                <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" @close="closeTaskLog()" />
</template>

<script lang="ts" setup>
import { upgradeContainer } from '@/api/modules/container';
import { Rules } from '@/global/form-rules';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/util';
import { ElForm } from 'element-plus';
import { reactive, ref } from 'vue';

const loading = ref(false);
const taskLogRef = ref();

const form = reactive({
    name: '',
    oldImageName: '',
    newImageName: '',
    hasName: true,

    fromApp: false,
    forcePull: false,
});

const formRef = ref<FormInstance>();

const drawerVisible = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;

interface DialogProps {
    container: string;
    image: string;
    fromApp: boolean;
}
const acceptParams = (props: DialogProps): void => {
    form.name = props.container;
    form.oldImageName = props.image;
    form.fromApp = props.fromApp;
    form.hasName = props.image.indexOf('sha256:') === -1;
    if (form.hasName) {
        form.newImageName = props.image;
    } else {
        form.newImageName = '';
    }
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('container.upgradeWarning2'), i18n.global.t('commons.button.upgrade'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            let taskID = newUUID();
            let param = {
                taskID: taskID,
                name: form.name,
                image: form.newImageName,
                forcePull: form.forcePull,
            };
            loading.value = true;
            await upgradeContainer(param)
                .then(() => {
                    loading.value = false;
                    openTaskLog(taskID);
                    drawerVisible.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const handleClose = async () => {
    drawerVisible.value = false;
    emit('search');
};

const closeTaskLog = () => {
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
