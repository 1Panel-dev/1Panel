<template>
    <el-dialog v-model="open" :title="$t('app.delete')" width="30%" :close-on-click-modal="false">
        <el-row>
            <el-col :span="20" :offset="2">
                <el-alert :title="$t('file.deleteHelper')" type="error" effect="dark" :closable="false"></el-alert>
                <div class="resource">
                    <table>
                        <tr v-for="(row, index) in files" :key="index">
                            <td>
                                <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                                <svg-icon
                                    v-else
                                    className="table-icon"
                                    :iconName="getIconName(row.extension)"
                                ></svg-icon>
                                <span>{{ row.name }}</span>
                            </td>
                        </tr>
                    </table>
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
<script lang="ts" setup>
import i18n from '@/lang';
import { ref } from 'vue';
import { File } from '@/api/interface/file';
import { getIcon } from '@/utils/util';
import { DeleteFile } from '@/api/modules/files';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const files = ref();
const loading = ref(false);
const em = defineEmits(['close']);

const acceptParams = (props: File.File[]) => {
    files.value = props;
    open.value = true;
};

const onConfirm = () => {
    const pros = [];
    for (const s of files.value) {
        pros.push(DeleteFile({ path: s['path'], isDir: s['isDir'] }));
    }
    loading.value = true;
    Promise.all(pros)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
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

<style scoped>
.resource {
    margin-top: 10px;
    max-height: 400px;
    overflow: auto;
}
</style>
