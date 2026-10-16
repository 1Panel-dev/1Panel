<template>
    <el-popover
        trigger="click"
        placement="bottom"
        :width="680"
        :popper-style="{ maxWidth: 'calc(100vw - 32px)' }"
        :visible="visible"
        @update:visible="setVisible"
        :disabled="disabled"
        :title="$t('home.quickDir')"
        @show="load"
    >
        <template #reference>
            <el-button link type="primary" :disabled="disabled">{{ count }}</el-button>
        </template>
        <div v-loading="loading" class="quick-directories">
            <div class="directory-search">
                <el-input
                    v-model="searchInput"
                    clearable
                    prefix-icon="Search"
                    :placeholder="`${$t('file.path')} / ${$t('home.alias')}`"
                    :aria-label="$t('commons.button.search')"
                    @keyup.enter="search"
                    @clear="search"
                />
                <el-button :disabled="loading" @click="search">{{ $t('commons.button.confirm') }}</el-button>
            </div>
            <div class="directory-row directory-header">
                <span>{{ $t('file.path') }}</span>
                <span>{{ $t('home.alias') }}</span>
                <span />
            </div>
            <div class="directory-list">
                <div v-if="!filteredRows.length" class="directory-empty">{{ $t('commons.msg.noneData') }}</div>
                <div v-for="row in filteredRows" :key="row.key" class="directory-row">
                    <el-input
                        v-model="row.path"
                        :readonly="!row.editing"
                        :disabled="row.busy"
                        :aria-label="$t('file.path')"
                    >
                        <template #prepend>
                            <el-button
                                icon="Folder"
                                :disabled="!row.editing || row.busy"
                                :aria-label="$t('file.list')"
                                @click="selectPath(row)"
                            />
                        </template>
                    </el-input>
                    <el-input
                        v-model="row.alias"
                        :readonly="!row.editing"
                        :disabled="row.busy"
                        :aria-label="$t('home.alias')"
                    />
                    <div class="directory-actions">
                        <el-button
                            link
                            :type="row.editing ? 'success' : 'primary'"
                            :icon="row.editing ? 'Check' : 'Edit'"
                            :aria-label="$t(row.editing ? 'commons.button.save' : 'commons.button.edit')"
                            :disabled="loading || row.busy || (row.editing && !row.path.trim())"
                            @click="row.editing ? save(row) : (row.editing = true)"
                        />
                        <el-button
                            v-if="row.id || rows.length > 1"
                            link
                            :type="row.editing ? 'info' : 'danger'"
                            :icon="row.editing ? 'Close' : 'Delete'"
                            :aria-label="$t('commons.button.delete')"
                            :disabled="loading || row.busy"
                            @click="remove(row)"
                        />
                        <el-button
                            v-if="row.id"
                            link
                            type="primary"
                            icon="Position"
                            :title="$t('firewall.quickJump')"
                            :aria-label="$t('firewall.quickJump')"
                            :disabled="loading || row.busy"
                            @click="routerToFileWithPath(row.savedPath)"
                        />
                    </div>
                </div>
            </div>
            <el-button
                class="directory-add"
                link
                type="primary"
                icon="Plus"
                :disabled="loading || !loaded"
                @click="add"
            >
                {{ $t('commons.button.add') }}
            </el-button>
        </div>
        <div class="directory-footer">
            <el-button @click="close">{{ $t('commons.button.close') }}</el-button>
        </div>
    </el-popover>
    <FileList ref="fileRef" @choose="loadPath" @close="selectingRow = undefined" />
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { routerToFileWithPath } from '@/utils/router';
import FileList from '@/components/file-list/index.vue';
import { Dashboard } from '@/api/interface/dashboard';
import {
    loadQuickDirectories,
    createQuickDirectory,
    updateQuickDirectory,
    deleteQuickDirectory,
} from '@/api/modules/dashboard';

