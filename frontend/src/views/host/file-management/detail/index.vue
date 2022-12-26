<template>
    <el-dialog v-model="open" :title="$t('file.info')" :column="1" width="30%">
        <el-row>
            <el-col>
                <el-descriptions :column="1" border>
                    <el-descriptions-item :label="$t('file.fileName')">{{ data.name }}</el-descriptions-item>
                    <!-- <el-descriptions-item :label="$t('file.type')">{{ data.type }}</el-descriptions-item> -->
                    <el-descriptions-item :label="$t('file.path')">{{ data.path }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('file.size')">{{ computeSize(data.size) }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('file.role')">{{ data.mode }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('file.user')">{{ data.user }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('file.group')">{{ data.group }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.updatedAt')">
                        {{ dateFromat(0, 0, data.modTime) }}
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </el-dialog>
</template>

<script lang="ts" setup>
import { GetFileContent } from '@/api/modules/files';
import { computeSize } from '@/utils/util';
import { ref } from 'vue';
import { dateFromat } from '@/utils/util';

interface InfoProps {
    path: string;
}
const props = ref<InfoProps>({
    path: '',
});

let open = ref(false);
let data = ref();

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
