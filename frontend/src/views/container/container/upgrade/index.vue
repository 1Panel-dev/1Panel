<template>
    <DrawerPro v-model="drawerVisible" :header="$t('commons.button.upgrade')" @close="handleClose" size="large">
        <el-alert
            :title="$t('container.appHelper')"
            v-if="form.fromApp"
            class="common-prompt"
            :closable="false"
            type="warning"
        />
        <el-form @submit.prevent ref="formRef" :model="form" label-position="top">
            <el-form-item :label="$t('container.oldImage')" prop="oldImage">
                <el-tooltip placement="top-start" :content="form.oldImageName" v-if="form.oldImageName.length > 50">
                    <el-tag>{{ form.oldImageName.substring(0, 50) }}...</el-tag>
                </el-tooltip>
                <el-tag v-else>{{ form.oldImageName }}</el-tag>
            </el-form-item>
            <el-form-item :label="$t('container.sameImageContainer')" v-if="containerOptions.length > 1">
                <div class="w-full">
                    <el-checkbox v-model="checkAll" :indeterminate="isIndeterminate" @change="handleCheckAllChange">
                        {{ $t('commons.table.all') }}
                    </el-checkbox>
                </div>
                <el-checkbox-group v-model="form.names" @change="handleCheckedChange">
                    <el-checkbox
                        v-for="item in containerOptions"
                        :key="item.name"
                        :label="item.name"
                        :value="item.name"
                    >
                        {{ item.name }}
                    </el-checkbox>
                </el-checkbox-group>
                <span class="input-help">{{ $t('container.sameImageHelper') }}</span>
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
                <el-button :disabled="loading || form.names.length === 0" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" @close="closeTaskLog()" />
</template>

<script lang="ts" setup>
import { listContainerByImage, upgradeContainer } from '@/api/modules/container';
import { Rules } from '@/global/form-rules';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/util';
import { CheckboxValueType, ElForm } from 'element-plus';
import { reactive, ref } from 'vue';

const loading = ref(false);
const taskLogRef = ref();

const form = reactive({
    names: [],
    oldImageName: '',
    newImageName: '',
    hasName: true,

    fromApp: false,
    forcePull: false,
});

const formRef = ref<FormInstance>();

const drawerVisible = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;

const containerOptions = ref([]);
const isIndeterminate = ref();
const checkAll = ref();

interface DialogProps {
    container: string;
    image: string;
    fromApp: boolean;
}
const acceptParams = (props: DialogProps): void => {
    form.names = [props.container];
    isIndeterminate.value = true;
    form.oldImageName = props.image;
    form.fromApp = props.fromApp;
    form.hasName = props.image.indexOf('sha256:') === -1;
    if (form.hasName) {
        form.newImageName = props.image;
    } else {
        form.newImageName = '';
    }
    loadContainers();
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const loadContainers = async () => {
    const res = await listContainerByImage(form.oldImageName);
    containerOptions.value = res.data || [];
};

const handleCheckAllChange = (val: CheckboxValueType) => {
    form.names = [];
    if (!val) {
        isIndeterminate.value = false;
        return;
    }
    for (const item of containerOptions.value) {
        form.names.push(item.name);
    }
};
const handleCheckedChange = (value: CheckboxValueType[]) => {
    const checkedCount = value.length;
    checkAll.value = checkedCount === containerOptions.value.length;
    isIndeterminate.value = checkedCount > 0 && checkedCount < containerOptions.value.length;
};

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
                names: form.names,
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
