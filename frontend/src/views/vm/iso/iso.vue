<template>
    <LayoutContent :title="$t('xpack.vm.iso')">
        <template #leftToolBar>
            <el-button v-permission type="primary" :disabled="vmUnavailable" @click="openCreate">
                {{ $t('commons.button.create') }}
            </el-button>
        </template>
        <template #rightToolBar>
            <TableSearch v-model:searchName="searchName" @search="search" />
            <TableRefresh @search="search" />
        </template>
        <template #main>
            <div v-loading="loading" class="iso-card-content">
                <div class="iso-card-list" v-if="items.length">
                    <article v-for="item in items" :key="item.id" class="iso-card">
                        <div class="iso-card__header">
                            <div class="iso-card__title">
                                <span class="iso-card__name" :title="item.name">{{ item.name }}</span>
                            </div>
                            <el-tag :type="isoStatusType(item)" round size="small" effect="plain">
                                {{ isoStatusLabel(item) }}
                            </el-tag>
                        </div>
                        <div class="iso-card__sub" :title="item.path">{{ item.path || '-' }}</div>
                        <div class="iso-card__detail">
                            <div class="iso-card__detail-item">
                                <span>{{ $t('commons.table.type') }}</span>
                                <strong :title="item.type">{{ formatIsoType(item.type) }}</strong>
                            </div>
                        </div>
                        <div class="iso-card__detail">
                            <div class="iso-card__detail-item">
                                <span>{{ $t('xpack.vm.isoSize') }}</span>
                                <strong :title="item.description">
                                    {{ item.isExist ? computeSize(item.size) : '-' }}
                                </strong>
                            </div>
                        </div>
                        <div class="iso-card__detail">
                            <div class="iso-card__detail-item">
                                <span>{{ $t('commons.table.date') }}</span>
                                <strong :title="item.description">{{ formatDate(item.createdAt) }}</strong>
                            </div>
                        </div>
                        <div class="iso-card__detail iso-card__detail--last">
                            <div class="iso-card__detail-item">
                                <span>{{ $t('commons.table.description') }}</span>
                                <strong :title="item.description">{{ item.description || '-' }}</strong>
                            </div>
                        </div>
                        <div v-if="item.status === 'failed' && item.error" class="iso-card__error" :title="item.error">
                            {{ item.error }}
                        </div>
                        <div class="iso-card__actions">
                            <el-button
                                v-permission
                                link
                                type="primary"
                                :disabled="isImporting(item)"
                                @click="openEdit(item)"
                            >
                                {{ $t('commons.button.edit') }}
                            </el-button>
                            <el-button v-permission link type="primary" @click="openDelete(item)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </div>
                    </article>
                </div>
                <el-empty v-else-if="!loading" :description="$t('commons.msg.noneData')" />
            </div>
            <div v-if="paginationConfig.total > 0" class="iso-card-pagination">
                <el-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    :total="paginationConfig.total"
                    :page-sizes="[10, 20, 50, 100]"
                    layout="total, sizes, prev, pager, next, jumper"
                    @size-change="handleSizeChange"
                    @current-change="search"
                />
            </div>
        </template>
    </LayoutContent>

    <OpDialog ref="deleteRef" @search="handleDeleteSuccess" @cancel="resetDelete">
        <template #content>
            <div class="iso-delete-dialog__option">
                <el-checkbox v-model="deleteReq.deleteFile" :disabled="!deleteRow?.isExist">
                    {{ $t('xpack.vm.deleteSourceFile') }}
                </el-checkbox>
                <div class="iso-delete-dialog__helper">
                    {{ $t('xpack.vm.deleteSourceFileHelper') }}
                </div>
            </div>
        </template>
    </OpDialog>

    <Operate ref="operateRef" @search="search" />
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import i18n from '@/lang';
import { VM } from '@/api/interface/vm';
import { deleteVMIso, searchVMIso } from '@/api/modules/vm';
import { stopChunkUpload, stopWgetFile } from '@/api/modules/files';
import { computeSize } from '@/utils/size';
import { dateFormat } from '@/utils/date';
import Operate from '@/views/vm/iso/create/index.vue';
import OpDialog from '@/components/del-dialog/index.vue';

const isoTypeOptions = [
    { label: 'Debian', value: 'debian' },
    { label: 'Ubuntu', value: 'ubuntu' },
    { label: 'CentOS', value: 'centos' },
    { label: 'Rocky Linux', value: 'rocky' },
    { label: 'AlmaLinux', value: 'almalinux' },
    { label: 'Fedora', value: 'fedora' },
    { label: 'openSUSE', value: 'opensuse' },
    { label: 'RHEL', value: 'rhel' },
    { label: 'Oracle Linux', value: 'oraclelinux' },
    { label: 'Arch Linux', value: 'archlinux' },
    { label: 'Windows', value: 'windows' },
    { label: 'FreeBSD', value: 'freebsd' },
];

const props = defineProps<{
    vmUnavailable: boolean;
    vmEnvironmentLoaded: boolean;
}>();
const vmUnavailable = computed(() => props.vmUnavailable);
const loading = ref(false);
const searchName = ref('');
const items = ref<VM.IsoItem[]>([]);
const operateRef = ref();
const deleteRef = ref();
const deleteRow = ref<VM.IsoItem>();
const deleteReq = reactive<VM.IsoDelete>({
    id: 0,
    deleteFile: false,
});

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const search = async () => {
    if (!props.vmEnvironmentLoaded || vmUnavailable.value) {
        items.value = [];
        paginationConfig.total = 0;
        return;
    }
    loading.value = true;
    try {
        const res = await searchVMIso({
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
            info: searchName.value,
        });
        items.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
    } finally {
        loading.value = false;
    }
};

