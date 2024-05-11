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
                <DrawerHeader header="Hosts" :back="handleClose" />
            </template>

            <el-row type="flex" justify="center" v-loading="loading">
                <el-col :span="22">
                    <el-radio-group v-model="confShowType" @change="changeMode">
                        <el-radio-button value="base">{{ $t('database.baseConf') }}</el-radio-button>
                        <el-radio-button value="all">{{ $t('database.allConf') }}</el-radio-button>
                    </el-radio-group>
                    <div v-if="confShowType === 'base'">
                        <el-table :data="form.hosts">
                            <el-table-column label="IP" min-width="60">
                                <template #default="{ row }">
                                    <el-input placeholder="172.16.10.111" v-model="row.ip" />
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('toolbox.device.hosts')" min-width="150">
                                <template #default="{ row }">
                                    <el-input placeholder="test.hostname.com" v-model="row.host" />
                                </template>
                            </el-table-column>
                            <el-table-column min-width="30">
                                <template #default="scope">
                                    <el-button link type="primary" @click="handleHostsDelete(scope.$index)">
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </template>
                            </el-table-column>
                        </el-table>
                        <el-button class="ml-3 mt-2" @click="handleHostsAdd()">
                            {{ $t('commons.button.add') }}
                        </el-button>
                    </div>
                    <div v-else>
                        <codemirror
                            :autofocus="true"
                            placeholder="# The hosts configuration file does not exist or is empty (/etc/hosts)"
                            :indent-with-tab="true"
                            :tabSize="4"
                            style="margin-top: 10px; height: calc(100vh - 200px)"
                            :lineWrapping="true"
                            :matchBrackets="true"
                            theme="cobalt"
                            :styleActiveLine="true"
                            :extensions="extensions"
                            v-model="hostsConf"
                        />
                    </div>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { loadDeviceConf, updateDeviceByConf, updateDeviceHost } from '@/api/modules/toolbox';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { Toolbox } from '@/api/interface/toolbox';

const emit = defineEmits<{ (e: 'search'): void }>();

const extensions = [javascript(), oneDark];
const confShowType = ref('base');
const hostsConf = ref();

const form = reactive({
    hosts: [],
});

interface DialogProps {
    hosts: Array<Toolbox.HostHelper>;
}

const drawerVisible = ref();
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    confShowType.value = 'base';
    form.hosts = params.hosts;
    drawerVisible.value = true;
};

const loadHostsConf = async () => {
    const res = await loadDeviceConf('Hosts');
    hostsConf.value = res.data || '';
};

const changeMode = async () => {
    if (confShowType.value === 'all') {
        loadHostsConf();
    }
};

const handleHostsAdd = () => {
    let item = {
        ip: '',
        host: '',
    };
    form.hosts.push(item);
};
const handleHostsDelete = (index: number) => {
    form.hosts.splice(index, 1);
};

const onSave = async () => {
    loading.value = true;
    if (confShowType.value === 'base') {
        await updateDeviceHost(form.hosts)
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
    for (const item of form.hosts) {
        if (item.ip === '' || item.host === '') {
            MsgError(i18n.global.t('toolbox.device.hostHelper'));
            return;
        }
    }
    await updateDeviceByConf('Hosts', hostsConf.value)
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
