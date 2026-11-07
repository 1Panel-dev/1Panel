<template>
    <el-dialog v-model="visible" :title="$t('xpack.vm.orphanClean')" width="860px" destroy-on-close>
        <el-alert :title="$t('xpack.vm.orphanCleanHelper')" type="warning" :closable="false" show-icon />
        <div class="orphan-clean__toolbar">
            <span>{{ $t('xpack.vm.orphanSelected', [selected.length, computeSize(selectedSize)]) }}</span>
            <el-button :loading="scanning" @click="scan">{{ $t('xpack.vm.orphanScan') }}</el-button>
        </div>
        <el-table
            :tooltip-options="TABLE_OVERFLOW_TOOLTIP_OPTIONS"
            v-loading="scanning"
            :data="items"
            max-height="420"
            @selection-change="selectionChange"
        >
            <el-table-column type="selection" width="48" />
            <el-table-column :label="$t('commons.table.name')" min-width="280" show-overflow-tooltip>
                <template #default="scope">{{ scope.row.path }}</template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.type')" min-width="130">
                <template #default="scope">{{ typeLabel(scope.row.type) }}</template>
            </el-table-column>
            <el-table-column :label="$t('xpack.vm.orphanSize')" width="120">
                <template #default="scope">{{ computeSize(scope.row.size) }}</template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.date')" width="180">
                <template #default="scope">{{ formatDate(scope.row.modifiedAt) }}</template>
            </el-table-column>
        </el-table>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :disabled="!selected.length" :loading="cleaning" @click="clean">
                {{ $t('xpack.vm.orphanClean') }}
            </el-button>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { TABLE_OVERFLOW_TOOLTIP_OPTIONS } from '@/components/table/shared';
import { computed, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { VM } from '@/api/interface/vm';
import { cleanVMOrphans, scanVMOrphans } from '@/api/modules/vm';
import { computeSize } from '@/utils/size';
import { dateFormat } from '@/utils/date';
import { MsgSuccess } from '@/utils/message';

const visible = ref(false);
const scanning = ref(false);
const cleaning = ref(false);
const items = ref<VM.OrphanItem[]>([]);
const selected = ref<VM.OrphanItem[]>([]);
const selectedSize = computed(() => selected.value.reduce((total, item) => total + (item.size || 0), 0));

const formatDate = (value: number) => (value ? dateFormat(null, null, value * 1000) : '-');
const typeLabel = (type: VM.OrphanItem['type']) => i18n.global.t(`xpack.vm.orphanType.${type}`);
const selectionChange = (value: VM.OrphanItem[]) => {
    selected.value = value;
};
const scan = async () => {
    scanning.value = true;
    try {
        const res = await scanVMOrphans();
        items.value = res.data || [];
        selected.value = [];
    } finally {
        scanning.value = false;
    }
};
const open = async () => {
    visible.value = true;
    await scan();
};
const clean = async () => {
    try {
        await ElMessageBox.confirm(
            i18n.global.t('xpack.vm.orphanCleanConfirm', [selected.value.length, computeSize(selectedSize.value)]),
            i18n.global.t('xpack.vm.orphanClean'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'warning',
            },
        );
    } catch {
        return;
    }
    cleaning.value = true;
    try {
        const res = await cleanVMOrphans(selected.value);
        const removed = res.data || [];
        const removedPaths = new Set(removed.map((item) => item.path));
        items.value = items.value.filter((item) => !removedPaths.has(item.path));
        selected.value = [];
        MsgSuccess(
            i18n.global.t('xpack.vm.orphanCleanSuccess', [
                removed.length,
                computeSize(removed.reduce((total, item) => total + item.size, 0)),
            ]),
        );
    } finally {
        cleaning.value = false;
    }
};

defineExpose({ open });
</script>

<style scoped lang="scss">
.orphan-clean__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 16px 0;
    color: var(--el-text-color-secondary);
}
</style>
