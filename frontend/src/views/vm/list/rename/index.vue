<template>
    <DialogPro v-model="visible" :title="$t('xpack.vm.rename') + ' - ' + (source?.name || '')" size="small">
        <el-form label-position="top" :model="form">
            <el-form-item :label="$t('xpack.vm.newName')">
                <el-input v-model.trim="form.newName" />
                <span class="input-help">{{ $t('xpack.vm.renameHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" :loading="loading" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { newUUID } from '@/utils/id';
import { MsgError, MsgSuccess } from '@/utils/message';
import { VM } from '@/api/interface/vm';
import { renameVM } from '@/api/modules/vm';

const emit = defineEmits<{
    (e: 'task-success', taskID: string): void;
}>();

const visible = ref(false);
const loading = ref(false);
const source = ref<VM.Item>();
const form = reactive({
    newName: '',
});

const open = (row: VM.Item) => {
    source.value = row;
    form.newName = row.name;
    visible.value = true;
};

const submit = async () => {
    if (!source.value) {
        return;
    }
    if (!form.newName || form.newName === source.value.name) {
        MsgError(i18n.global.t('xpack.vm.renameHelper'));
        return;
    }
    await ElMessageBox.confirm(
        i18n.global.t('xpack.vm.renameConfirm', [source.value.name, form.newName]),
        i18n.global.t('xpack.vm.rename'),
    );
    const taskID = newUUID();
    loading.value = true;
    await renameVM({
        taskID,
        name: source.value.name,
        newName: form.newName,
    })
        .then((res) => {
            visible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('task-success', res.data?.taskID || taskID);
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({ open });
</script>
