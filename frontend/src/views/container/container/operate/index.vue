<template>
    <el-drawer
        v-model="drawerVisible"
        @close="handleClose"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader
                :header="title"
                :hideResource="dialogData.title === 'create'"
                :resource="dialogData.rowData?.name"
                :back="handleClose"
            />
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
                    <el-form-item class="mt-5" :label="$t('commons.table.name')" prop="name">
                        <el-input
                            :disabled="isFromApp(dialogData.rowData!)"
                            clearable
                            v-model.trim="dialogData.rowData!.name"
                        />
                        <div v-if="dialogData.title === 'edit' && isFromApp(dialogData.rowData!)">
                            <span class="input-help">
                                {{ $t('container.containerFromAppHelper1') }}
                                <el-button
                                    style="margin-left: -5px"
                                    size="small"
                                    text
                                    type="primary"
                                    @click="goRouter()"
                                >
                                    <el-icon><Position /></el-icon>
                                    {{ $t('firewall.quickJump') }}
                                </el-button>
                            </span>
                        </div>
                    </el-form-item>
                    <el-form-item :label="$t('container.image')" prop="image">
                        <el-checkbox v-model="dialogData.rowData!.imageInput" :label="$t('container.input')" />
                        <el-select
                            v-if="!dialogData.rowData!.imageInput"
                            filterable
                            v-model="dialogData.rowData!.image"
                        >
                            <el-option
                                v-for="(item, index) of images"
                                :key="index"
                                :value="item.option"
                                :label="item.option"
                            />
                        </el-select>
                        <el-input v-else v-model="dialogData.rowData!.image" />
                    </el-form-item>
                    <el-form-item prop="forcePull">
                        <el-checkbox v-model="dialogData.rowData!.forcePull">
                            {{ $t('container.forcePull') }}
                        </el-checkbox>
                        <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.port')">
                        <el-radio-group v-model="dialogData.rowData!.publishAllPorts" class="ml-4">
                            <el-radio :value="false">{{ $t('container.exposePort') }}</el-radio>
                            <el-radio :value="true">{{ $t('container.exposeAll') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item v-if="!dialogData.rowData!.publishAllPorts">
                        <el-card class="widthClass">
                            <el-table
                                v-if="dialogData.rowData!.exposedPorts.length !== 0"
                                :data="dialogData.rowData!.exposedPorts"
                            >
                                <el-table-column :label="$t('container.server')" min-width="150">
                                    <template #default="{ row }">
                                        <el-input :placeholder="$t('container.serverExample')" v-model="row.host" />
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('container.container')" min-width="80">
                                    <template #default="{ row }">
                                        <el-input
                                            :placeholder="$t('container.containerExample')"
                                            v-model="row.containerPort"
                                        />
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.protocol')" min-width="50">
                                    <template #default="{ row }">
                                        <el-select
                                            v-model="row.protocol"
                                            style="width: 100%"
                                            :placeholder="$t('container.serverExample')"
                                        >
                                            <el-option label="tcp" value="tcp" />
                                            <el-option label="udp" value="udp" />
                                        </el-select>
                                    </template>
                                </el-table-column>
                                <el-table-column min-width="35">
                                    <template #default="scope">
                                        <el-button link type="primary" @click="handlePortsDelete(scope.$index)">
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </template>
                                </el-table-column>
                            </el-table>

                            <el-button class="ml-3 mt-2" @click="handlePortsAdd()">
                                {{ $t('commons.button.add') }}
                            </el-button>
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

                    <el-form-item label="IPv4" prop="ipv4">
                        <el-input v-model="dialogData.rowData!.ipv4" :placeholder="$t('container.inputIpv4')" />
                    </el-form-item>
                    <el-form-item label="IPv6" prop="ipv6">
                        <el-input v-model="dialogData.rowData!.ipv6" :placeholder="$t('container.inputIpv6')" />
                    </el-form-item>

                    <el-form-item :label="$t('container.mount')">
                        <div v-for="(row, index) in dialogData.rowData!.volumes" :key="index" style="width: 100%">
                            <el-card class="mt-1">
                                <el-radio-group v-model="row.type">
                                    <el-radio-button value="volume">{{ $t('container.volumeOption') }}</el-radio-button>
                                    <el-radio-button value="bind">{{ $t('container.hostOption') }}</el-radio-button>
                                </el-radio-group>
                                <el-button
                                    class="float-right mt-3"
                                    link
                                    type="primary"
                                    @click="handleVolumesDelete(index)"
                                >
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                                <el-row class="mt-4" :gutter="5">
                                    <el-col :span="10">
                                        <el-form-item
                                            v-if="row.type === 'volume'"
                                            :label="$t('container.volumeOption')"
                                        >
                                            <el-select filterable v-model="row.sourceDir">
                                                <div v-for="(item, indexV) of volumes" :key="indexV">
                                                    <el-tooltip :hide-after="20" :content="item.option" placement="top">
                                                        <el-option
                                                            :value="item.option"
                                                            :label="item.option.substring(0, 30)"
                                                        />
                                                    </el-tooltip>
                                                </div>
                                            </el-select>
                                        </el-form-item>
                                        <el-form-item v-else :label="$t('container.hostOption')">
                                            <el-input v-model="row.sourceDir" />
                                        </el-form-item>
                                    </el-col>
                                    <el-col :span="5">
                                        <el-form-item :label="$t('container.mode')">
                                            <el-select class="widthClass" filterable v-model="row.mode">
                                                <el-option value="rw" :label="$t('container.modeRW')" />
                                                <el-option value="ro" :label="$t('container.modeR')" />
                                            </el-select>
                                        </el-form-item>
                                    </el-col>
                                    <el-col :span="9">
                                        <el-form-item :label="$t('container.containerDir')">
                                            <el-input v-model="row.containerDir" />
                                        </el-form-item>
                                    </el-col>
                                </el-row>
                            </el-card>
                        </div>
                        <el-button @click="handleVolumesAdd()">
                            {{ $t('commons.button.add') }}
                        </el-button>
                    </el-form-item>
                    <el-form-item label="Command" prop="cmdStr">
                        <el-input
                            type="textarea"
                            v-model="dialogData.rowData!.cmdStr"
                            :placeholder="$t('container.cmdHelper')"
                        />
                    </el-form-item>
                    <el-form-item label="Entrypoint" prop="entrypointStr">
                        <el-input
                            v-model="dialogData.rowData!.entrypointStr"
                            :placeholder="$t('container.entrypointHelper')"
                        />
                    </el-form-item>
                    <el-form-item prop="autoRemove">
                        <el-checkbox v-model="dialogData.rowData!.autoRemove">
                            {{ $t('container.autoRemove') }}
                        </el-checkbox>
                    </el-form-item>
                    <el-form-item>
                        <el-checkbox v-model="dialogData.rowData!.privileged">
                            {{ $t('container.privileged') }}
                        </el-checkbox>
                        <span class="input-help">{{ $t('container.privilegedHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('container.console')">
                        <el-checkbox v-model="dialogData.rowData!.tty">{{ $t('container.tty') }}</el-checkbox>
                        <el-checkbox v-model="dialogData.rowData!.openStdin">
                            {{ $t('container.openStdin') }}
                        </el-checkbox>
                    </el-form-item>
                    <el-form-item :label="$t('container.restartPolicy')" prop="restartPolicy">
                        <el-radio-group v-model="dialogData.rowData!.restartPolicy">
                            <el-radio value="no">{{ $t('container.no') }}</el-radio>
                            <el-radio value="always">{{ $t('container.always') }}</el-radio>
                            <el-radio value="on-failure">{{ $t('container.onFailure') }}</el-radio>
                            <el-radio value="unless-stopped">{{ $t('container.unlessStopped') }}</el-radio>
                        </el-radio-group>
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
                    <el-form-item
                        :label="$t('container.memoryLimit')"
                        prop="memory"
                        :rules="checkFloatNumberRange(0, Number(limits.memory))"
                    >
                        <el-input class="mini-form-item" v-model="dialogData.rowData!.memory">
                            <template #append><div style="width: 35px">MB</div></template>
                        </el-input>
                        <span class="input-help">{{ $t('container.limitHelper', [limits.memory]) }}MB</span>
                    </el-form-item>
                    <el-form-item :label="$t('container.tag')" prop="labelsStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :rows="3"
                            v-model="dialogData.rowData!.labelsStr"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('container.env')" prop="envStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :rows="3"
                            v-model="dialogData.rowData!.envStr"
                        />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
    <el-dialog v-model="dialogVisible" width="30%" :title="$t('commons.button.edit')">
        <div v-if="dialogData.title === 'edit' && isFromApp(dialogData.rowData!)" class="leading-6">
            <div>
                <span>{{ $t('container.updateHelper1') }}</span>
            </div>
            <br />
            <div>
                <span>{{ $t('container.updateHelper2') }}</span>
            </div>
            <div>
                <span>{{ $t('container.updateHelper3') }}</span>
            </div>
            <br />
        </div>
        <div>
            <span>{{ $t('container.updateHelper4') }}</span>
        </div>
        <template #footer>
            <el-button :disabled="loading" @click="dialogVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" type="primary" @click="submit()">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </el-dialog>
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
    searchContainer,
} from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { MsgError, MsgSuccess } from '@/utils/message';
import { checkIpV4V6, checkPort } from '@/utils/util';
import router from '@/routers';

const loading = ref(false);
interface DialogProps {
    title: string;
    rowData?: Container.ContainerHelper;
    getTableList?: () => Promise<any>;
}

const title = ref<string>('');
const drawerVisible = ref(false);
const dialogVisible = ref(false);

const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('container.' + dialogData.value.title);
    if (params.title === 'edit') {
        dialogData.value.rowData.memory = Number(dialogData.value.rowData.memory.toFixed(2));

        let itemCmd = '';
        dialogData.value.rowData.cmd = dialogData.value.rowData?.cmd || [];
        for (const item of dialogData.value.rowData.cmd) {
            if (item.indexOf(' ') !== -1) {
                itemCmd += `"${escapeQuotes(item)}" `;
            } else {
                itemCmd += item + ' ';
            }
        }
        dialogData.value.rowData.cmdStr = itemCmd.trimEnd();
        let itemEntrypoint = '';
        dialogData.value.rowData.entrypoint = dialogData.value.rowData?.entrypoint || [];
        for (const item of dialogData.value.rowData.entrypoint) {
            if (item.indexOf(' ') !== -1) {
                itemEntrypoint += `"${escapeQuotes(item)}" `;
            } else {
                itemEntrypoint += item + ' ';
            }
        }
        dialogData.value.rowData.entrypointStr = itemEntrypoint.trimEnd();

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
    drawerVisible.value = true;
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
    emit('search');
    drawerVisible.value = false;
    dialogVisible.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput, Rules.containerName],
    image: [Rules.imageName],
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

const goRouter = async () => {
    router.push({ name: 'AppInstalled' });
};

const handleVolumesAdd = () => {
    let item = {
        type: 'bind',
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
        if (dialogData.value.title === 'create') {
            submit();
        } else {
            dialogVisible.value = true;
        }
    });
};

const submit = async () => {
    dialogVisible.value = false;
    if (dialogData.value.rowData?.envStr) {
        dialogData.value.rowData.env = dialogData.value.rowData!.envStr.split('\n');
    }
    if (dialogData.value.rowData?.labelsStr) {
        dialogData.value.rowData!.labels = dialogData.value.rowData!.labelsStr.split('\n');
    }
    dialogData.value.rowData!.cmd = [];
    if (dialogData.value.rowData?.cmdStr) {
        let itemCmd = splitStringIgnoringQuotes(dialogData.value.rowData?.cmdStr);
        for (const item of itemCmd) {
            dialogData.value.rowData!.cmd.push(item.replace(/(?<!\\)"/g, '').replaceAll('\\"', '"'));
        }
    }
    dialogData.value.rowData!.entrypoint = [];
    if (dialogData.value.rowData?.entrypointStr) {
        let itemEntrypoint = splitStringIgnoringQuotes(dialogData.value.rowData?.entrypointStr);
        for (const item of itemEntrypoint) {
            dialogData.value.rowData!.entrypoint.push(item.replace(/(?<!\\)"/g, '').replaceAll('\\"', '"'));
        }
    }
    if (dialogData.value.rowData!.publishAllPorts) {
        dialogData.value.rowData!.exposedPorts = [];
    } else {
        if (!checkPortValid()) {
            return;
        }
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
                drawerVisible.value = false;
                dialogVisible.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
        return;
    }
    await updateContainer(dialogData.value.rowData!)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            drawerVisible.value = false;
            dialogVisible.value = false;
        })
        .catch(() => {
            updateContainerID();
            loading.value = false;
        });
};

const updateContainerID = async () => {
    let params = {
        page: 1,
        pageSize: 1,
        state: 'all',
        name: dialogData.value.rowData.name,
        filters: '',
        orderBy: 'created_at',
        order: 'null',
    };
    await searchContainer(params).then((res) => {
        if (res.data.items?.length === 1) {
            dialogData.value.rowData.containerID = res.data.items[0].containerID;
            return;
        }
    });
};

const checkPortValid = () => {
    if (dialogData.value.rowData!.exposedPorts.length === 0) {
        return true;
    }
    for (const port of dialogData.value.rowData!.exposedPorts) {
        if (port.host.indexOf(':') !== -1) {
            port.hostIP = port.host.substring(0, port.host.lastIndexOf(':'));
            if (checkIpV4V6(port.hostIP)) {
                MsgError(i18n.global.t('firewall.addressFormatError'));
                return false;
            }
            port.hostPort = port.host.substring(port.host.lastIndexOf(':') + 1);
        } else {
            port.hostIP = '';
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

const isFromApp = (rowData: Container.ContainerHelper) => {
    if (rowData && rowData.labels) {
        return rowData.labels.indexOf('createdBy=Apps') > -1;
    }
    return false;
};

const escapeQuotes = (input) => {
    return input.replace(/(?<!\\)"/g, '\\"');
};

const splitStringIgnoringQuotes = (input) => {
    input = input.replace(/\\"/g, '<quota>');
    const regex = /"([^"]*)"|(\S+)/g;
    const result = [];
    let match;

    while ((match = regex.exec(input)) !== null) {
        if (match[1]) {
            result.push(match[1].replaceAll('<quota>', '\\"'));
        } else if (match[2]) {
            result.push(match[2].replaceAll('<quota>', '\\"'));
        }
    }

    return result;
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
