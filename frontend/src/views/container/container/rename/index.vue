<template>
    <el-drawer v-model="newNameVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('container.rename')" :resource="renameForm.name" :back="handleClose" />
        </template>
        <el-form @submit.prevent ref="newNameRef" v-loading="loading" :model="renameForm" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('container.newName')" :rules="Rules.volumeName" prop="newName">
                        <el-input v-model="renameForm.newName"></el-input>
                    </el-form-item>
                </el-col>
            </el-row>
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
    </el-drawer>
</template>

<script lang="ts" setup>
import { containerOperator } from '@/api/modules/container';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElForm } from 'element-plus';
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';

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
        await containerOperator(renameForm)
            .then(() => {
                loading.value = false;
                emit('search');
                newNameVisiable.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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
    renameForm.name = props.container;
    renameForm.newName = '';
    newNameVisiable.value = true;
};

const handleClose = async () => {
    newNameVisiable.value = false;
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
