<template>
    <el-drawer v-model="open" width="30%" :close-on-click-modal="false" :close-on-press-escape="false">
        <template #header>
            <DrawerHeader :header="$t('file.info')" :back="handleClose" />
        </template>
        <el-descriptions :column="1" border>
            <el-descriptions-item label-class-name="detail-label" :label="$t('file.fileName')">
                {{ data.name }}
            </el-descriptions-item>
            <el-descriptions-item
                label-class-name="detail-label"
                :label="$t('commons.table.type')"
                v-if="data.type != ''"
            >
                {{ data.type }}
            </el-descriptions-item>
            <el-descriptions-item class-name="detail-content" label-class-name="detail-label" :label="$t('file.path')">
                {{ data.path }}
            </el-descriptions-item>
            <el-descriptions-item label-class-name="detail-label" :label="$t('file.size')">
                <span v-if="data.isDir">
                    <el-button type="primary" link small @click="getDirSize(data)" :loading="loading">
                        <span v-if="data.dirSize == undefined">
                            {{ $t('file.calculate') }}
                        </span>
                        <span v-else>{{ computeSize(data.dirSize) }}</span>
                    </el-button>
                </span>
                <span v-else>{{ computeSize(data.size) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label-class-name="detail-label" :label="$t('file.role')">
                {{ data.mode }}
            </el-descriptions-item>
            <el-descriptions-item label-class-name="detail-label" :label="$t('commons.table.user')">
                {{ data.user }}
            </el-descriptions-item>
            <el-descriptions-item label-class-name="detail-label" :label="$t('file.group')">
                {{ data.group }}
            </el-descriptions-item>
            <el-descriptions-item label-class-name="detail-label" :label="$t('commons.table.updatedAt')">
                {{ dateFormatSimple(data.modTime) }}
            </el-descriptions-item>
        </el-descriptions>
    </el-drawer>
</template>

<script lang="ts" setup>
import { ComputeDirSize, GetFileContent } from '@/api/modules/files';
import { computeSize } from '@/utils/util';
import { ref } from 'vue';
import { dateFormatSimple } from '@/utils/util';
import DrawerHeader from '@/components/drawer-header/index.vue';

interface InfoProps {
    path: string;
}
const props = ref<InfoProps>({
    path: '',
});

let open = ref(false);
let data = ref();
let loading = ref(false);

const handleClose = () => {
    open.value = false;
};

const acceptParams = async (params: InfoProps): Promise<void> => {
    props.value = params;
    GetFileContent({ path: params.path, expand: false, page: 1, pageSize: 1, isDetail: true }).then((res) => {
        data.value = res.data;
        open.value = true;
    });
};

const getDirSize = async (row: any) => {
    const req = {
        path: row.path,
    };
    loading.value = true;
    await ComputeDirSize(req)
        .then(async (res) => {
            data.value.dirSize = res.data.size;
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
<style scoped>
:deep(.detail-label) {
    min-width: 100px !important;
}

:deep(.detail-content) {
    max-width: 295px;
    word-break: break-all;
    word-wrap: break-word;
}
</style>
