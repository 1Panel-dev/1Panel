<template>
    <el-dialog v-model="dialogVisible" :title="$t('commons.button.delete')" width="30%" :close-on-click-modal="false">
        <el-form ref="deleteRef" v-loading="loading" @submit.prevent>
            <el-form-item>
                <el-alert :title="$t('license.deleteHelper')" :closable="false" type="warning" />
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.forceDelete" :label="$t('database.unBindForce')" />
                <span class="input-help">
                    {{ $t('license.forceDelete') }}
                </span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { DeleteLicense } from '@/api/modules/setting';

let form = reactive({
    id: 0,
    licenseName: '',
    forceDelete: false,
});
let dialogVisible = ref(false);
let loading = ref(false);

const deleteRef = ref<FormInstance>();

interface DialogProps {
    id: number;
    name: string;
    database: string;
}
const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = async (prop: DialogProps) => {
    form.id = prop.id;
    form.licenseName = prop.name;
    form.forceDelete = false;
    dialogVisible.value = true;
};

const submit = async () => {
    loading.value = true;
    DeleteLicense(form.id, form.forceDelete)
        .then(() => {
            loading.value = false;
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            dialogVisible.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
