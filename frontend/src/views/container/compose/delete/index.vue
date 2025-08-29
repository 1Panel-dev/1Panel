<template>
    <DialogPro v-model="open" :title="$t('commons.button.delete') + ' - ' + composeName" size="small">
        <el-form ref="deleteForm" v-loading="loading">
            <el-form-item>
                <el-checkbox v-model="deleteFile" :label="$t('container.deleteFile')" />
                <span class="input-help whitespace-break-spaces">
                    {{ $t('container.deleteComposeHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="force" :label="$t('website.forceDelete')" />
                <span class="input-help">
                    {{ $t('website.forceDeleteHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <div class="font">
                    <span>{{ $t('database.delete') }}</span>
                    <span class="warning">{{ composeName }}</span>
                    <span>{{ $t('container.deleteCompose') }}</span>
                </div>
                <el-input v-model="deleteInfo" :placeholder="composeName"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="deleteInfo != composeName || loading">
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
import { composeOperator } from '@/api/modules/container';

let open = ref(false);
let loading = ref(false);
let deleteInfo = ref('');

const deleteFile = ref();
const force = ref();
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
    force.value = false;
    composeName.value = prop.name;
    composePath.value = prop.path;
    deleteInfo.value = '';
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    let params = {
        name: composeName.value,
        path: composePath.value,
        operation: 'delete',
        withFile: deleteFile.value,
        force: force.value,
    };
    await composeOperator(params)
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

<style lang="scss" scoped>
.font {
    font-size: 12px;
    .warning {
        color: red;
        font-weight: 500;
    }
}
</style>
