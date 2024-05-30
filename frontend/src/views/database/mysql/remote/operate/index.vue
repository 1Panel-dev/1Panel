<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
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
                    <el-form-item :label="$t('commons.table.type')" prop="type">
                        <el-radio-group v-model="dialogData.rowData!.type" @change="changeType">
                            <el-radio-button value="mysql">MySQL</el-radio-button>
                            <el-radio-button value="mariadb">MariaDB</el-radio-button>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item :label="$t('database.version')" prop="version">
                        <el-radio-group v-model="dialogData.rowData!.version" @change="isOK = false">
                            <div v-if="dialogData.rowData!.type === 'mysql'">
                                <el-radio label="8.x" value="8.x" />
                                <el-radio label="5.7" value="5.7" />
                                <el-radio label="5.6" value="5.6" />
                            </div>
                            <div v-else>
                                <el-radio label="10.x" value="10.x" />
                                <el-radio label="11.x" value="11.x" />
                            </div>
                        </el-radio-group>
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
                    <el-form-item>
                        <el-checkbox
                            @change="isOK = false"
                            v-model="dialogData.rowData!.ssl"
                            :label="$t('database.ssl')"
                        />
                    </el-form-item>
                    <div v-if="dialogData.rowData!.ssl">
                        <el-form-item>
                            <el-checkbox
                                @change="isOK = false"
                                v-model="dialogData.rowData!.hasCA"
                                :label="$t('database.hasCA')"
                            />
                        </el-form-item>
                        <el-form-item>
                            <el-checkbox
                                @change="isOK = false"
                                v-model="dialogData.rowData!.skipVerify"
                                :label="$t('database.skipVerify')"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('database.clientKey')" prop="clientKey">
                            <el-input
                                type="textarea"
                                @change="isOK = false"
                                clearable
                                v-model="dialogData.rowData!.clientKey"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('database.clientCert')" prop="clientCert">
                            <el-input
                                type="textarea"
                                @change="isOK = false"
                                clearable
                                v-model="dialogData.rowData!.clientCert"
                            />
                        </el-form-item>
                        <el-form-item v-if="dialogData.rowData!.hasCA" :label="$t('database.caCert')" prop="rootCert">
                            <el-input
                                type="textarea"
                                @change="isOK = false"
                                clearable
                                v-model="dialogData.rowData!.rootCert"
                            />
                        </el-form-item>
                    </div>
                    <el-form-item :label="$t('commons.table.description')" prop="description">
                        <el-input clearable v-model.trim="dialogData.rowData!.description" />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { addDatabase, checkDatabase, editDatabase } from '@/api/modules/database';

interface DialogProps {
    title: string;
    rowData?: Database.DatabaseInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const isOK = ref(false);
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (dialogData.value.rowData.version.startsWith('5.6')) {
        dialogData.value.rowData.version = '5.6';
    }
    if (dialogData.value.rowData.version.startsWith('5.7')) {
        dialogData.value.rowData.version = '5.7';
    }
    if (dialogData.value.rowData.version.startsWith('8.')) {
        dialogData.value.rowData.version = '8.x';
    }
    if (dialogData.value.rowData.version.startsWith('10.')) {
        dialogData.value.rowData.version = '10.x';
    }
    dialogData.value.rowData.hasCA = dialogData.value.rowData.rootCert?.length !== 0;
    title.value = i18n.global.t('database.' + dialogData.value.title + 'RemoteDB');
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    name: [Rules.simpleName, Rules.noSpace],
    type: [Rules.requiredSelect],
    version: [Rules.requiredSelect],
    address: [Rules.ipV4V6OrDomain],
    port: [Rules.port],
    username: [Rules.requiredInput],
    password: [Rules.requiredInput],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const changeType = () => {
    dialogData.value.rowData.version = dialogData.value.rowData.type === 'mysql' ? '5.6' : '10.x';
    isOK.value = false;
};

const onSubmit = async (formEl: FormInstance | undefined, operation: string) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        dialogData.value.rowData.from = 'remote';
        loading.value = true;
        dialogData.value.rowData.rootCert = dialogData.value.rowData.hasCA ? dialogData.value.rowData.rootCert : '';
        if (operation === 'check') {
            await checkDatabase(dialogData.value.rowData)
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
            await addDatabase(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
        }
        if (operation === 'edit') {
            await editDatabase(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
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
