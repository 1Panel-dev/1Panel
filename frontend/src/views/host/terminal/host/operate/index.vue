<template>
    <div v-loading="loading">
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="$t('terminal.host')" :back="handleClose" />
            </template>
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form ref="hostInfoRef" label-position="top" :model="dialogData.rowData" :rules="rules">
                        <el-form-item :label="$t('terminal.ip')" prop="addr">
                            <el-tag v-if="dialogData.rowData!.addr === '127.0.0.1' && dialogData.title === 'edit'">
                                {{ dialogData.rowData!.addr }}
                            </el-tag>
                            <el-input @change="isOK = false" v-else clearable v-model.trim="dialogData.rowData!.addr" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.username')" prop="user">
                            <el-input @change="isOK = false" clearable v-model="dialogData.rowData!.user" />
                        </el-form-item>
                        <el-form-item :label="$t('terminal.authMode')" prop="authMode">
                            <el-radio-group @change="isOK = false" v-model="dialogData.rowData!.authMode">
                                <el-radio value="password">{{ $t('terminal.passwordMode') }}</el-radio>
                                <el-radio value="key">{{ $t('terminal.keyMode') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item
                            :label="$t('commons.login.password')"
                            v-if="dialogData.rowData!.authMode === 'password'"
                            prop="password"
                        >
                            <el-input
                                @change="isOK = false"
                                clearable
                                show-password
                                type="password"
                                v-model="dialogData.rowData!.password"
                            />
                        </el-form-item>
                        <el-form-item
                            :label="$t('terminal.key')"
                            v-if="dialogData.rowData!.authMode === 'key'"
                            prop="privateKey"
                        >
                            <el-input
                                @change="isOK = false"
                                clearable
                                type="textarea"
                                v-model="dialogData.rowData!.privateKey"
                            />
                        </el-form-item>
                        <el-form-item
                            :label="$t('terminal.keyPassword')"
                            v-if="dialogData.rowData!.authMode === 'key'"
                            prop="passPhrase"
                        >
                            <el-input
                                @change="isOK = false"
                                type="password"
                                show-password
                                clearable
                                v-model="dialogData.rowData!.passPhrase"
                            />
                        </el-form-item>
                        <el-checkbox clearable v-model.number="dialogData.rowData!.rememberPassword">
                            {{ $t('terminal.rememberPassword') }}
                        </el-checkbox>
                        <el-form-item style="margin-top: 10px" :label="$t('commons.table.port')" prop="port">
                            <el-input @change="isOK = false" clearable v-model.number="dialogData.rowData!.port" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.group')" prop="groupID">
                            <el-select filterable v-model="dialogData.rowData!.groupID" clearable style="width: 100%">
                                <el-option
                                    v-for="item in groupList"
                                    :key="item.id"
                                    :label="item.name"
                                    :value="item.id"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.title')" prop="name">
                            <el-input clearable v-model="dialogData.rowData!.name" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input clearable type="textarea" v-model="dialogData.rowData!.description" />
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button @click="submitAddHost(hostInfoRef, 'testconn')">
                        {{ $t('terminal.testConn') }}
                    </el-button>
                    <el-button type="primary" :disabled="!isOK" @click="submitAddHost(hostInfoRef, dialogData.title)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { addHost, editHost, testByInfo } from '@/api/modules/host';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GetGroupList } from '@/api/modules/group';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';

const loading = ref();
const isOK = ref(false);
interface DialogProps {
    title: string;
    rowData?: any;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});

const groupList = ref();
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisible.value = true;
    loadGroups();
};
const emit = defineEmits<{ (e: 'search'): void }>();
const handleClose = () => {
    drawerVisible.value = false;
};

type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    groupID: [Rules.requiredSelect],
    addr: [Rules.ipV4V6OrDomain],
    port: [Rules.requiredInput, Rules.port],
    user: [Rules.requiredInput],
    authMode: [Rules.requiredSelect],
});

const loadGroups = async () => {
    const res = await GetGroupList({ type: 'host' });
    groupList.value = res.data;
    if (dialogData.value.title === 'create') {
        for (const item of groupList.value) {
            if (item.isDefault) {
                dialogData.value.rowData.groupID = item.id;
                break;
            }
        }
    }
};

const submitAddHost = (formEl: FormInstance | undefined, ops: string) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (ops === 'create') {
            loading.value = true;
            await addHost(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        }
        if (ops === 'edit') {
            loading.value = true;
            await editHost(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        }
        if (ops === 'testconn') {
            loading.value = true;
            await testByInfo(dialogData.value.rowData).then((res) => {
                loading.value = false;
                if (res.data) {
                    isOK.value = true;
                    MsgSuccess(i18n.global.t('terminal.connTestOk'));
                } else {
                    isOK.value = false;
                    MsgError(i18n.global.t('terminal.connTestFailed'));
                }
            });
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
