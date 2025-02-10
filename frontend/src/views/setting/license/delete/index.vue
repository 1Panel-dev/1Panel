<template>
    <DialogPro v-model="open" :title="$t('commons.button.delete')" size="small">
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
                <el-button @click="open = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>
<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { deleteLicense } from '@/api/modules/setting';

let form = reactive({
    id: 0,
    licenseName: '',
    forceDelete: false,
});
let open = ref(false);
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
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    deleteLicense(form.id, form.forceDelete)
        .then(() => {
            loading.value = false;
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            open.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
