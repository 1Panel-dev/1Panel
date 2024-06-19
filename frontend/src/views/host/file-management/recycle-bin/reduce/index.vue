<template>
    <el-dialog v-model="open" :title="$t('file.reduce')" width="30%" :close-on-click-modal="false">
        <el-row>
            <el-col :span="20" :offset="2">
                <el-alert :title="$t('file.confirmReduce')" show-icon type="error" :closable="false"></el-alert>
                <div class="flx-align-center mb-1 mt-1" v-for="(row, index) in files" :key="index">
                    <div>
                        <svg-icon v-if="row.isDir" className="table-icon mr-1 " iconName="p-file-folder"></svg-icon>
                        <svg-icon v-else className="table-icon mr-1" :iconName="getIconName(row.extension)"></svg-icon>
                    </div>
                    <span class="sle">{{ row.name }}</span>
                </div>
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onConfirm" v-loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { getIcon } from '@/utils/util';
import { reduceFile } from '@/api/modules/files';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { File } from '@/api/interface/file';

const open = ref(false);
const em = defineEmits(['close']);

const files = ref();
const loading = ref(false);
const forceDelete = ref(false);

const acceptParams = (props: File.RecycleBin[]) => {
    files.value = props;
    open.value = true;
    forceDelete.value = false;
};

const onConfirm = () => {
    const pros = [];
    for (const s of files.value) {
        pros.push(reduceFile({ from: s.from, rName: s.rName, name: s.name }));
    }
    loading.value = true;
    Promise.all(pros)
        .then(() => {
            MsgSuccess(i18n.global.t('file.reduceSuccess'));
            open.value = false;
            em('close');
        })
        .finally(() => {
            loading.value = false;
        });
};

const getIconName = (extension: string) => {
    return getIcon(extension);
};

defineExpose({
    acceptParams,
});
</script>
