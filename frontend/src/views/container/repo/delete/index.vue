<template>
    <el-dialog v-model="repoVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('commons.button.delete') }}</span>
            </div>
        </template>
        <el-form v-loading="loading" label-width="20px">
            <el-form-item>
                <el-checkbox v-model="isDelete">{{ $t('container.delInsecure') }}</el-checkbox>
                <span class="input-help">{{ $t('container.delInsecureHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="repoVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { deleteImageRepo } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const isDelete = ref(false);

interface DialogProps {
    ids?: Array<number>;
}
const repoVisiable = ref(false);
const ids = ref();

const acceptParams = (params: DialogProps): void => {
    ids.value = params.ids;
    repoVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmit = async () => {
    loading.value = true;
    await deleteImageRepo({ ids: ids.value, deleteInsecure: isDelete.value })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            repoVisiable.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
    return;
};

defineExpose({
    acceptParams,
});
</script>
