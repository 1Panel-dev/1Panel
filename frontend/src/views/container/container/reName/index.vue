<template>
    <el-dialog
        @close="onClose()"
        v-model="newNameVisiable"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="30%"
    >
        <template #header>
            <div class="card-header">
                <span>{{ $t('container.rename') }}</span>
            </div>
        </template>
        <el-form ref="newNameRef" v-loading="loading" :model="renameForm">
            <el-form-item :label="$t('container.newName')" :rules="Rules.requiredInput" prop="newName">
                <el-input v-model="renameForm.newName"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="newNameVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmitName(newNameRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ContainerOperator } from '@/api/modules/container';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { reactive, ref } from 'vue';

const loading = ref(false);

const renameForm = reactive({
    name: '',
    operation: 'rename',
    newName: '',
});

const newNameRef = ref<FormInstance>();

const newNameVisiable = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;

const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmitName = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await ContainerOperator(renameForm)
            .then(() => {
                loading.value = false;
                emit('search');
                newNameVisiable.value = false;
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

interface DialogProps {
    container: string;
}

const acceptParams = (props: DialogProps): void => {
    console.log(props.container);
    renameForm.name = props.container;
    renameForm.newName = '';
    newNameVisiable.value = true;
};

const onClose = async () => {
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
