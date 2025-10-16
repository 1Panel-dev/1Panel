<template>
    <div>
        <LayoutContent
            back-name="ContainerItem"
            :title="isCreate ? $t('container.create') : $t('commons.button.edit') + ' - ' + form.name"
        >
            <template #main>
                <el-form
                    ref="formRef"
                    label-position="top"
                    v-loading="loading"
                    :model="form"
                    :rules="rules"
                    label-width="80px"
                >
                    <el-row type="flex" justify="center" :gutter="20">
                        <el-col :span="20">
                            <el-card>
                                <el-alert
                                    v-if="!isCreate && isFromApp(form)"
                                    :title="$t('container.containerFromAppHelper')"
                                    :closable="false"
                                    type="error"
                                />

                                <el-button v-if="isCreate" type="primary" icon="EditPen" plain @click="toTerminal()">
                                    {{ $t('container.commandInput') }}
                                </el-button>
                                <el-form-item class="mt-5" :label="$t('commons.table.name')" prop="name">
                                    <el-input
                                        :disabled="!isCreate"
                                        class="mini-form-item"
                                        clearable
                                        v-model.trim="form.name"
                                    />
                                    <span class="input-help" v-if="!isCreate && isFromApp(form)">
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
                                </el-form-item>
                                <el-form-item :label="$t('container.image')" prop="image">
                                    <el-checkbox v-model="form.imageInput" :label="$t('container.input')" />
                                </el-form-item>
                                <el-form-item>
                                    <el-select
                                        class="mini-form-item"
                                        v-if="!form.imageInput"
                                        filterable
                                        v-model="form.image"
                                    >
                                        <el-option
                                            v-for="(item, index) of images"
                                            :key="index"
                                            :value="item.option"
                                            :label="item.option"
                                        />
                                    </el-select>
                                    <el-input class="mini-form-item" v-else v-model="form.image" />
                                </el-form-item>
                                <el-form-item prop="forcePull">
                                    <el-checkbox v-model="form.forcePull">
                                        {{ $t('container.forcePull') }}
                                    </el-checkbox>
                                    <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
                                </el-form-item>

                                <el-form-item prop="autoRemove">
                                    <el-checkbox v-model="form.autoRemove">
                                        {{ $t('container.autoRemove') }}
                                    </el-checkbox>
                                </el-form-item>
                                <el-form-item :label="$t('commons.table.port')">
                                    <el-radio-group v-model="form.publishAllPorts" class="ml-4">
                                        <el-radio :value="false">{{ $t('container.exposePort') }}</el-radio>
                                        <el-radio :value="true">{{ $t('container.exposeAll') }}</el-radio>
                                    </el-radio-group>
                                </el-form-item>
                                <el-form-item v-if="!form.publishAllPorts">
                                    <el-table v-if="form.exposedPorts.length !== 0" :data="form.exposedPorts">
                                        <el-table-column :label="$t('container.server')" min-width="200">
                                            <template #default="{ row }">
                                                <el-input
                                                    :placeholder="$t('container.serverExample')"
                                                    v-model="row.host"
                                                />
                                            </template>
                                        </el-table-column>
                                        <el-table-column :label="$t('menu.container')" min-width="120">
                                            <template #default="{ row }">
                                                <el-input
                                                    :placeholder="$t('container.containerExample')"
                                                    v-model="row.containerPort"
                                                />
                                            </template>
                                        </el-table-column>
                                        <el-table-column :label="$t('commons.table.protocol')" min-width="100">
                                            <template #default="{ row }">
                                                <el-radio-group v-model="row.protocol">
                                                    <el-radio value="tcp">tcp</el-radio>
                                                    <el-radio value="udp">udp</el-radio>
                                                </el-radio-group>
                                            </template>
                                        </el-table-column>
                                        <el-table-column min-width="80">
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
                                </el-form-item>
                            </el-card>

                            <el-tabs type="border-card" class="mt-5">
                                <el-tab-pane :label="$t('container.network')">
                                    <el-row :gutter="20">
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item :label="$t('container.network')" prop="network">
                                                <el-select v-model="form.network">
                                                    <el-option
                                                        v-for="(item, indexV) of networks"
                                                        :key="indexV"
                                                        :value="item.option"
                                                        :label="item.option"
                                                    />
                                                </el-select>
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item :label="$t('toolbox.device.hostname')" prop="hostname">
                                                <el-input v-model="form.hostname" />
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item label="Domain" prop="domainName">
                                                <el-input v-model="form.domainName" />
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item :label="$t('container.macAddr')" prop="macAddr">
                                                <el-input v-model="form.macAddr" />
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item label="IPv4" prop="ipv4">
                                                <el-input
                                                    v-model="form.ipv4"
                                                    :placeholder="$t('container.inputIpv4')"
                                                />
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item label="IPv6" prop="ipv6">
                                                <el-input
                                                    v-model="form.ipv6"
                                                    :placeholder="$t('container.inputIpv6')"
                                                />
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item label="DNS" prop="dns">
                                                <div v-for="(_, index) of form.dns" :key="index" class="w-full">
                                                    <el-input class="mt-2" v-model="form.dns[index]">
                                                        <template #append>
                                                            <el-button
                                                                link
                                                                icon="Delete"
                                                                @click="form.dns.splice(index, 1)"
                                                            />
                                                        </template>
                                                    </el-input>
                                                </div>
                                                <el-button class="mt-2" @click="form.dns.push('')">
                                                    {{ $t('commons.button.add') }}
                                                </el-button>
                                            </el-form-item>
                                        </el-col>
                                    </el-row>
                                </el-tab-pane>

                                <el-tab-pane :label="$t('container.mount')">
                                    <Volume ref="volumeRef" :volumes="form.volumes"></Volume>
                                </el-tab-pane>

                                <el-tab-pane :label="$t('terminal.command')">
                                    <el-row :gutter="20">
                                        <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                                            <el-form-item label="Command" prop="cmdStr">
                                                <el-input
                                                    v-model="form.cmdStr"
                                                    :placeholder="$t('container.cmdHelper')"
                                                />
                                            </el-form-item>
                                        </el-col>
                                    </el-row>
                                    <el-row :gutter="20">
                                        <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                                            <el-form-item label="Entrypoint" prop="entrypointStr">
                                                <el-input
                                                    v-model="form.entrypointStr"
                                                    :placeholder="$t('container.entrypointHelper')"
                                                />
                                            </el-form-item>
                                        </el-col>
                                    </el-row>

                                    <el-row :gutter="20">
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item :label="$t('container.workingDir')" prop="workingDir">
                                                <el-input v-model="form.workingDir" />
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="10" :md="10" :lg="10" :xl="10">
                                            <el-form-item :label="$t('commons.table.user')" prop="user">
                                                <el-input v-model="form.user" />
                                            </el-form-item>
                                        </el-col>
                                    </el-row>
                                    <el-form-item :label="$t('container.console')">
                                        <el-checkbox v-model="form.tty">{{ $t('container.tty') }}</el-checkbox>
                                        <el-checkbox v-model="form.openStdin">
                                            {{ $t('container.openStdin') }}
                                        </el-checkbox>
                                    </el-form-item>
                                </el-tab-pane>

                                <el-tab-pane :label="$t('container.resource')">
                                    <el-form-item :label="$t('container.cpuShare')" prop="cpuShares">
                                        <el-input class="mini-form-item" v-model.number="form.cpuShares" />
                                        <span class="input-help">{{ $t('container.cpuShareHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item
                                        :label="$t('container.cpuQuota')"
                                        prop="nanoCPUs"
                                        :rules="checkFloatNumberRange(0, Number(limits.cpu))"
                                    >
                                        <el-input class="mini-form-item" v-model="form.nanoCPUs">
                                            <template #append>
                                                <div style="width: 35px">{{ $t('commons.units.core') }}</div>
                                            </template>
                                        </el-input>
                                        <span class="input-help">
                                            {{ $t('container.limitHelper', [limits.cpu])
                                            }}{{ $t('commons.units.core') }}
                                        </span>
                                    </el-form-item>
                                    <el-form-item
                                        :label="$t('container.memoryLimit')"
                                        prop="memory"
                                        :rules="checkFloatNumberRange(0, Number(limits.memory))"
                                    >
                                        <el-input class="mini-form-item" v-model="form.memory">
                                            <template #append><div style="width: 35px">MB</div></template>
                                        </el-input>
                                        <span class="input-help">
                                            {{ $t('container.limitHelper', [limits.memory]) }}MB
                                        </span>
                                    </el-form-item>
                                    <el-form-item>
                                        <el-checkbox v-model="form.privileged">
                                            {{ $t('container.privileged') }}
                                        </el-checkbox>
                                        <span class="input-help">{{ $t('container.privilegedHelper') }}</span>
                                    </el-form-item>
                                </el-tab-pane>

                                <el-tab-pane :label="$t('container.tag') + ' & ' + $t('container.env')">
                                    <el-row :gutter="20">
                                        <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                                            <el-form-item :label="$t('container.tag')" prop="labels">
                                                <div v-for="(_, index) of form.labels" :key="index" class="w-full">
                                                    <el-input
                                                        class="mt-2"
                                                        placeholder="e.g. key=val"
                                                        v-model="form.labels[index]"
                                                    >
                                                        <template #append>
                                                            <el-button
                                                                link
                                                                icon="Delete"
                                                                @click="form.labels.splice(index, 1)"
                                                            />
                                                        </template>
                                                    </el-input>
                                                </div>
                                                <el-button class="mt-2" @click="form.labels.push('')">
                                                    {{ $t('commons.button.add') }}
                                                </el-button>
                                            </el-form-item>
                                        </el-col>
                                        <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                                            <el-form-item :label="$t('container.env')" prop="envStr">
                                                <div v-for="(_, index) of form.env" :key="index" class="w-full">
                                                    <el-input
                                                        class="mt-2"
                                                        placeholder="e.g. key=val"
                                                        v-model="form.env[index]"
                                                    >
                                                        <template #append>
                                                            <el-button
                                                                link
                                                                icon="Delete"
                                                                @click="form.env.splice(index, 1)"
                                                            />
                                                        </template>
                                                    </el-input>
                                                </div>
                                                <el-button class="mt-2" @click="form.env.push('')">
                                                    {{ $t('commons.button.add') }}
                                                </el-button>
                                            </el-form-item>
                                        </el-col>
                                    </el-row>
                                </el-tab-pane>

                                <el-tab-pane :label="$t('container.restartPolicy')">
                                    <el-form-item prop="restartPolicy">
                                        <el-radio-group v-model="form.restartPolicy">
                                            <el-radio value="no">{{ $t('container.no') }}</el-radio>
                                            <el-radio value="always">{{ $t('container.always') }}</el-radio>
                                            <el-radio value="on-failure">{{ $t('container.onFailure') }}</el-radio>
                                            <el-radio value="unless-stopped">
                                                {{ $t('container.unlessStopped') }}
                                            </el-radio>
                                        </el-radio-group>
                                    </el-form-item>
                                </el-tab-pane>
                            </el-tabs>

                            <el-form-item class="mt-5">
                                <el-button :disabled="loading" @click="goBack">
                                    {{ $t('commons.button.back') }}
                                </el-button>
                                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                                    {{ $t('commons.button.confirm') }}
                                </el-button>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
        <Confirm ref="confirmRef" @submit="submit" />
        <TaskLog ref="taskLogRef" width="70%" :show-close="false">
            <template #task-footer>
                <el-button @click="goBack">{{ $t('commons.table.backToList') }}</el-button>
                <el-button type="primary" @click="closeTask">{{ $t('commons.table.keepEdit') }}</el-button>
            </template>
        </TaskLog>
        <TerminalDialog ref="dialogTerminalRef" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules, checkFloatNumberRange, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import Confirm from '@/views/container/container/operate/confirm.vue';
import Volume from '@/views/container/container/operate/volume.vue';
import {
    listImage,
    createContainer,
    updateContainer,
    loadResourceLimit,
    listNetwork,
    loadContainerInfo,
    searchContainer,
} from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { MsgError } from '@/utils/message';
import TaskLog from '@/components/log/task/index.vue';
import { checkIpV4V6, checkPort, newUUID } from '@/utils/util';
import router from '@/routers';
import TerminalDialog from '@/views/host/file-management/terminal/index.vue';
import { routerToName, routerToNameWithQuery } from '@/utils/router';

const loading = ref(false);
const isCreate = ref();
const confirmRef = ref();
const volumeRef = ref();
const form = reactive<Container.ContainerHelper>({
    taskID: '',
    name: '',
    image: '',
    imageInput: false,
    forcePull: false,
    network: 'bridge',
    hostname: '',
    domainName: '',
    macAddr: '',
    ipv4: '',
    ipv6: '',
    dns: [],
    cmdStr: '',
    entrypointStr: '',
    memoryItem: 0,
    cmd: [],
    workingDir: '',
    user: '',
    openStdin: false,
    tty: false,
    entrypoint: [],
    publishAllPorts: false,
    exposedPorts: [],
    nanoCPUs: 0,
    cpuShares: 1024,
    memory: 0,
    volumes: [],
    privileged: false,
    autoRemove: false,
    labels: [],
    env: [],
    restartPolicy: 'no',
});
const search = async () => {
    if (!isCreate.value) {
        loading.value = true;
        await loadContainerInfo(form.name)
            .then((res) => {
                loading.value = false;
                form.name = res.data.name;
                form.image = res.data.image;
                form.network = res.data.network;
                form.hostname = res.data.hostname;
                form.domainName = res.data.domainName;
                form.dns = res.data.dns || [];
                form.ipv4 = res.data.ipv4;
                form.ipv6 = res.data.ipv6;
                form.openStdin = res.data.openStdin;
                form.tty = res.data.tty;
                form.publishAllPorts = res.data.publishAllPorts;
                form.nanoCPUs = res.data.nanoCPUs;
                form.cpuShares = res.data.cpuShares;
                form.privileged = res.data.privileged;
                form.autoRemove = res.data.autoRemove;
                form.restartPolicy = res.data.restartPolicy;
                form.memory = Number(res.data.memory.toFixed(2));
                form.user = res.data.user;
                form.workingDir = res.data.workingDir;

                let itemCmd = '';
                form.cmd = res.data.cmd || [];
                for (const item of form.cmd) {
                    if (item.indexOf(' ') !== -1) {
                        itemCmd += `"${escapeQuotes(item)}" `;
                    } else {
                        itemCmd += item + ' ';
                    }
                }
                form.cmdStr = itemCmd.trimEnd();
                let itemEntrypoint = '';
                form.entrypoint = res.data.entrypoint || [];
                for (const item of form.entrypoint) {
                    if (item.indexOf(' ') !== -1) {
                        itemEntrypoint += `"${escapeQuotes(item)}" `;
                    } else {
                        itemEntrypoint += item + ' ';
                    }
                }
                form.entrypointStr = itemEntrypoint.trimEnd();

                form.labels = res.data.labels || [];
                form.env = res.data.env || [];
                form.exposedPorts = res.data.exposedPorts || [];
                for (const item of form.exposedPorts) {
                    if (item.hostIP) {
                        item.host = item.hostIP + ':' + item.hostPort;
                    } else {
                        item.host = item.hostPort;
                    }
                }
                form.volumes = res.data.volumes || [];
            })
            .catch(() => {
                loading.value = false;
            });
    }
    loadLimit();
    loadImageOptions();
    loadNetworkOptions();
};

const taskLogRef = ref();
const images = ref();
const networks = ref();
const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});

const rules = reactive({
    name: [Rules.requiredInput, Rules.containerName],
    image: [Rules.imageName],
    cpuShares: [Rules.integerNumberWith0, checkNumberRange(0, 262144)],
    nanoCPUs: [Rules.floatNumber],
    memory: [Rules.floatNumber],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const goBack = () => {
    routerToName('ContainerItem');
};
const closeTask = () => {
    taskLogRef.value.handleClose();
    checkExist();
};
const dialogTerminalRef = ref();
const toTerminal = () => {
    dialogTerminalRef.value!.acceptParams({ cwd: '$HOME', command: '/bin/sh' });
};

const handlePortsAdd = () => {
    let item = {
        host: '',
        hostIP: '',
        containerPort: '',
        hostPort: '',
        protocol: 'tcp',
    };
    form.exposedPorts.push(item);
};
const handlePortsDelete = (index: number) => {
    form.exposedPorts.splice(index, 1);
};

const goRouter = async () => {
    routerToName('AppInstalled');
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
const loadNetworkOptions = async () => {
    const res = await listNetwork();
    networks.value = res.data;
};
const onSubmit = async (formEl: FormInstance | undefined) => {
    form.volumes = volumeRef.value.loadVolumes();
    if (form.volumes.length !== 0) {
        for (const item of form.volumes) {
            if (!item.containerDir || !item.sourceDir) {
                MsgError(i18n.global.t('container.volumeHelper'));
                return;
            }
        }
    }
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (isCreate.value) {
            submit();
        } else {
            confirmRef.value.acceptParams({ isFromApp: isFromApp(form) });
        }
    });
};
const submit = async () => {
    form.cmd = [];
    form.taskID = newUUID();
    if (form.cmdStr) {
        let itemCmd = splitStringIgnoringQuotes(form.cmdStr);
        for (const item of itemCmd) {
            form.cmd.push(item.replace(/(?<!\\)"/g, '').replaceAll('\\"', '"'));
        }
    }
    form.entrypoint = [];
    if (form.entrypointStr) {
        let itemEntrypoint = splitStringIgnoringQuotes(form.entrypointStr);
        for (const item of itemEntrypoint) {
            form.entrypoint.push(item.replace(/(?<!\\)"/g, '').replaceAll('\\"', '"'));
        }
    }
    if (form.publishAllPorts) {
        form.exposedPorts = [];
    } else {
        if (!checkPortValid()) {
            return;
        }
    }
    form.memory = Number(form.memory);
    form.nanoCPUs = Number(form.nanoCPUs);

    loading.value = true;
    if (isCreate.value) {
        await createContainer(form)
            .then(() => {
                loading.value = false;
                openTaskLog(form.taskID);
            })
            .catch(() => {
                loading.value = false;
            });
    } else {
        await updateContainer(form)
            .then(() => {
                loading.value = false;
                openTaskLog(form.taskID);
            })
            .catch(() => {
                loading.value = false;
            });
    }
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const checkPortValid = () => {
    if (form.exposedPorts.length === 0) {
        return true;
    }
    for (const port of form.exposedPorts) {
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

const checkExist = async () => {
    let params = {
        page: 1,
        pageSize: 1,
        state: 'all',
        name: form.name,
        filters: '',
        orderBy: 'createdAt',
        order: 'null',
    };
    await searchContainer(params).then((res) => {
        if (res.data.items?.length === 1) {
            isCreate.value = false;
            routerToNameWithQuery('ContainerCreate', { name: form.name, t: Date.now() });
            return;
        } else {
            isCreate.value = true;
        }
    });
};

onMounted(() => {
    if (router.currentRoute.value.query.name) {
        isCreate.value = false;
        form.name = String(router.currentRoute.value.query.name);
    } else {
        isCreate.value = true;
    }
    search();
});
</script>

<style lang="scss" scoped>
.widthClass {
    width: 100%;
}
.el-card {
    border: 1px solid var(--el-border-color-light) !important;
}
</style>
