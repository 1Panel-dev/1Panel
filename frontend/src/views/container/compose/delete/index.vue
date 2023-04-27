<template>
    <el-dialog
        v-model="dialogVisiable"
        :title="$t('commons.button.delete') + ' - ' + composeName"
        width="30%"
        :close-on-click-modal="false"
    >
        <el-form ref="deleteForm" v-loading="loading">
            <el-form-item>
                <el-checkbox v-model="deleteFile" :label="$t('container.deleteFile')" />
                <span class="input-help">
                    {{ $t('container.deleteComposeHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <div>
                    <span style="font-size: 12px">{{ $t('database.delete') }}</span>
                    <span style="font-size: 12px; color: red; font-weight: 500">{{ composeName }}</span>
                    <span style="font-size: 12px">{{ $t('container.deleteCompose') }}</span>
                </div>
                <el-input v-model="deleteInfo" :placeholder="composeName"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisiable = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="deleteInfo != composeName || loading">
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
import { composeOperator } from '@/api/modules/container';

let dialogVisiable = ref(false);
let loading = ref(false);
let deleteInfo = ref('');

const deleteFile = ref();
const composeName = ref();
const composePath = ref();

const deleteForm = ref<FormInstance>();

interface DialogProps {
    name: string;
    path: string;
}
const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = async (prop: DialogProps) => {
    deleteFile.value = false;
    composeName.value = prop.name;
    composePath.value = prop.path;
    deleteInfo.value = '';
    dialogVisiable.value = true;
};

const submit = async () => {
    loading.value = true;
    let params = {
        name: composeName.value,
        path: composePath.value,
        operation: 'down',
        withFile: deleteFile.value,
    };
    await composeOperator(params)
        .then(() => {
            loading.value = false;
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            dialogVisiable.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
