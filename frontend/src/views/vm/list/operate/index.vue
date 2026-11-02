<template>
    <DrawerPro v-model="formVisible" :header="formTitle" size="large" :auto-close="!submitLoading" @close="resetForm">
        <el-form ref="formRef" v-loading="optionsLoading" label-position="top" :model="form" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" :disabled="formMode === 'edit'" />
            </el-form-item>

            <el-form-item v-if="formMode === 'create'">
                <el-checkbox v-model="createFromTemplate" @change="changeCreateFromTemplate">
                    {{ $t('xpack.vm.createFromTemplate') }}
                </el-checkbox>
            </el-form-item>
            <el-form-item
                v-if="formMode === 'create' && createFromTemplate"
                :label="$t('xpack.vm.template')"
                prop="templateID"
            >
                <el-select v-model="form.templateID" class="full-width" filterable @change="changeTemplate">
                    <el-option
                        v-for="item in availableTemplateOptions"
                        :key="item.id"
                        :label="`${item.name} (${computeSizeFromMB(item.memory)}, ${formatDiskSize(item.diskSize)})`"
                        :value="item.id"
                    />
                </el-select>
            </el-form-item>

            <el-form-item v-if="!createFromTemplate" :label="$t('xpack.vm.iso')" prop="isoPaths">
                <el-select
                    v-model="form.isoPaths"
                    multiple
                    class="full-width"
                    filterable
                    clearable
                    default-first-option
                    @change="changeISO"
                >
                    <el-option
                        v-for="item in isoOptions"
                        :key="item.id"
                        :label="item.name + ' (' + item.path + ')'"
                        :value="item.path"
                    />
                </el-select>
            </el-form-item>
            <div v-if="!createFromTemplate" class="resource-create-helper">
                <span>{{ $t('xpack.vm.isoHelper') }}</span>
                <el-button link type="primary" @click="isoOperateRef?.acceptParams()">
                    {{ $t('commons.button.create') }}
                </el-button>
            </div>
            <div
                v-if="formMode === 'create' && !createFromTemplate && form.osType === 'windows'"
                class="resource-create-helper"
            >
                <span>
                    {{ $t('xpack.vm.windowsDriverHelper') }}
                    <span v-if="form.diskBus === 'virtio'">{{ $t('xpack.vm.windowsDiskDriverHelper') }}</span>
                </span>
            </div>

            <template v-if="formMode === 'create'">
                <el-form-item v-if="!createFromTemplate" :label="$t('xpack.vm.osType')" prop="osType">
                    <el-select v-model="form.osType" class="full-width" @change="changeOSType">
                        <el-option label="Windows" value="windows" />
                        <el-option label="Linux" value="linux" />
                        <el-option :label="$t('xpack.vm.osOther')" value="other" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('xpack.vm.diskBus')" prop="diskBus">
                    <el-select v-model="form.diskBus" class="full-width" :disabled="createFromTemplate">
                        <el-option label="SATA (AHCI)" value="sata" />
                        <el-option label="VirtIO" value="virtio" />
                    </el-select>
                </el-form-item>
            </template>

            <el-form-item :label="$t('xpack.vm.network')" prop="network">
                <el-select v-model="form.network" class="full-width" filterable clearable default-first-option>
                    <el-option
                        v-for="item in networkOptions"
                        :key="item.id"
                        :label="`${item.name} (${item.dhcpStart} - ${item.dhcpEnd})`"
                        :value="item.name"
                        :disabled="isVMResourceUnavailable(item.status)"
                    />
                </el-select>
            </el-form-item>
            <div class="resource-create-helper">
                <span>{{ $t('xpack.vm.networkHelper') }}</span>
                <el-button link type="primary" @click="networkOperateRef?.acceptParams()">
                    {{ $t('commons.button.create') }}
                </el-button>
            </div>
            <el-form-item v-if="formMode === 'create'" :label="$t('xpack.vm.storagePool')" prop="storagePool">
                <el-select v-model="form.storagePool" class="full-width" filterable clearable default-first-option>
                    <el-option
                        v-for="item in storagePoolOptions"
                        :key="item.id"
                        :label="item.path ? `${item.name} (${item.path})` : item.name"
                        :value="item.name"
                        :disabled="isVMResourceUnavailable(item.status)"
                    />
                </el-select>
            </el-form-item>
            <div v-if="formMode === 'create'" class="resource-create-helper">
                <span>{{ $t('xpack.vm.storagePoolHelper') }}</span>
                <el-button link type="primary" @click="storageOperateRef?.acceptParams()">
                    {{ $t('commons.button.create') }}
                </el-button>
            </div>
            <el-form-item v-else :label="$t('xpack.vm.bootOrder')" prop="bootOrder">
                <el-select v-model="form.bootOrder" class="full-width">
                    <el-option :label="$t('xpack.vm.bootDiskFirst')" value="hd,cdrom" />
                    <el-option
                        :label="$t('xpack.vm.bootIsoFirst')"
                        value="cdrom,hd"
                        :disabled="!form.isoPaths.length"
                    />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('xpack.vm.cpu')" prop="cpu">
                <el-input v-model.number="form.cpu" class="p-w-300" :min="1" :max="maxCpu || 256">
                    <template #append>
                        <div style="width: 40px">{{ $t('commons.units.core') }}</div>
                    </template>
                </el-input>
                <span v-if="cpuLimitText" class="input-help" :class="{ 'is-warning': cpuLimitExceeded }">
                    {{ cpuLimitText }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('xpack.vm.memory')" prop="memory">
                <el-input
                    v-model.number="form.memory"
                    class="p-w-300"
                    :min="minMemorySize"
                    :max="maxMemorySize || undefined"
                >
                    <template #append>
                        <el-select v-model="memoryUnit" class="p-w-80" @change="changeMemoryUnit">
                            <el-option label="MiB" value="MiB" />
                            <el-option label="GiB" value="GiB" />
                        </el-select>
                    </template>
                </el-input>
                <span v-if="memoryLimitText" class="input-help" :class="{ 'is-warning': memoryLimitExceeded }">
                    {{ memoryLimitText }}
                </span>
            </el-form-item>

            <template v-if="formMode === 'create'">
                <el-form-item :label="$t('xpack.vm.diskSize')" prop="diskSize">
                    <el-input
                        v-model.number="form.diskSize"
                        class="p-w-300"
                        :min="minDiskSize"
                        :max="maxDiskSize || undefined"
                    >
                        <template #append>
                            <el-select v-model="diskUnit" class="p-w-80" @change="changeDiskUnit">
                                <el-option label="MiB" value="MiB" />
                                <el-option label="GiB" value="GiB" />
                                <el-option label="TiB" value="TiB" />
                            </el-select>
                        </template>
                    </el-input>
                    <span v-if="storageLimitText" class="input-help" :class="{ 'is-warning': diskLimitExceeded }">
                        {{ storageLimitText }}
                    </span>
                </el-form-item>
            </template>
            <el-form-item>
                <el-checkbox v-if="formMode === 'create'" v-model="form.startAfterCreate">
                    {{ $t('xpack.vm.startAfterCreate') }}
                </el-checkbox>
                <el-checkbox v-model="form.autoStart">
                    {{ $t('xpack.vm.autoStart') }}
                </el-checkbox>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="submitLoading" @click="formVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button
                v-permission
                type="primary"
                :disabled="submitLoading || optionsLoading || props.resourceLimitLoading || selectedResourceUnavailable"
                @click="submitForm"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <IsoOperate ref="isoOperateRef" @search="loadIsoOptions" />
    <NetworkOperate ref="networkOperateRef" @search="loadNetworkOptions" />
    <StorageOperate ref="storageOperateRef" @search="loadStorageOptions" />
    <div v-if="formVisible && submitLoading" v-loading="submitLoading" class="vm-drawer-loading-mask"></div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { newUUID } from '@/utils/id';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { VM } from '@/api/interface/vm';
import { isVMMemoryAllocated, isVMResourceUnavailable } from '@/utils/vm';
import { createVM, listVMIso, listVMNetworks, listVMStorages, listVMTemplates, updateVM } from '@/api/modules/vm';
import { Rules } from '@/global/form-rules';
import { computeSize, computeSizeFromMB, convertBinarySize, type BinarySizeUnit } from '@/utils/size';
import IsoOperate from '@/views/vm/iso/create/index.vue';
import NetworkOperate from '@/views/vm/network/create/index.vue';
import StorageOperate from '@/views/vm/storage/create/index.vue';

type FormMode = 'create' | 'edit';
type SizeUnit = Extract<BinarySizeUnit, 'MiB' | 'GiB' | 'TiB'>;

const BYTE_PER_GIB = 1024 * 1024 * 1024;
const MIN_MEMORY_MIB = 128;
const MIN_DISK_GIB = 1;

interface ResourceLimit {
    cpuTotal: number;
    memoryTotal: number;
    memoryAllocated: number;
    memoryReserved: number;
}

const props = defineProps<{
    resourceLimit?: ResourceLimit;
    resourceLimitLoading?: boolean;
}>();

const emit = defineEmits(['open-log']);

const submitLoading = ref(false);
const optionsLoading = ref(false);
const formRef = ref<FormInstance>();
const formVisible = ref(false);
const formMode = ref<FormMode>('create');
const createFromTemplate = ref(false);
const memoryUnit = ref<SizeUnit>('GiB');
const diskUnit = ref<SizeUnit>('GiB');
const lastMemoryUnit = ref<SizeUnit>('GiB');
const lastDiskUnit = ref<SizeUnit>('GiB');
const originalMemoryMiB = ref(0);
const originalBootOrder = ref('hd');

const networkOptions = ref<VM.NetworkItem[]>([]);
const isoOptions = ref<VM.IsoItem[]>([]);
const storagePoolOptions = ref<VM.StorageItem[]>([]);
const templateOptions = ref<VM.TemplateItem[]>([]);
const isoOperateRef = ref<InstanceType<typeof IsoOperate>>();
const networkOperateRef = ref<InstanceType<typeof NetworkOperate>>();
const storageOperateRef = ref<InstanceType<typeof StorageOperate>>();

const form = reactive<VM.Create>({
    name: '',
    cpu: 2,
    memory: 2,
    diskSize: 20,
    osType: 'linux',
    diskBus: 'virtio',
    templateID: undefined,
    diskPath: '',
    storagePool: 'default',
    isoPaths: [],
    imagePath: '',
    network: 'default',
    bootOrder: 'hd',
    startAfterCreate: true,
    autoStart: false,
    description: '',
});

const rules = computed<FormRules>(() => ({
    name: [Rules.requiredInput, Rules.vmName],
    cpu: [Rules.requiredInput, Rules.integerNumber],
    memory: [Rules.requiredInput, Rules.integerNumber],
    diskSize: [Rules.requiredInput, Rules.integerNumber],
    osType: [Rules.requiredSelect],
    diskBus: [Rules.requiredSelect],
    templateID: [Rules.requiredSelect],
    isoPaths:
        formMode.value === 'create' && !createFromTemplate.value ? [{ ...Rules.requiredSelect, type: 'array' }] : [],
    network: [Rules.requiredInput, Rules.vmNetwork],
    storagePool: [Rules.requiredSelect],
}));

const formTitle = computed(() =>
    formMode.value === 'create' ? i18n.global.t('commons.button.create') : i18n.global.t('commons.button.edit'),
);
const selectedStoragePool = computed(() => storagePoolOptions.value.find((item) => item.name === form.storagePool));
const selectedResourceUnavailable = computed(
    () =>
        isVMResourceUnavailable(networkOptions.value.find((item) => item.name === form.network)?.status) ||
        (formMode.value === 'create' && isVMResourceUnavailable(selectedStoragePool.value?.status)),
);
const selectedTemplate = computed(() => templateOptions.value.find((item) => item.id === form.templateID));
const availableTemplateOptions = computed(() => templateOptions.value.filter((item) => item.isExist));
const diskAvailableBytes = computed(() => selectedStoragePool.value?.available || 0);
const maxCpu = computed(() => Math.floor(props.resourceLimit?.cpuTotal || 0));
const cpuLimitExceeded = computed(() => maxCpu.value > 0 && Number(form.cpu || 0) > maxCpu.value);
const requestMemoryBytes = computed(() => toMiB(form.memory, memoryUnit.value) * 1024 * 1024);
const originalMemoryBytes = computed(() => originalMemoryMiB.value * 1024 * 1024);
const memoryLimitBytes = computed(() =>
    Math.max((props.resourceLimit?.memoryTotal || 0) - (props.resourceLimit?.memoryReserved || 0), 0),
);
const memoryAvailableBytes = computed(() =>
    Math.max(memoryLimitBytes.value - (props.resourceLimit?.memoryAllocated || 0) + originalMemoryBytes.value, 0),
);
const memoryLimitExceeded = computed(
    () => memoryLimitBytes.value > 0 && requestMemoryBytes.value > memoryAvailableBytes.value,
);
const minMemorySize = computed(() => (memoryUnit.value === 'MiB' ? MIN_MEMORY_MIB : 1));
const maxMemorySize = computed(() => {
    if (!memoryAvailableBytes.value) {
        return 0;
    }
    return Math.floor(convertBinarySize(memoryAvailableBytes.value, 'B', memoryUnit.value, 6));
});
const templateMinDiskSizeGiB = computed(() => (createFromTemplate.value ? selectedTemplate.value?.diskSize || 0 : 0));
const minDiskSize = computed(() => {
    const fallback = convertBinarySize(MIN_DISK_GIB, 'GiB', diskUnit.value, 6);
    if (!templateMinDiskSizeGiB.value) {
        return fallback;
    }
    return Math.max(convertBinarySize(templateMinDiskSizeGiB.value, 'GiB', diskUnit.value, 6), fallback);
});
const maxDiskSize = computed(() => {
    if (formMode.value !== 'create' || !diskAvailableBytes.value) {
        return 0;
    }
    return Math.floor(convertBinarySize(diskAvailableBytes.value, 'B', diskUnit.value, 6));
});
const requestDiskBytes = computed(() => toGiB(form.diskSize, diskUnit.value) * BYTE_PER_GIB);
const diskLimitExceeded = computed(
    () =>
        formMode.value === 'create' &&
        diskAvailableBytes.value > 0 &&
        requestDiskBytes.value > diskAvailableBytes.value,
);
const cpuLimitText = computed(() => {
    if (!maxCpu.value) {
        return '';
    }
    return i18n.global.t('xpack.vm.cpuLimitHelper', [maxCpu.value]);
});
const memoryLimitText = computed(() => {
    if (!memoryLimitBytes.value) {
        return '';
    }
    return i18n.global.t('xpack.vm.memoryLimitHelper', [
        computeSize(memoryAvailableBytes.value),
        computeSize(props.resourceLimit?.memoryReserved || 0),
    ]);
});
const storageLimitText = computed(() => {
    if (formMode.value !== 'create' || !diskAvailableBytes.value) {
        return '';
    }
    return i18n.global.t('xpack.vm.storageLimitHelper', [computeSize(diskAvailableBytes.value)]);
});

const loadResourceOptions = async (includeTemplates = false) => {
    optionsLoading.value = true;
    try {
        const requests = [
            listVMNetworks(),
            listVMStorages(),
            listVMIso(),
            ...(includeTemplates ? [listVMTemplates()] : []),
        ] as const;
        const [networkRes, storageRes, isoRes, templateRes] = await Promise.all(requests);
        networkOptions.value = networkRes.data || [];
        isoOptions.value = isoRes.data || [];
        storagePoolOptions.value = storageRes.data || [];
        templateOptions.value = templateRes?.data || [];
    } finally {
        optionsLoading.value = false;
    }
};

const loadIsoOptions = async () => {
    const res = await listVMIso();
    isoOptions.value = res.data || [];
};

const loadNetworkOptions = async () => {
    const res = await listVMNetworks();
    networkOptions.value = res.data || [];
};

const loadStorageOptions = async () => {
    const res = await listVMStorages();
    storagePoolOptions.value = res.data || [];
};

const openCreate = () => {
    loadResourceOptions(true);
    resetForm();
    formMode.value = 'create';
    formVisible.value = true;
};

const openEdit = (row: VM.Item) => {
    loadResourceOptions();
    resetForm();
    formMode.value = 'edit';
    form.name = row.name;
    form.cpu = row.cpu || 1;
    originalMemoryMiB.value = isVMMemoryAllocated(row.status) ? row.memory || 0 : 0;
    setMemoryFromMiB(row.memory || 2048);
    form.isoPaths = [...(row.isoPaths || [])];
    form.network = row.network || 'default';
    form.bootOrder = normalizeBootOrder(row.bootOrder, form.isoPaths);
    originalBootOrder.value = form.bootOrder;
    form.autoStart = row.autoStart || false;
    form.description = row.description || '';
    formVisible.value = true;
};

const submitForm = async () => {
    if (!formRef.value) {
        return;
    }
    if (optionsLoading.value || props.resourceLimitLoading || selectedResourceUnavailable.value) {
        return;
    }
    await formRef.value.validate();
    if (cpuLimitExceeded.value) {
        MsgError(i18n.global.t('xpack.vm.cpuLimitExceeded', [maxCpu.value]));
        return;
    }
    if (memoryLimitExceeded.value) {
        MsgError(i18n.global.t('xpack.vm.memoryLimitExceeded', [computeSize(memoryAvailableBytes.value)]));
        return;
    }
    if (toMiB(form.memory, memoryUnit.value) < MIN_MEMORY_MIB) {
        MsgError(`${i18n.global.t('xpack.vm.memory')} >= ${MIN_MEMORY_MIB} MiB`);
        return;
    }
    if (createFromTemplate.value && toGiB(form.diskSize, diskUnit.value) < templateMinDiskSizeGiB.value) {
        MsgError(`${i18n.global.t('xpack.vm.diskSize')} >= ${formatDiskSize(templateMinDiskSizeGiB.value)}`);
        return;
    }
    if (formMode.value === 'create') {
        const diskSizeGiB = toGiB(form.diskSize, diskUnit.value);
        if (diskSizeGiB < MIN_DISK_GIB) {
            MsgError(`${i18n.global.t('xpack.vm.diskSize')} >= ${formatDiskSize(MIN_DISK_GIB)}`);
            return;
        }
        if (diskLimitExceeded.value) {
            MsgError(i18n.global.t('xpack.vm.storageLimitExceeded', [computeSize(diskAvailableBytes.value)]));
            return;
        }
    }
    if (formMode.value === 'edit') {
        form.bootOrder = normalizeBootOrder(form.bootOrder, form.isoPaths);
        if (form.bootOrder !== originalBootOrder.value) {
            await ElMessageBox.confirm(
                i18n.global.t('xpack.vm.bootOrderConfirm', [formatBootOrder(form.bootOrder)]),
                i18n.global.t('xpack.vm.bootOrder'),
            );
        }
    }
    submitLoading.value = true;
    try {
        const taskID = newUUID();
        const memoryMiB = toMiB(form.memory, memoryUnit.value);
        const diskSizeGiB = toGiB(form.diskSize, diskUnit.value);
        const res =
            formMode.value === 'create'
                ? await createVM({
                      ...form,
                      taskID,
                      memory: memoryMiB,
                      diskSize: diskSizeGiB,
                      diskPath: form.diskPath || '',
                      templateID: createFromTemplate.value ? form.templateID : undefined,
                      isoPaths: createFromTemplate.value ? [] : form.isoPaths,
                      imagePath: '',
                      bootOrder: createFromTemplate.value ? 'hd' : 'hd,cdrom',
                  })
                : await updateVM({
                      taskID,
                      name: form.name,
                      cpu: form.cpu,
                      memory: toMiB(form.memory, memoryUnit.value),
                      isoPaths: form.isoPaths,
                      network: form.network,
                      bootOrder: form.bootOrder,
                      autoStart: form.autoStart,
                      description: form.description,
                  });
        formVisible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('open-log', res.data?.taskID || taskID);
    } finally {
        submitLoading.value = false;
    }
};

const resetForm = () => {
    form.name = '';
    form.cpu = 2;
    form.memory = 2;
    originalMemoryMiB.value = 0;
    originalBootOrder.value = 'hd';
    form.diskSize = 20;
    form.osType = 'linux';
    form.diskBus = 'virtio';
    createFromTemplate.value = false;
    setUnit(memoryUnit, lastMemoryUnit, 'GiB');
    setUnit(diskUnit, lastDiskUnit, 'GiB');
    form.templateID = undefined;
    form.diskPath = '';
    form.storagePool = 'default';
    form.isoPaths = [];
    form.imagePath = '';
    form.network = 'default';
    form.bootOrder = 'hd';
    form.startAfterCreate = true;
    form.autoStart = false;
    form.description = '';
    formRef.value?.clearValidate();
};

const changeCreateFromTemplate = () => {
    form.templateID = undefined;
    form.imagePath = '';
    if (!createFromTemplate.value) {
        changeOSType();
        return;
    }
    form.isoPaths = [];
    form.bootOrder = 'hd';
    const first = availableTemplateOptions.value[0];
    if (first) {
        form.templateID = first.id;
        applyTemplate(first);
    }
};

const changeTemplate = () => {
    if (selectedTemplate.value) {
        applyTemplate(selectedTemplate.value);
    }
};

const applyTemplate = (item: VM.TemplateItem) => {
    form.cpu = item.cpu || form.cpu;
    setMemoryFromMiB(item.memory || 2048);
    form.diskSize = item.diskSize || form.diskSize;
    form.diskBus = item.diskBus === 'sata' ? 'sata' : 'virtio';
    setUnit(diskUnit, lastDiskUnit, 'GiB');
    form.storagePool = item.storagePool || form.storagePool;
    form.network = item.network || form.network;
    form.bootOrder = createFromTemplate.value ? 'hd' : normalizeBootOrder(item.bootOrder, form.isoPaths);
};

const changeOSType = () => {
    form.diskBus = form.osType === 'windows' ? 'sata' : 'virtio';
};

const normalizeBootOrder = (bootOrder: string | undefined, isoPaths: string[]) => {
    return isoPaths.length && String(bootOrder || '').startsWith('cdrom') ? 'cdrom,hd' : 'hd,cdrom';
};

const changeISO = () => {
    form.bootOrder = normalizeBootOrder(form.bootOrder, form.isoPaths);
    if (formMode.value !== 'create' || createFromTemplate.value) {
        return;
    }
    const iso = isoOptions.value.find((item) => item.path === form.isoPaths[0]);
    if (!iso) {
        return;
    }
    const type = iso.type.trim().toLowerCase();
    const linuxTypes = [
        'linux',
        'debian',
        'ubuntu',
        'centos',
        'rocky',
        'almalinux',
        'fedora',
        'opensuse',
        'rhel',
        'oraclelinux',
        'archlinux',
    ];
    const osType = type === 'windows' ? 'windows' : linuxTypes.includes(type) ? 'linux' : 'other';
    if (form.osType !== osType) {
        form.osType = osType;
        changeOSType();
    }
};

const formatBootOrder = (bootOrder: string) => {
    return bootOrder.startsWith('cdrom')
        ? i18n.global.t('xpack.vm.bootIsoFirst')
        : i18n.global.t('xpack.vm.bootDiskFirst');
};

const setUnit = (target: typeof memoryUnit, lastTarget: typeof lastMemoryUnit, unit: SizeUnit) => {
    target.value = unit;
    lastTarget.value = unit;
};

const convertDisplaySize = (value: number, from: SizeUnit, to: SizeUnit) => {
    if (!value || from === to) {
        return value;
    }
    return convertBinarySize(value, from, to);
};

const toMiB = (value: number, unit: SizeUnit) => {
    return Math.ceil(convertBinarySize(Number(value) || 0, unit, 'MiB', 6));
};

const toGiB = (value: number, unit: SizeUnit) => {
    return Math.ceil(convertBinarySize(Number(value) || 0, unit, 'GiB', 6));
};

const formatDiskSize = (size: number) => {
    if (!size) {
        return '-';
    }
    return `${size} GiB`;
};

const normalizeSizeUnit = (unit: string | number | boolean | undefined): SizeUnit => {
    const target = String(unit);
    return ['MiB', 'GiB', 'TiB'].includes(target) ? (target as SizeUnit) : 'GiB';
};

const changeMemoryUnit = (unit: string | number | boolean | undefined) => {
    const target = normalizeSizeUnit(unit);
    form.memory = convertDisplaySize(form.memory, lastMemoryUnit.value, target);
    lastMemoryUnit.value = target;
};

const changeDiskUnit = (unit: string | number | boolean | undefined) => {
    const target = normalizeSizeUnit(unit);
    form.diskSize = convertDisplaySize(form.diskSize, lastDiskUnit.value, target);
    lastDiskUnit.value = target;
};

const setMemoryFromMiB = (memoryMiB: number) => {
    if (memoryMiB % 1024 === 0) {
        setUnit(memoryUnit, lastMemoryUnit, 'GiB');
        form.memory = convertBinarySize(memoryMiB, 'MiB', 'GiB');
        return;
    }
    setUnit(memoryUnit, lastMemoryUnit, 'MiB');
    form.memory = memoryMiB;
};

defineExpose({
    openCreate,
    openEdit,
});
</script>

<style scoped lang="scss">
.full-width {
    width: 100%;
}

.input-help.is-warning {
    color: var(--el-color-danger);
}

.resource-create-helper {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: -14px 0 18px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-drawer-loading-mask {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 3000;
    width: 50%;
    background: var(--el-mask-color);
}
</style>
