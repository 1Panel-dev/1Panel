<template>
    <DrawerPro v-model="open" :header="$t('commons.button.ignore')" :resource="resourceName" @close="handleClose">
        <el-form @submit.prevent ref="updateRef" :rules="rules" label-position="top" :model="req">
            <el-form-item>
                <el-radio-group v-model="req.scope">
                    <el-radio-button :label="$t('app.ignoreAll')" value="all" />
                    <el-radio-button :label="$t('app.ignoreVersion')" value="version" />
                </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('app.versionSelect')" prop="appDetailID" v-if="req.scope === 'version'">
                <el-select v-model="req.appDetailID">
                    <el-option
                        v-for="(version, index) in versions"
                        :key="index"
                        :value="version.detailId"
                        :label="version.version"
                    ></el-option>
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { getAppUpdateVersions, ignoreUpgrade } from '@/api/modules/app';
import bus from '@/global/bus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const resourceName = ref('');
const req = reactive({
    appID: 0,
    appDetailID: 0,
    scope: 'all',
});
const versions = ref();
const appInstallID = ref(0);
const rules = ref<any>({
    appDetailID: [Rules.requiredSelect],
});
const loading = ref(false);
const emit = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
};

const acceptParams = (apppInstall: App.AppInstalled) => {
    appInstallID.value = apppInstall.id;
    req.appID = apppInstall.appID;
    getVersions();
    open.value = true;
};

const getVersions = async () => {
    try {
        const res = await getAppUpdateVersions({ appInstallID: appInstallID.value });
        versions.value = res.data || [];
        if (versions.value.length > 0) {
            req.appDetailID = versions.value[0].detailId;
        }
    } catch (error) {}
};

const submit = async () => {
    loading.value = true;
    try {
        await ignoreUpgrade(req);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        handleClose();
        bus.emit('upgrade', true);
        emit('close');
    } catch (error) {
        return;
    } finally {
        loading.value = false;
    }
};
defineExpose({
    acceptParams,
});
</script>
