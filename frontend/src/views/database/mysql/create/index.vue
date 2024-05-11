<template>
    <el-drawer
        v-model="createVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="30%"
    >
        <template #header>
            <DrawerHeader :header="$t('database.create')" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.name')" prop="name">
                            <el-input clearable v-model.trim="form.name" @input="form.username = form.name">
                                <template #append>
                                    <el-select v-model="form.format" style="width: 120px">
                                        <el-option label="utf8mb4" value="utf8mb4" />
                                        <el-option label="utf-8" value="utf8" />
                                        <el-option label="gbk" value="gbk" />
                                        <el-option label="big5" value="big5" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.username')" prop="username">
                            <el-input clearable v-model.trim="form.username" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.password')" prop="password">
                            <el-input type="password" clearable show-password v-model.trim="form.password">
                                <template #append>
                                    <el-button @click="random">{{ $t('commons.button.random') }}</el-button>
                                </template>
                            </el-input>
                        </el-form-item>

                        <el-form-item :label="$t('database.permission')" prop="permission">
                            <el-select v-model="form.permission">
                                <el-option value="%" :label="$t('database.permissionAll')" />
                                <el-option
                                    v-if="form.from !== 'local'"
                                    value="localhost"
                                    :label="$t('terminal.localhost')"
                                />
                                <el-option value="ip" :label="$t('database.permissionForIP')" />
                            </el-select>
                        </el-form-item>
                        <el-form-item v-if="form.permission === 'ip'" prop="permissionIPs">
                            <el-input clearable :rows="3" type="textarea" v-model="form.permissionIPs" />
                            <span class="input-help">{{ $t('database.remoteHelper') }}</span>
                        </el-form-item>

                        <el-form-item :label="$t('commons.table.type')" prop="database">
                            <el-tag>{{ form.database + ' [' + form.type + ']' }}</el-tag>
                        </el-form-item>

                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input type="textarea" clearable v-model="form.description" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
        </div>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="createVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { addMysqlDB } from '@/api/modules/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { checkIp, getRandomStr } from '@/utils/util';

const loading = ref();
const createVisible = ref(false);
const form = reactive({
    name: '',
    from: 'local',
    type: '',
    database: '',
    format: '',
    username: '',
    password: '',
    permission: '',
    permissionIPs: '',
    description: '',
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.dbName],
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.paramComplexity],
    permission: [Rules.requiredSelect],
    permissionIPs: [{ validator: checkIPs, trigger: 'blur', required: true }],
});
function checkIPs(rule: any, value: any, callback: any) {
    let ips = form.permissionIPs.split(',');
    for (const item of ips) {
        if (checkIp(item)) {
            return callback(new Error(i18n.global.t('commons.rule.ip')));
        }
    }
    callback();
}

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    from: string;
    type: string;
    database: string;
}
const acceptParams = (params: DialogProps): void => {
    form.name = '';
    form.from = params.from;
    form.type = params.type;
    form.database = params.database;
    form.format = 'utf8mb4';
    form.username = '';
    form.permission = '%';
    form.permissionIPs = '';
    form.description = '';
    random();
    createVisible.value = true;
};
const handleClose = () => {
    createVisible.value = false;
};

const random = async () => {
    form.password = getRandomStr(16);
};

const emit = defineEmits<{ (e: 'search'): void }>();
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.permission === 'ip') {
            form.permission = form.permissionIPs;
        }
        loading.value = true;
        await addMysqlDB(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                createVisible.value = false;
            })
            .catch(() => {
                if (form.permission != '%') {
                    form.permissionIPs = form.permission;
                    form.permission = 'ip';
                }
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
