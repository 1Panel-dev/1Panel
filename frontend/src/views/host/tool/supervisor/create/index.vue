<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="30%">
        <template #header>
            <DrawerHeader :header="$t('commons.button.create')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="processForm" label-position="top" :model="process" label-width="100px" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="process.name"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('tool.supervisor.user')" prop="user">
                        <el-input v-model.trim="process.user"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('tool.supervisor.dir')" prop="dir">
                        <el-input v-model.trim="process.dir">
                            <template #prepend><FileList @choose="getPath" :dir="true"></FileList></template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('tool.supervisor.command')" prop="command">
                        <el-input v-model.trim="process.command"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('tool.supervisor.numprocs')" prop="numprocs">
                        <el-input v-model.trim="process.numprocs"></el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(processForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { OperateSupervisorProcess } from '@/api/modules/host-tool';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const loading = ref(false);
const processForm = ref<FormInstance>();
const rules = ref({
    name: [Rules.requiredInput],
    dir: [Rules.requiredInput],
    command: [Rules.requiredInput],
    user: [Rules.requiredInput],
    numprocs: [Rules.requiredInput],
});
const process = ref({
    operate: 'create',
    name: '',
    command: '',
    user: '',
    dir: '',
    numprocs: '1',
});

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    resetForm();
    em('close', open);
};

const getPath = (path: string) => {
    process.value.dir = path;
};

const resetForm = () => {
    processForm.value?.resetFields();
};

const acceptParams = () => {
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        OperateSupervisorProcess(process.value)
            .then(() => {
                open.value = false;
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
