<template>
    <DrawerPro
        v-model="open"
        :header="process.operate == 'create' ? $t('commons.button.create') : $t('commons.button.edit')"
        :back="handleClose"
        size="small"
    >
        <el-form
            ref="processForm"
            label-position="top"
            :model="process"
            label-width="100px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="process.name" :disabled="process.operate == 'update'"></el-input>
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
                <el-input v-model="process.command"></el-input>
            </el-form-item>
            <el-form-item :label="$t('tool.supervisor.numprocs')" prop="numprocsNum">
                <el-input type="number" v-model.number="process.numprocsNum"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(processForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { createSupervisorProcess } from '@/api/modules/runtime';
import { Rules, checkNumberRange } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { HostTool } from '@/api/interface/host-tool';

const open = ref(false);
const loading = ref(false);
const processForm = ref<FormInstance>();
const rules = ref({
    name: [Rules.requiredInput, Rules.supervisorName],
    dir: [Rules.requiredInput],
    command: [Rules.requiredInput],
    user: [Rules.requiredInput],
    numprocsNum: [Rules.requiredInput, Rules.integerNumber, checkNumberRange(1, 9999)],
});
const initData = (runtimeID: number) => ({
    operate: 'create',
    name: '',
    command: '',
    user: 'www-data',
    dir: '',
    numprocsNum: 1,
    numprocs: '1',
    id: runtimeID,
});
const process = ref(initData(0));
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
    process.value = initData(0);
    processForm.value?.resetFields();
};

const acceptParams = (operate: string, config: HostTool.SupersivorProcess, id: number) => {
    process.value = initData(id);
    if (operate == 'update') {
        process.value = {
            operate: 'update',
            name: config.name,
            command: config.command,
            user: config.user,
            dir: config.dir,
            numprocsNum: 1,
            numprocs: config.numprocs,
            id: id,
        };
        process.value.numprocsNum = Number(config.numprocs);
    }
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        process.value.numprocs = String(process.value.numprocsNum);
        createSupervisorProcess(process.value)
            .then(() => {
                open.value = false;
                em('close', open);
                MsgSuccess(i18n.global.t('commons.msg.' + process.value.operate + 'Success'));
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
