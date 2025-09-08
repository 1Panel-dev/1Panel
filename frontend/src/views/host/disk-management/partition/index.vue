<template>
    <DrawerPro
        v-model="open"
        :header="$t('disk.' + operate)"
        :resource="form.device.split('/').pop()"
        @close="handleClose"
        v-loading="loading"
    >
        <el-alert
            v-if="operate == 'partition' || (operate == 'mount' && filesystem == '')"
            :title="$t('disk.partitionAlert')"
            type="warning"
            :closable="false"
        />
        <el-form
            @submit.prevent
            ref="partitionRef"
            :rules="rules"
            label-position="top"
            :model="form"
            class="mt-2"
            v-loading="loading"
        >
            <el-form-item :label="$t('disk.filesystem')" prop="filesystem">
                <el-radio-group v-model="form.filesystem" :disabled="operate == 'mount' && filesystem != ''">
                    <el-radio-button label="ext4" value="ext4" />
                    <el-radio-button label="xfs" value="xfs" />
                </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('disk.autoMount')" prop="autoMount">
                <el-switch v-model="form.autoMount" />
            </el-form-item>
            <el-form-item :label="$t('disk.mountPoint')" prop="mountPoint">
                <el-input v-model="form.mountPoint">
                    <template #prepend>
                        <el-button icon="Folder" @click="fileRef.acceptParams({ dir: true })" />
                    </template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
        <FileList ref="fileRef" @choose="loadBuildDir" />
    </DrawerPro>
</template>

<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { mountDisk, partitionDisk } from '@/api/modules/host';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const rules = ref<any>({
    filesystem: [Rules.requiredInput],
    autoMount: [Rules.requiredInput],
    mountPoint: [Rules.requiredInput],
});
const form = reactive({
    device: '',
    filesystem: 'ext4',
    autoMount: true,
    mountPoint: '',
    label: '',
});
const open = ref(false);
const loading = ref(false);
const operate = ref('mount');
const fileRef = ref();
const emit = defineEmits(['search']);
const filesystem = ref('');

const loadBuildDir = async (path: string) => {
    form.mountPoint = path;
};

const acceptParams = (diskInfo: Host.DiskInfo, operateType: string) => {
    operate.value = operateType;
    form.device = diskInfo.device;
    form.mountPoint = '';
    if (operateType == 'mount' && diskInfo.filesystem) {
        filesystem.value = diskInfo.filesystem;
        form.filesystem = diskInfo.filesystem;
    }
    open.value = true;
};

const submit = async () => {
    try {
        loading.value = true;
        if (operate.value == 'mount') {
            await mountDisk(form);
            MsgSuccess(i18n.global.t('disk.mount') + i18n.global.t('commons.status.success'));
            handleClose();
        } else {
            await partitionDisk(form);
            MsgSuccess(i18n.global.t('disk.partition') + i18n.global.t('commons.status.success'));
            handleClose();
        }
    } catch (error) {
        loading.value = false;
        return;
    }
};

const handleClose = () => {
    open.value = false;
    loading.value = false;
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
