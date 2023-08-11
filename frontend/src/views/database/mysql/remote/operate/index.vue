<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader
                :hideResource="dialogData.title === 'create'"
                :header="title"
                :resource="dialogData.rowData?.name"
                :back="handleClose"
            />
        </template>
        <el-form ref="formRef" v-loading="loading" label-position="top" :model="dialogData.rowData" :rules="rules">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input
                            v-if="dialogData.title === 'create'"
                            clearable
                            v-model.trim="dialogData.rowData!.name"
                        />
                        <el-tag v-else>{{ dialogData.rowData!.name }}</el-tag>
                    </el-form-item>
                    <el-form-item :label="$t('database.version')" prop="version">
                        <el-select @change="isOK = false" v-model="dialogData.rowData!.version">
                            <el-option value="5.6" label="5.6" />
                            <el-option value="5.7" label="5.7" />
                            <el-option value="8.0" label="8.0" />
                        </el-select>
                        <span class="input-help">{{ $t('database.versionHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('database.address')" prop="address">
                        <el-input @change="isOK = false" clearable v-model.trim="dialogData.rowData!.address" />
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.port')" prop="port">
                        <el-input @change="isOK = false" clearable v-model.number="dialogData.rowData!.port" />
                    </el-form-item>
                    <el-form-item :label="$t('commons.login.username')" prop="username">
                        <el-input @change="isOK = false" clearable v-model.trim="dialogData.rowData!.username" />
                        <span class="input-help">{{ $t('database.userHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('commons.login.password')" prop="password">
                        <el-input
                            @change="isOK = false"
                            type="password"
                            clearable
                            show-password
                            v-model.trim="dialogData.rowData!.password"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.description')" prop="description">
                        <el-input clearable v-model.trim="dialogData.rowData!.description" />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button @click="onSubmit(formRef, 'check')">
                    {{ $t('terminal.testConn') }}
                </el-button>
                <el-button type="primary" :disabled="!isOK" @click="onSubmit(formRef, dialogData.title)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Database } from '@/api/interface/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import { addRemoteDB, checkRemoteDB, editRemoteDB } from '@/api/modules/database';

interface DialogProps {
    title: string;
    rowData?: Database.RemoteDBInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisiable = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const isOK = ref(false);
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('database.' + dialogData.value.title + 'RemoteDB');
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput],
    version: [Rules.requiredSelect],
    address: [Rules.host],
    port: [Rules.port],
    username: [Rules.requiredInput],
    password: [Rules.requiredInput],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined, operation: string) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            id: dialogData.value.rowData.id,
            name: dialogData.value.rowData.name,
            type: 'mysql',
            version: dialogData.value.rowData.version,
            from: 'remote',
            address: dialogData.value.rowData.address,
            port: dialogData.value.rowData.port,
            username: dialogData.value.rowData.username,
            password: dialogData.value.rowData.password,
            description: dialogData.value.rowData.description,
        };
        loading.value = true;

        if (operation === 'check') {
            await checkRemoteDB(param)
                .then((res) => {
                    loading.value = false;
                    if (res.data) {
                        isOK.value = true;
                        MsgSuccess(i18n.global.t('terminal.connTestOk'));
                    } else {
                        MsgError(i18n.global.t('terminal.connTestFailed'));
                    }
                })
                .catch(() => {
                    loading.value = false;
                    MsgError(i18n.global.t('terminal.connTestFailed'));
                });
        }

        if (operation === 'create') {
            await addRemoteDB(param)
                .then(() => {
                    loading.value = true;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisiable.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
        }
        if (operation === 'edit') {
            await editRemoteDB(param)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisiable.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
