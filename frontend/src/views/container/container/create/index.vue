<template>
    <el-dialog v-model="createVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>容器创建</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" label-position="left" :rules="rules" label-width="120px">
            <el-form-item label="容器名称" prop="name">
                <el-input clearable v-model="form.name" />
            </el-form-item>
            <el-form-item label="镜像" prop="image">
                <el-input clearable v-model="form.image" />
            </el-form-item>
            <el-form-item label="端口" prop="image">
                <el-radio-group v-model="form.publishAllPorts" class="ml-4">
                    <el-radio :label="false">暴露端口</el-radio>
                    <el-radio :label="true">暴露所有</el-radio>
                </el-radio-group>

                <div style="margin-top: 20px"></div>
                <table style="width: 100%; margin-top: 5px" class="tab-table">
                    <tr v-for="(row, index) in ports" :key="index">
                        <td width="48%">
                            <el-input v-model="row['key']" />
                        </td>
                        <td width="48%">
                            <el-input v-model="row['value']" />
                        </td>
                        <td>
                            <el-button type="text" style="font-size: 10px" @click="handlePortsDelete(index)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </td>
                    </tr>
                    <tr>
                        <td align="left">
                            <el-button @click="handlePortsAdd()">{{ $t('commons.button.add') }}</el-button>
                        </td>
                    </tr>
                </table>
            </el-form-item>
            <el-form-item label="启动命令" prop="command">
                <el-input clearable v-model="form.command" />
            </el-form-item>
            <el-form-item prop="autoRemove">
                <el-checkbox v-model="form.autoRemove">容器停止后自动删除容器</el-checkbox>
            </el-form-item>
            <el-form-item label="限制CPU" prop="cpusetCpus">
                <el-input v-model="form.cpusetCpus" />
            </el-form-item>
            <el-form-item label="内存" prop="memeryLimit">
                <el-input v-model="form.memeryLimit" />
            </el-form-item>
            <el-form-item label="挂载卷">
                <div style="margin-top: 20px"></div>
                <table style="width: 100%; margin-top: 5px" class="tab-table">
                    <tr v-for="(row, index) in volumes" :key="index">
                        <td width="30%">
                            <el-input v-model="row['name']" />
                        </td>
                        <td width="30%">
                            <el-input v-model="row['bind']" />
                        </td>
                        <td width="30%">
                            <el-input v-model="row['mode']" />
                        </td>
                        <td>
                            <el-button type="text" style="font-size: 10px" @click="handleVolumesDelete(index)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </td>
                    </tr>
                    <tr>
                        <td align="left">
                            <el-button @click="handleVolumesAdd()">{{ $t('commons.button.add') }}</el-button>
                        </td>
                    </tr>
                </table>
            </el-form-item>
            <el-form-item label="标签" prop="labels">
                <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" v-model="form.labels" />
            </el-form-item>
            <el-form-item label="环境变量(每行一个)" prop="environment">
                <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" v-model="form.environment" />
            </el-form-item>
            <el-form-item label="重启规则" prop="restartPolicy.value">
                <el-radio-group v-model="form.restartPolicy.value">
                    <el-radio :label="false">关闭后马上重启</el-radio>
                    <el-radio :label="false">错误时重启（默认重启 5 次）</el-radio>
                    <el-radio :label="true">不重启</el-radio>
                </el-radio-group>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="createVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';

const createVisiable = ref(false);
const form = reactive({
    name: '',
    image: '',
    command: '',
    publishAllPorts: false,
    ports: [],
    cpusetCpus: 1,
    memeryLimit: 100,
    volumes: [],
    autoRemove: false,
    labels: '',
    environment: '',
    restartPolicy: {
        value: '',
        name: '',
        maximumRetryCount: '',
    },
});
const ports = ref();
const volumes = ref();

const acceptParams = (): void => {
    createVisiable.value = true;
};

const emit = defineEmits<{ (e: 'search'): void }>();

const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    type: [Rules.requiredSelect],
    specType: [Rules.requiredSelect],
    week: [Rules.requiredSelect, Rules.number],
    day: [Rules.number, { max: 31, min: 1 }],
    hour: [Rules.number, { max: 23, min: 0 }],
    minute: [Rules.number, { max: 60, min: 1 }],

    script: [Rules.requiredInput],
    website: [Rules.requiredSelect],
    database: [Rules.requiredSelect],
    url: [Rules.requiredInput],
    sourceDir: [Rules.requiredSelect],
    targetDirID: [Rules.requiredSelect, Rules.number],
    retainCopies: [Rules.number],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const handlePortsAdd = () => {
    let item = {
        key: '',
        value: '',
    };
    ports.value.push(item);
};
const handlePortsDelete = (index: number) => {
    ports.value.splice(index, 1);
};

const handleVolumesAdd = () => {
    let item = {
        from: '',
        bind: '',
        mode: '',
    };
    volumes.value.push(item);
};
const handleVolumesDelete = (index: number) => {
    volumes.value.splice(index, 1);
};

function restForm() {
    if (formRef.value) {
        formRef.value.resetFields();
    }
}
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;

        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        restForm();
        emit('search');
        createVisiable.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