defineProps<{ count: string; disabled: boolean }>();
const emit = defineEmits<{ (e: 'change', count: number): void }>();
type DirectoryRow = Dashboard.QuickDirectory & {
    key: number;
    savedPath: string;
    editing: boolean;
    busy: boolean;
};
const rows = ref<DirectoryRow[]>([]);
const searchInput = ref('');
const searchTerm = ref('');
const filteredRows = computed(() =>
    rows.value.filter(
        (row) =>
            !row.id ||
            row.editing ||
            row.path.toLowerCase().includes(searchTerm.value) ||
            (row.alias ?? '').toLowerCase().includes(searchTerm.value),
    ),
);
const search = () => {
    searchTerm.value = searchInput.value.trim().toLowerCase();
};
const fileRef = ref<InstanceType<typeof FileList>>();
const selectingRow = ref<DirectoryRow>();
const visible = ref(false);
const loading = ref(false);
const loaded = ref(false);
let nextKey = 0;

const setVisible = (value: boolean) => {
    if (value) visible.value = true;
};
const close = () => {
    visible.value = false;
};

const selectPath = (row: DirectoryRow) => {
    if (!row.editing || row.busy) return;
    selectingRow.value = row;
    fileRef.value?.acceptParams({ path: row.path, dir: false, isAll: true });
};
const loadPath = (path: string) => {
    const row = selectingRow.value;
    if (row?.editing && !row.busy) row.path = path;
};

const add = () => {
    rows.value.push({
        key: nextKey++,
        id: 0,
        path: '',
        alias: '',
        savedPath: '',
        editing: true,
        busy: false,
    });
};
const updateCount = () => emit('change', rows.value.filter((row) => row.id).length);
const load = async () => {
    if (loading.value || rows.value.some((row) => row.busy)) return;
    loading.value = true;
    loaded.value = false;
    searchInput.value = '';
    searchTerm.value = '';
    rows.value = [];
    try {
        const { data } = await loadQuickDirectories();
        rows.value = data.map((item) => ({
            ...item,
            alias: item.alias ?? '',
            key: nextKey++,
            savedPath: item.path,
            editing: false,
            busy: false,
        }));
        if (!rows.value.length) add();
        loaded.value = true;
        updateCount();
    } finally {
        loading.value = false;
    }
};
const save = async (row: DirectoryRow) => {
    if (row.busy || !row.path.trim()) return;
    row.busy = true;
    try {
        const payload = { path: row.path, alias: (row.alias ?? '').trim() };
        const { data } = row.id
            ? await updateQuickDirectory({ id: row.id, legacy: row.legacy ?? false, ...payload })
            : await createQuickDirectory(payload);
        Object.assign(row, data, { savedPath: data.path, editing: false });
        updateCount();
    } finally {
        row.busy = false;
    }
};
const remove = async (row: DirectoryRow) => {
    if (row.busy) return;
    if (row.id) {
        row.busy = true;
        try {
            await ElMessageBox.confirm('', i18n.global.t('home.quickDirDelete', [row.savedPath]), {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                customClass: 'quick-directory-delete',
                type: 'warning',
                center: true,
            });
        } catch {
            row.busy = false;
            return;
        }
        try {
            await deleteQuickDirectory(row.id, row.legacy ?? false);
        } catch (error) {
            row.busy = false;
            throw error;
        }
    }
    rows.value = rows.value.filter((item) => item.key !== row.key);
    if (!rows.value.length) add();
    updateCount();
};
</script>

<style lang="scss" scoped>
.quick-directories {
    padding: 12px;
    background: var(--el-fill-color-light);
    border-radius: 6px;
}
.directory-search {
    display: flex;
    gap: 12px;
    padding: 0 10px 10px;
}
.directory-empty {
    padding: 20px;
    text-align: center;
    color: var(--el-text-color-secondary);
}
.directory-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 96px;
    gap: 12px;
    align-items: center;
    padding: 10px;
}
.directory-header {
    font-weight: 600;
}
.directory-list {
    max-height: 360px;
    overflow-y: auto;
}
.directory-actions {
    display: flex;
    align-items: center;
    :deep(.el-button + .el-button) {
        margin-left: 8px;
    }
}
.directory-footer {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
}
.directory-add {
    width: 100%;
    margin-top: 8px;
    padding: 12px;
}
</style>
<style lang="scss">
.quick-directory-delete {
    .el-message-box__title {
        justify-content: center;
        gap: 10px;
        padding: 14px 16px;
        border-radius: 4px;
        background-color: #59472a;
        color: #f5cf87;
        overflow-wrap: anywhere;
    }
    .el-message-box__status {
        flex-shrink: 0;
        font-size: 24px;
        color: #f2b329;
    }
    .el-message-box__content {
        display: none;
    }
}
</style>
