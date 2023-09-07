<template>
    <el-dialog v-model="open" :title="$t('app.delete')" width="30%" :close-on-click-modal="false">
        <el-row>
            <el-col :span="20" :offset="2">
                <el-alert :title="msg" show-icon type="error" :closable="false"></el-alert>
                <div class="resource">
                    <table>
                        <tr v-for="(row, index) in containers" :key="index">
                            <td>
                                <span>{{ row }}</span>
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
import { containerOperator } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const containers = ref();
const msg = ref();
const operation = ref();
const loading = ref(false);
const em = defineEmits(['search']);

interface DialogProps {
    containers: Array<string>;
    operation: string;
    msg: string;
}

const acceptParams = (props: DialogProps) => {
    containers.value = props.containers;
    operation.value = props.operation;
    msg.value = props.msg;
    open.value = true;
};

const onConfirm = () => {
    const pros = [];
    for (const item of containers.value) {
        pros.push(containerOperator({ name: item, operation: operation.value, newName: '' }));
    }
    loading.value = true;
    Promise.all(pros)
        .then(() => {
            open.value = false;
            loading.value = false;
            em('search');
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        })
        .finally(() => {
            open.value = false;
            loading.value = false;
            em('search');
        });
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
