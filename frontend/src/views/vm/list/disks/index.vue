<template>
    <DrawerPro
        v-model="open"
        :header="$t('xpack.vm.disks')"
        :resource="name"
        size="70%"
        full-screen
        @close="handleClose"
    >
        <template #content>
            <p class="input-help mb-4">{{ $t('xpack.vm.diskResizeLimit') }}</p>
            <div class="mb-4">
                <el-button :disabled="loading" @click="search">{{ $t('commons.button.refresh') }}</el-button>
            </div>
            <el-table :tooltip-options="TABLE_OVERFLOW_TOOLTIP_OPTIONS" v-loading="loading" :data="items">
                <el-table-column prop="target" :label="$t('xpack.vm.diskTarget')" width="85" />
                <el-table-column :label="$t('xpack.vm.diskSize')" width="110">
                    <template #default="{ row }">{{ row.capacity > 0 ? computeSize(row.capacity) : '-' }}</template>
                </el-table-column>
                <el-table-column prop="bus" :label="$t('xpack.vm.diskBus')" width="100" />
                <el-table-column prop="source" :label="$t('xpack.vm.diskPath')" min-width="160" show-overflow-tooltip />
                <el-table-column :label="$t('commons.table.status')" min-width="250">
                    <template #default="{ row }">
                        {{ row.resizeBlocked ? $t(`xpack.vm.diskResizeReason.${row.resizeBlocked}`) : '-' }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="90" fixed="right">
                    <template #default="{ row }">
                        <el-button
                            v-permission
                            link
                            type="primary"
                            :disabled="loading || !!row.resizeBlocked || !row.capacity"
                            @click="openResize(row)"
                        >
                            {{ $t('xpack.vm.resizeDisk') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </template>
    </DrawerPro>

    <DialogPro v-model="resizeVisible" :title="$t('xpack.vm.resizeDisk')" size="small" :close-on-click-modal="false">
        <el-form v-if="selected" ref="formRef" :model="form" :rules="rules" label-position="top">
            <el-form-item :label="$t('xpack.vm.diskTarget')">{{ selected.target }}</el-form-item>
            <el-form-item :label="$t('xpack.vm.diskSize')">{{ computeSize(selected.capacity) }}</el-form-item>
            <el-form-item :label="$t('xpack.vm.diskCapacityAfterResize')" prop="sizeGiB">
                <el-input-number
                    v-model="form.sizeGiB"
                    :min="minSize"
                    :max="8388607"
                    :precision="0"
                    :disabled="submitting"
                />
                <span class="ml-2">GiB</span>
            </el-form-item>
            <p class="input-help">{{ $t('xpack.vm.diskResizeHelper') }}</p>
        </el-form>
        <template #footer>
            <el-button :disabled="submitting" @click="resizeVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button v-permission type="primary" :loading="submitting" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
    <TaskLog ref="taskLogRef" width="70%" @close="refreshAfterTask" />
</template>

<script setup lang="ts">
import { TABLE_OVERFLOW_TOOLTIP_OPTIONS } from '@/components/table/shared';
import { computed, reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import DrawerPro from '@/components/drawer-pro/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import { Rules } from '@/global/form-rules';
import { newUUID } from '@/utils/id';
import { computeSizeForDocker as computeSize } from '@/utils/size';
import { VM } from '@/api/interface/vm';
import { getVM, resizeVMDisk } from '@/api/modules/vm';

const emit = defineEmits(['refresh']);
const open = ref(false);
const name = ref('');
const loading = ref(false);
const items = ref<VM.Disk[]>([]);
const selected = ref<VM.Disk>();
const resizeVisible = ref(false);
const submitting = ref(false);
const formRef = ref<FormInstance>();
const taskLogRef = ref();
const form = reactive({ sizeGiB: 1 });
const rules = { sizeGiB: [Rules.requiredInput, Rules.integerNumber] };
const minSize = computed(() => Math.floor((selected.value?.capacity || 0) / 1024 ** 3) + 1);

const search = async () => {
    if (!open.value || !name.value) return;
    const requestedName = name.value;
    loading.value = true;
    items.value = [];
    try {
        const res = await getVM(requestedName, true);
        if (open.value && name.value === requestedName) {
            items.value = (res.data.disks || []).filter((disk) => disk.device === 'disk');
        }
    } finally {
        loading.value = false;
    }
};

const acceptParams = (row: VM.Item) => {
    name.value = row.name;
    open.value = true;
    search();
};

const openResize = (disk: VM.Disk) => {
    selected.value = disk;
    form.sizeGiB = Math.floor(disk.capacity / 1024 ** 3) + 1;
    resizeVisible.value = true;
    formRef.value?.clearValidate();
};

const submit = async () => {
    if (!selected.value || submitting.value) return;
    await formRef.value?.validate();
    submitting.value = true;
    const taskID = newUUID();
    try {
        const res = await resizeVMDisk({
            name: name.value,
            target: selected.value.target,
            sizeGiB: form.sizeGiB,
            taskID,
        });
        resizeVisible.value = false;
        taskLogRef.value?.openWithTaskID(res.data.taskID || taskID, true);
    } finally {
        submitting.value = false;
    }
};

const refreshAfterTask = () => {
    search();
    emit('refresh');
};

const handleClose = () => {
    items.value = [];
    resizeVisible.value = false;
    selected.value = undefined;
};

defineExpose({ acceptParams });
</script>