const handleSizeChange = () => {
    paginationConfig.currentPage = 1;
    search();
};

const formatDate = (value: string) => (value ? dateFormat(null, null, value) : '-');
const formatIsoType = (value: string) => isoTypeOptions.find((item) => item.value === value)?.label || value || '-';
const isoStatusType = (item: VM.IsoItem) => {
    if (item.status === 'failed' || (!item.isExist && item.status === 'ready')) return 'danger';
    if (item.status === 'uploading' || item.status === 'downloading') return 'warning';
    return 'success';
};
const isoStatusLabel = (item: VM.IsoItem) => {
    if (item.status === 'uploading') return i18n.global.t('commons.status.uploading');
    if (item.status === 'downloading') return i18n.global.t('commons.status.downloading');
    if (item.status === 'failed') return i18n.global.t('commons.status.failed');
    return item.isExist ? i18n.global.t('commons.status.ready') : i18n.global.t('commons.status.deleted');
};
const isImporting = (item: VM.IsoItem) => ['uploading', 'downloading'].includes(item.status);

const openCreate = () => {
    if (vmUnavailable.value) {
        return;
    }
    operateRef.value?.acceptParams();
};

const openEdit = (row: VM.IsoItem) => operateRef.value?.acceptParams(row);

const openDelete = (row: VM.IsoItem) => {
    deleteRow.value = row;
    deleteReq.id = row.id;
    deleteReq.deleteFile = false;
    deleteRef.value?.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        msg: i18n.global.t('xpack.vm.deleteIsoConfirm'),
        names: [],
        api: async (params: VM.IsoDelete) => {
            if (deleteRow.value?.tempFileID) {
                if (deleteRow.value.tempFileType === 'upload') {
                    await stopChunkUpload(deleteRow.value.tempFileID).catch(() => undefined);
                }
                if (deleteRow.value.tempFileType === 'download') {
                    await stopWgetFile(deleteRow.value.tempFileID).catch(() => undefined);
                }
            }
            return deleteVMIso(params);
        },
        params: deleteReq,
        successMsg: i18n.global.t('commons.msg.operationSuccess'),
    });
};

const resetDelete = () => {
    deleteRow.value = undefined;
    deleteReq.id = 0;
    deleteReq.deleteFile = false;
};

const handleDeleteSuccess = () => {
    resetDelete();
    search();
};

watch(() => [props.vmEnvironmentLoaded, props.vmUnavailable], search, { immediate: true });
</script>

<style scoped lang="scss">
.iso-card-content {
    min-height: 200px;
}

.iso-delete-dialog__option {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.iso-delete-dialog__helper {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 20px;
}

.iso-card-list {
    display: grid;
    grid-template-columns: repeat(3, minmax(360px, 1fr));
    gap: 14px;
}

.iso-card {
    display: flex;
    flex-direction: column;
    min-height: 160px;
    padding: 16px;
    border: 1px solid var(--el-card-border-color, var(--el-border-color-lighter));
    border-radius: 8px;
    background: var(--el-card-bg-color, var(--el-fill-color-blank));
}

.iso-card__header,
.iso-card__title,
.iso-card__detail-item,
.iso-card__actions {
    display: flex;
    align-items: center;
}

.iso-card__error {
    overflow: hidden;
    color: var(--el-color-danger);
    font-size: 12px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.iso-card__header {
    justify-content: space-between;
    gap: 12px;
}

.iso-card__title {
    min-width: 0;
    gap: 8px;
}

.iso-card__name {
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 16px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.iso-card__sub {
    overflow: hidden;
    min-height: 22px;
    margin-top: 8px;
    color: var(--el-text-color-regular);
    font-size: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.iso-card__meta {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
    margin-top: 14px;
}

.iso-card__meta-item {
    display: block;
    min-width: 0;
    padding: 10px;
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.iso-card__meta-label,
.iso-card__detail-item span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.iso-card__meta-label {
    display: block;
    margin-bottom: 4px;
}

.iso-card__meta-value {
    display: block;
    overflow: hidden;
    min-width: 0;
    color: var(--el-text-color-regular);
    font-size: 13px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.iso-card__detail {
    display: grid;
    gap: 8px;
    margin-top: 16px;
}

.iso-card__detail--last {
    margin-bottom: 14px;
}

.iso-card__detail-item {
    min-width: 0;
    justify-content: space-between;
    gap: 12px;
}

.iso-card__detail-item strong {
    overflow: hidden;
    min-width: 0;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 500;
    text-align: right;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.iso-card__actions {
    justify-content: flex-end;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: auto;
    padding-top: 14px;
    border-top: 1px solid var(--el-border-color-lighter);
}

.iso-card__actions :deep(.el-button) {
    margin-left: 0;
}

.iso-card-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
}

.iso-exist-status {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
}

@media (max-width: 1280px) {
    .iso-card-list {
        grid-template-columns: repeat(2, minmax(360px, 1fr));
    }
}

@media (max-width: 768px) {
    .iso-card-list {
        grid-template-columns: 1fr;
    }

    .iso-card-pagination {
        overflow-x: auto;
    }
}
</style>
