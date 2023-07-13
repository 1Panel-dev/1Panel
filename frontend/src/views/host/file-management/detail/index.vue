<template>
    <el-drawer v-model="open" width="30%">
        <template #header>
            <DrawerHeader :header="$t('file.info')" :back="handleClose" />
        </template>
        <el-row>
            <el-col>
                <el-descriptions :column="1" border>
                    <el-descriptions-item :label="$t('file.fileName')">{{ data.name }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.type')">{{ data.type }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('file.path')">{{ data.path }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('file.size')">
                        {{ computeSize(data.size) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('file.role')">{{ data.mode }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.user')">{{ data.user }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('file.group')">{{ data.group }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.updatedAt')">
                        {{ dateFormatSimple(data.modTime) }}
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </el-drawer>
</template>

<script lang="ts" setup>
import { GetFileContent } from '@/api/modules/files';
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

const handleClose = () => {
    open.value = false;
};

const acceptParams = async (params: InfoProps): Promise<void> => {
    props.value = params;
    GetFileContent({ path: params.path, expand: false, page: 1, pageSize: 1 }).then((res) => {
        data.value = res.data;
        open.value = true;
    });
};

defineExpose({
    acceptParams,
});
</script>
