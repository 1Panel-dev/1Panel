<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader header="DNS" :back="handleClose" />
            </template>

            <el-row type="flex" justify="center" v-loading="loading">
                <el-col :span="22">
                    <el-alert
                        :title="$t('toolbox.device.dnsAlert')"
                        class="common-prompt"
                        :closable="false"
                        type="warning"
                    />
                    <el-radio-group v-model="confShowType" @change="changeMode">
                        <el-radio-button value="form">{{ $t('database.baseConf') }}</el-radio-button>
                        <el-radio-button value="all">{{ $t('database.allConf') }}</el-radio-button>
                    </el-radio-group>
                    <el-form
                        class="mt-4"
                        v-if="confShowType === 'form'"
                        ref="formRef"
                        label-position="top"
                        @submit.prevent
                        :model="form"
                    >
                        <el-form-item label="DNS" prop="dns">
                            <el-input
                                type="textarea"
                                :placeholder="$t('toolbox.device.dnsHelper1')"
                                :rows="8"
                                v-model="form.dns"
                            />
                        </el-form-item>
                    </el-form>
                    <div v-else>
                        <codemirror
                            :autofocus="true"
                            placeholder="# The DNS configuration file does not exist or is empty (/etc/resolv.conf)"
                            :indent-with-tab="true"
                            :tabSize="4"
                            style="margin-top: 10px; height: calc(100vh - 260px)"
                            :lineWrapping="true"
                            :matchBrackets="true"
                            theme="cobalt"
                            :styleActiveLine="true"
                            :extensions="extensions"
                            v-model="dnsConf"
                        />
                    </div>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" @click="onTest()">
                        {{ $t('toolbox.device.dnsCheck') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave()">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { loadDeviceConf, checkDNS, updateDevice, updateDeviceByConf } from '@/api/modules/toolbox';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';

const emit = defineEmits<{ (e: 'search'): void }>();

const extensions = [javascript(), oneDark];
const confShowType = ref('form');
const dnsConf = ref();

const form = reactive({
    dns: '',
});

interface DialogProps {
    dns: Array<string>;
}

const drawerVisible = ref();
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    form.dns = params.dns ? params.dns.join('\n') : '';
    drawerVisible.value = true;
};

const loadDNSConf = async () => {
    const res = await loadDeviceConf('DNS');
    dnsConf.value = res.data || '';
};

const changeMode = async () => {
    if (confShowType.value === 'all') {
        loadDNSConf();
    }
};

const onTest = async () => {
    loading.value = true;
    let value = '';
    if (confShowType.value === 'form') {
        value = form.dns.replaceAll('\n', ',');
    } else {
        value = dnsConf.value;
    }
    await checkDNS(confShowType.value, value)
        .then((res) => {
            loading.value = false;
            if (res.data) {
                MsgSuccess(i18n.global.t('toolbox.device.dnsOK'));
            } else {
                MsgError(i18n.global.t('toolbox.device.dnsTestFailed'));
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSave = async () => {
    loading.value = true;
    if (confShowType.value == 'form') {
        await updateDevice('DNS', form.dns.replaceAll('\n', ','))
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                handleClose();
            })
            .catch(() => {
                loading.value = false;
            });
        return;
    }
    await updateDeviceByConf('DNS', dnsConf.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            handleClose();
        })
        .catch(() => {
            loading.value = false;
        });
};
const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
