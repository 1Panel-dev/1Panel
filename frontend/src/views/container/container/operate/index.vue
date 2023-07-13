<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="title" :resource="dialogData.rowData?.name" :back="handleClose" />
        </template>
        <el-form
            ref="formRef"
            label-position="top"
            v-loading="loading"
            :model="dialogData.rowData!"
            :rules="rules"
            label-width="80px"
        >
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input clearable v-model.trim="dialogData.rowData!.name" />
                    </el-form-item>
                    <el-form-item :label="$t('container.image')" prop="image">
                        <el-select class="widthClass" allow-create filterable v-model="dialogData.rowData!.image">
                            <el-option
                                v-for="(item, index) of images"
                                :key="index"
                                :value="item.option"
                                :label="item.option"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item prop="forcePull">
                        <el-checkbox v-model="dialogData.rowData!.forcePull">
                            {{ $t('container.forcePull') }}
                        </el-checkbox>
                        <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.port')">
                        <el-radio-group v-model="dialogData.rowData!.publishAllPorts" class="ml-4">
                            <el-radio :label="false">{{ $t('container.exposePort') }}</el-radio>
                            <el-radio :label="true">{{ $t('container.exposeAll') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item v-if="!dialogData.rowData!.publishAllPorts">
                        <el-card class="widthClass">
                            <table style="width: 100%" class="tab-table">
                                <tr v-if="dialogData.rowData!.exposedPorts.length !== 0">
                                    <th scope="col" width="45%" align="left">
                                        <label>{{ $t('container.server') }}</label>
                                    </th>
                                    <th scope="col" width="35%" align="left">
                                        <label>{{ $t('container.container') }}</label>
                                    </th>
                                    <th scope="col" width="20%" align="left">
                                        <label>{{ $t('commons.table.protocol') }}</label>
                                    </th>
                                    <th align="left"></th>
                                </tr>
                                <tr v-for="(row, index) in dialogData.rowData!.exposedPorts" :key="index">
                                    <td width="45%">
                                        <el-input
                                            :placeholder="$t('container.serverExample')"
                                            style="width: 100%"
                                            v-model="row.host"
                                        />
                                    </td>
                                    <td width="35%">
                                        <el-input
                                            :placeholder="$t('container.containerExample')"
                                            style="width: 100%"
                                            v-model="row.containerPort"
                                        />
                                    </td>
                                    <td width="20%">
                                        <el-select v-model="row.protocol" style="width: 100%">
                                            <el-option label="tcp" value="tcp" />
                                            <el-option label="udp" value="udp" />
                                        </el-select>
                                    </td>
                                    <td>
                                        <el-button link style="font-size: 10px" @click="handlePortsDelete(index)">
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
                        </el-card>
                    </el-form-item>
                    <el-form-item :label="$t('container.network')" prop="network">
                        <el-select v-model="dialogData.rowData!.network">
                            <el-option
                                v-for="(item, indexV) of networks"
                                :key="indexV"
                                :value="item.option"
                                :label="item.option"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.cmd')" prop="cmdStr">
                        <el-input :placeholder="$t('container.cmdHelper')" v-model="dialogData.rowData!.cmdStr" />
                    </el-form-item>
                    <el-form-item prop="autoRemove">
                        <el-checkbox v-model="dialogData.rowData!.autoRemove">
                            {{ $t('container.autoRemove') }}
                        </el-checkbox>
                    </el-form-item>
                    <el-form-item :label="$t('container.cpuShare')" prop="cpuShares">
                        <el-input class="mini-form-item" v-model.number="dialogData.rowData!.cpuShares" />
                        <span class="input-help">{{ $t('container.cpuShareHelper') }}</span>
                    </el-form-item>
                    <el-form-item
                        :label="$t('container.cpuQuota')"
                        prop="nanoCPUs"
                        :rules="checkFloatNumberRange(0, Number(limits.cpu))"
                    >
                        <el-input class="mini-form-item" v-model="dialogData.rowData!.nanoCPUs">
                            <template #append>
                                <div style="width: 35px">{{ $t('commons.units.core') }}</div>
                            </template>
                        </el-input>
                        <span class="input-help">
                            {{ $t('container.limitHelper', [limits.cpu]) }}{{ $t('commons.units.core') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('container.memoryLimit')" prop="memory">
                        <el-input class="mini-form-item" v-model="dialogData.rowData!.memory">
                            <template #append><div style="width: 35px">MB</div></template>
                        </el-input>
                        <span class="input-help">{{ $t('container.limitHelper', [limits.memory]) }}MB</span>
                    </el-form-item>
                    <el-form-item :label="$t('container.mount')">
                        <el-card style="width: 100%">
                            <table style="width: 100%" class="tab-table">
                                <tr v-if="dialogData.rowData!.volumes!.length !== 0">
                                    <th scope="col" width="39%" align="left">
                                        <label>{{ $t('container.serverPath') }}</label>
                                    </th>
                                    <th scope="col" width="18%" align="left">
                                        <label>{{ $t('container.mode') }}</label>
                                    </th>
                                    <th scope="col" width="39%" align="left">
                                        <label>{{ $t('container.containerDir') }}</label>
                                    </th>
                                    <th align="left"></th>
                                </tr>
                                <tr v-for="(row, index) in dialogData.rowData!.volumes" :key="index">
                                    <td width="39%">
                                        <el-select
                                            class="widthClass"
                                            allow-create
                                            clearable
                                            :placeholder="$t('commons.msg.inputOrSelect')"
                                            filterable
                                            v-model="row.sourceDir"
                                        >
                                            <el-option
                                                v-for="(item, indexV) of volumes"
                                                :key="indexV"
                                                :value="item.option"
                                                :label="item.option"
                                            />
                                        </el-select>
                                    </td>
                                    <td width="18%">
                                        <el-select class="widthClass" filterable v-model="row.mode">
                                            <el-option value="rw" :label="$t('container.modeRW')" />
                                            <el-option value="ro" :label="$t('container.modeR')" />
                                        </el-select>
                                    </td>
                                    <td width="39%">
                                        <el-input v-model="row.containerDir" />
                                    </td>
                                    <td>
                                        <el-button link style="font-size: 10px" @click="handleVolumesDelete(index)">
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </td>
                                </tr>
                                <tr>
                                    <td align="left">
                                        <el-button @click="handleVolumesAdd()">
                                            {{ $t('commons.button.add') }}
                                        </el-button>
                                    </td>
                                </tr>
                            </table>
                        </el-card>
                    </el-form-item>
                    <el-form-item :label="$t('container.tag')" prop="labelsStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :autosize="{ minRows: 2, maxRows: 10 }"
                            v-model="dialogData.rowData!.labelsStr"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('container.env')" prop="envStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :autosize="{ minRows: 2, maxRows: 10 }"
                            v-model="dialogData.rowData!.envStr"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('container.restartPolicy')" prop="restartPolicy">
                        <el-radio-group v-model="dialogData.rowData!.restartPolicy">
                            <el-radio label="no">{{ $t('container.no') }}</el-radio>
                            <el-radio label="always">{{ $t('container.always') }}</el-radio>
                            <el-radio label="on-failure">{{ $t('container.onFailure') }}</el-radio>
                            <el-radio label="unless-stopped">{{ $t('container.unlessStopped') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisiable = false">
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
import { Rules, checkFloatNumberRange, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import {
    listImage,
    listVolume,
    createContainer,
    updateContainer,
    loadResourceLimit,
    listNetwork,
} from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { MsgError, MsgSuccess } from '@/utils/message';
import { checkIpV4V6, checkPort } from '@/utils/util';

const loading = ref(false);
interface DialogProps {
    title: string;
    rowData?: Container.ContainerHelper;
    getTableList?: () => Promise<any>;
}

const title = ref<string>('');
const drawerVisiable = ref(false);

const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('container.' + dialogData.value.title);
    if (params.title === 'edit') {
        dialogData.value.rowData.memory = Number(dialogData.value.rowData.memory.toFixed(2));
        dialogData.value.rowData.cmd = dialogData.value.rowData.cmd || [];
        let itemCmd = '';
        for (const item of dialogData.value.rowData.cmd) {
            itemCmd += `'${item}' `;
        }
        dialogData.value.rowData.cmdStr = itemCmd ? itemCmd.substring(0, itemCmd.length - 1) : '';
        dialogData.value.rowData.labels = dialogData.value.rowData.labels || [];
        dialogData.value.rowData.env = dialogData.value.rowData.env || [];
        dialogData.value.rowData.labelsStr = dialogData.value.rowData.labels.join('\n');
        dialogData.value.rowData.envStr = dialogData.value.rowData.env.join('\n');
        dialogData.value.rowData.exposedPorts = dialogData.value.rowData.exposedPorts || [];
        for (const item of dialogData.value.rowData.exposedPorts) {
            if (item.hostIP) {
                item.host = item.hostIP + ':' + item.hostPort;
            } else {
                item.host = item.hostPort;
            }
        }
        dialogData.value.rowData.volumes = dialogData.value.rowData.volumes || [];
    }
    loadLimit();
    loadImageOptions();
    loadVolumeOptions();
    loadNetworkOptions();
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const images = ref();
const volumes = ref();
const networks = ref();
const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});

const handleClose = () => {
    drawerVisiable.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput, Rules.volumeName],
    image: [Rules.requiredSelect],
    network: [Rules.requiredSelect],
    cpuShares: [Rules.integerNumberWith0, checkNumberRange(0, 262144)],
    nanoCPUs: [Rules.floatNumber],
    memory: [Rules.floatNumber],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const handlePortsAdd = () => {
    let item = {
        host: '',
        hostIP: '',
        containerPort: '',
        hostPort: '',
        protocol: 'tcp',
    };
    dialogData.value.rowData!.exposedPorts.push(item);
};
const handlePortsDelete = (index: number) => {
    dialogData.value.rowData!.exposedPorts.splice(index, 1);
};

const handleVolumesAdd = () => {
    let item = {
        sourceDir: '',
        containerDir: '',
        mode: 'rw',
    };
    dialogData.value.rowData!.volumes.push(item);
};
const handleVolumesDelete = (index: number) => {
    dialogData.value.rowData!.volumes.splice(index, 1);
};

const loadLimit = async () => {
    const res = await loadResourceLimit();
    limits.value = res.data;
    limits.value.memory = Number((limits.value.memory / 1024 / 1024).toFixed(2));
};

const loadImageOptions = async () => {
    const res = await listImage();
    images.value = res.data;
};
const loadVolumeOptions = async () => {
    const res = await listVolume();
    volumes.value = res.data;
};
const loadNetworkOptions = async () => {
    const res = await listNetwork();
    networks.value = res.data;
};
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (dialogData.value.rowData!.volumes.length !== 0) {
        for (const item of dialogData.value.rowData!.volumes) {
            if (!item.containerDir || !item.sourceDir) {
                MsgError(i18n.global.t('container.volumeHelper'));
                return;
            }
        }
    }
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (dialogData.value.rowData?.envStr) {
            dialogData.value.rowData.env = dialogData.value.rowData!.envStr.split('\n');
        }
        if (dialogData.value.rowData?.labelsStr) {
            dialogData.value.rowData!.labels = dialogData.value.rowData!.labelsStr.split('\n');
        }
        dialogData.value.rowData!.cmd = [];
        if (dialogData.value.rowData?.cmdStr) {
            let itemCmd = dialogData.value.rowData!.cmdStr.split(`'`);
            for (const cmd of itemCmd) {
                if (cmd && cmd !== ' ') {
                    dialogData.value.rowData!.cmd.push(cmd);
                }
            }
        }
        if (!checkPortValid()) {
            return;
        }
        dialogData.value.rowData!.memory = Number(dialogData.value.rowData!.memory);
        dialogData.value.rowData!.nanoCPUs = Number(dialogData.value.rowData!.nanoCPUs);

        loading.value = true;
        if (dialogData.value.title === 'create') {
            await createContainer(dialogData.value.rowData!)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisiable.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
        } else {
            ElMessageBox.confirm(
                i18n.global.t('container.updateContaienrHelper'),
                i18n.global.t('commons.button.edit'),
                {
                    confirmButtonText: i18n.global.t('commons.button.confirm'),
                    cancelButtonText: i18n.global.t('commons.button.cancel'),
                },
            )
                .then(async () => {
                    await updateContainer(dialogData.value.rowData!)
                        .then(() => {
                            loading.value = false;
                            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                            emit('search');
                            drawerVisiable.value = false;
                        })
                        .catch(() => {
                            loading.value = false;
                        });
                })
                .catch(() => {
                    loading.value = false;
                });
        }
    });
};

const checkPortValid = () => {
    if (dialogData.value.rowData!.exposedPorts.length === 0) {
        return true;
    }
    for (const port of dialogData.value.rowData!.exposedPorts) {
        if (port.host.indexOf(':') !== -1) {
            port.hostIP = port.host.split(':')[0];
            if (checkIpV4V6(port.hostIP)) {
                MsgError(i18n.global.t('firewall.addressFormatError'));
                return false;
            }
            port.hostPort = port.host.split(':')[1];
        } else {
            port.hostPort = port.host;
        }
        if (port.hostPort.indexOf('-') !== -1) {
            if (checkPort(port.hostPort.split('-')[0])) {
                MsgError(i18n.global.t('firewall.portFormatError'));
                return false;
            }
            if (checkPort(port.hostPort.split('-')[1])) {
                MsgError(i18n.global.t('firewall.portFormatError'));
                return false;
            }
        } else {
            if (checkPort(port.hostPort)) {
                MsgError(i18n.global.t('firewall.portFormatError'));
                return false;
            }
        }
        if (port.containerPort.indexOf('-') !== -1) {
            if (checkPort(port.containerPort.split('-')[0])) {
                MsgError(i18n.global.t('firewall.portFormatError'));
                return false;
            }
            if (checkPort(port.containerPort.split('-')[1])) {
                MsgError(i18n.global.t('firewall.portFormatError'));
                return false;
            }
        } else {
            if (checkPort(port.containerPort)) {
                MsgError(i18n.global.t('firewall.portFormatError'));
                return false;
            }
        }
    }
    return true;
};
defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.widthClass {
    width: 100%;
}
</style>
