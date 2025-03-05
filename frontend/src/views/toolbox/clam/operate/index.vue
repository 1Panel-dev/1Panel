<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="title"
        size="large"
        :resource="dialogData.title === 'add' ? '' : dialogData.rowData?.name"
        @close="handleClose"
    >
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input
                            :disabled="dialogData.title === 'edit'"
                            clearable
                            v-model.trim="dialogData.rowData!.name"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('toolbox.clam.scanDir')" prop="path">
                        <el-input v-model="dialogData.rowData!.path">
                            <template #prepend>
                                <FileList @choose="loadDir" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item :label="$t('toolbox.clam.infectedStrategy')" prop="infectedStrategy">
                        <el-radio-group v-model="dialogData.rowData!.infectedStrategy">
                            <el-radio value="none">{{ $t('toolbox.clam.none') }}</el-radio>
                            <el-radio value="remove">{{ $t('commons.button.delete') }}</el-radio>
                            <el-radio value="move">{{ $t('toolbox.clam.move') }}</el-radio>
                            <el-radio value="copy">{{ $t('commons.button.copy') }}</el-radio>
                        </el-radio-group>
                        <span class="input-help">
                            {{ $t('toolbox.clam.' + dialogData.rowData!.infectedStrategy + 'Helper') }}
                        </span>
                    </el-form-item>
                    <el-form-item v-if="hasInfectedDir()" :label="$t('toolbox.clam.infectedDir')" prop="infectedDir">
                        <el-input v-model="dialogData.rowData!.infectedDir">
                            <template #prepend>
                                <FileList @choose="loadInfectedDir" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item prop="hasSpec">
                        <el-checkbox v-model="dialogData.rowData!.hasSpec" :label="$t('toolbox.clam.cron')" />
                    </el-form-item>
                    <el-form-item prop="spec" v-if="dialogData.rowData!.hasSpec && isProductPro">
                        <div class="grid sm:grid-cols-4 gap-4 grid-cols-1">
                            <el-select v-model="dialogData.rowData!.specObj.specType" @change="changeSpecType()">
                                <el-option
                                    v-for="item in specOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                            <el-select
                                v-if="dialogData.rowData!.specObj.specType === 'perWeek'"
                                v-model="dialogData.rowData!.specObj.week"
                            >
                                <el-option
                                    v-for="item in weekOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                            <el-input
                                v-if="hasDay(dialogData.rowData!.specObj)"
                                v-model.number="dialogData.rowData!.specObj.day"
                            >
                                <template #append>
                                    <div class="sm:min-w-8 min-w-14 text-center">
                                        <el-tooltip :content="$t('commons.units.day')" placement="top">
                                            {{ $t('commons.units.dayUnit') }}
                                        </el-tooltip>
                                    </div>
                                </template>
                            </el-input>
                            <el-input
                                v-if="hasHour(dialogData.rowData!.specObj)"
                                v-model.number="dialogData.rowData!.specObj.hour"
                            >
                                <template #append>
                                    <div class="sm:min-w-8 min-w-14 text-center">
                                        <el-tooltip :content="$t('commons.units.hour')" placement="top">
                                            {{ $t('commons.units.hourUnit') }}
                                        </el-tooltip>
                                    </div>
                                </template>
                            </el-input>
                            <el-input
                                v-if="dialogData.rowData!.specObj.specType !== 'perNSecond'"
                                v-model.number="dialogData.rowData!.specObj.minute"
                            >
                                <template #append>
                                    <div class="sm:min-w-8 min-w-14 text-center">
                                        <el-tooltip :content="$t('commons.units.minute')" placement="top">
                                            {{ $t('commons.units.minuteUnit') }}
                                        </el-tooltip>
                                    </div>
                                </template>
                            </el-input>
                            <el-input
                                v-if="dialogData.rowData!.specObj.specType === 'perNSecond'"
                                v-model.number="dialogData.rowData!.specObj.second"
                            >
                                <template #append>
                                    <div class="sm:min-w-8 min-w-14 text-center">
                                        <el-tooltip :content="$t('commons.units.second')" placement="top">
                                            {{ $t('commons.units.secondUnit') }}
                                        </el-tooltip>
                                    </div>
                                </template>
                            </el-input>
                        </div>
                    </el-form-item>
                    <div v-if="globalStore.isIntl">
                        <el-form-item v-if="(dialogData.rowData!.hasSpec) && !isProductPro">
                            <span>{{ $t('toolbox.clam.alertHelper') }}</span>
                            <el-button link type="primary" @click="toUpload">
                                {{ $t('license.levelUpPro') }}
                            </el-button>
                        </el-form-item>
                    </div>
                    <div v-if="!globalStore.isIntl">
                        <el-form-item prop="hasAlert">
                            <el-checkbox v-model="dialogData.rowData!.hasAlert" :label="$t('xpack.alert.isAlert')" />
                            <span class="input-help">{{ $t('xpack.alert.clamHelper') }}</span>
                        </el-form-item>
                        <el-form-item
                            v-if="(dialogData.rowData!.hasAlert || dialogData.rowData!.hasSpec) && !isProductPro"
                        >
                            <span>{{ $t('toolbox.clam.alertHelper') }}</span>
                            <el-button link type="primary" @click="toUpload">
                                {{ $t('license.levelUpPro') }}
                            </el-button>
                        </el-form-item>
                        <el-form-item
                            prop="alertCount"
                            v-if="dialogData.rowData!.hasAlert && isProductPro"
                            :label="$t('xpack.alert.alertCount')"
                        >
                            <el-input-number
                                style="width: 200px"
                                :min="1"
                                step-strictly
                                :step="1"
                                v-model.number="dialogData.rowData!.alertCount"
                            ></el-input-number>
                            <span class="input-help">{{ $t('xpack.alert.alertCountHelper') }}</span>
                        </el-form-item>
                    </div>
                    <el-form-item :label="$t('commons.table.description')" prop="description">
                        <el-input type="textarea" :rows="3" clearable v-model="dialogData.rowData!.description" />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
        <LicenseImport ref="licenseRef" />
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import LicenseImport from '@/components/license-import/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Toolbox } from '@/api/interface/toolbox';
import { createClam, updateClam } from '@/api/modules/toolbox';
import { storeToRefs } from 'pinia';
import { GlobalStore } from '@/store';
import { specOptions, transObjToSpec, transSpecToObj, weekOptions } from '@/views/cronjob/cronjob/helper';

const globalStore = GlobalStore();
const licenseRef = ref();
const { isProductPro } = storeToRefs(globalStore);
interface DialogProps {
    title: string;
    rowData?: Toolbox.ClamInfo;
    getTableList?: () => Promise<any>;
}
const loading = ref();
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (dialogData.value.rowData?.spec) {
        dialogData.value.rowData.hasSpec = true;
        dialogData.value.rowData.specObj = transSpecToObj(dialogData.value.rowData.spec);
    } else {
        dialogData.value.rowData.specObj = {
            specType: 'perDay',
            week: 1,
            day: 3,
            hour: 1,
            minute: 30,
            second: 30,
        };
    }
    dialogData.value.rowData.hasAlert = dialogData.value.rowData!.alertCount > 0;
    dialogData.value.rowData!.alertCount = dialogData.value.rowData!.alertCount || 3;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const verifySpec = (rule: any, value: any, callback: any) => {
    let item = dialogData.value.rowData!.specObj;
    if (
        !Number.isInteger(item.day) ||
        !Number.isInteger(item.hour) ||
        !Number.isInteger(item.minute) ||
        !Number.isInteger(item.second) ||
        !Number.isInteger(item.week)
    ) {
        callback(new Error(i18n.global.t('cronjob.specErr')));
        return;
    }
    switch (item.specType) {
        case 'perMonth':
            if (
                item.day < 0 ||
                item.day > 31 ||
                item.hour < 0 ||
                item.hour > 23 ||
                item.minute < 0 ||
                item.minute > 59
            ) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
            break;
        case 'perNDay':
            if (
                item.day < 0 ||
                item.day > 366 ||
                item.hour < 0 ||
                item.hour > 23 ||
                item.minute < 0 ||
                item.minute > 59
            ) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
            break;
        case 'perWeek':
            if (
                item.week < 0 ||
                item.week > 6 ||
                item.hour < 0 ||
                item.hour > 23 ||
                item.minute < 0 ||
                item.minute > 59
            ) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
            break;
        case 'perDay':
            if (item.hour < 0 || item.hour > 23 || item.minute < 0 || item.minute > 59) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
            break;
        case 'perNHour':
            if (item.hour < 0 || item.hour > 8784 || item.minute < 0 || item.minute > 59) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
            break;
        case 'perHour':
            if (item.minute < 0 || item.minute > 59) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
        case 'perNMinute':
            if (item.minute < 0 || item.minute > 527040) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
            break;
        case 'perNSecond':
            if (item.second < 0 || item.second > 31622400) {
                callback(new Error(i18n.global.t('cronjob.specErr')));
                return;
            }
            break;
    }
    callback();
};

const checkSendCount = (rule: any, value: any, callback: any) => {
    if (value === '') {
        callback();
    }
    const regex = /^(?:[1-9]|[12][0-9]|30)$/;
    if (!regex.test(value)) {
        return callback(new Error(i18n.global.t('commons.rule.numberRange', [1, 30])));
    }
    callback();
};
const rules = reactive({
    name: [Rules.simpleName],
    path: [Rules.requiredInput, Rules.noSpace],
    spec: [
        { validator: verifySpec, trigger: 'blur', required: true },
        { validator: verifySpec, trigger: 'change', required: true },
    ],
    alertCount: [Rules.integerNumber, { validator: checkSendCount, trigger: 'blur' }],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const hasInfectedDir = () => {
    return (
        dialogData.value.rowData!.infectedStrategy === 'move' || dialogData.value.rowData!.infectedStrategy === 'copy'
    );
};
const loadDir = async (path: string) => {
    dialogData.value.rowData!.path = path;
};
const loadInfectedDir = async (path: string) => {
    dialogData.value.rowData!.infectedDir = path;
};
const hasDay = (item: any) => {
    return item.specType === 'perMonth' || item.specType === 'perNDay';
};
const hasHour = (item: any) => {
    return item.specType !== 'perHour' && item.specType !== 'perNMinute' && item.specType !== 'perNSecond';
};

const toUpload = () => {
    licenseRef.value.acceptParams();
};

const changeSpecType = () => {
    let item = dialogData.value.rowData!.specObj;
    switch (item.specType) {
        case 'perMonth':
        case 'perNDay':
            item.day = 3;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'perWeek':
            item.week = 1;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'perDay':
        case 'perNHour':
            item.hour = 2;
            item.minute = 30;
            break;
        case 'perHour':
        case 'perNMinute':
            item.minute = 30;
            break;
        case 'perNSecond':
            item.second = 30;
            break;
    }
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        let spec = '';
        let item = dialogData.value.rowData.specObj;
        if (dialogData.value.rowData!.hasSpec) {
            spec = transObjToSpec(item.specType, item.week, item.day, item.hour, item.minute, item.second);
            if (spec === '') {
                MsgError(i18n.global.t('cronjob.cronSpecHelper'));
                return;
            }
        }
        dialogData.value.rowData.spec = spec;
        if (dialogData.value.rowData!.hasAlert) {
            dialogData.value.rowData.alertCount = dialogData.value.rowData!.hasAlert
                ? dialogData.value.rowData.alertCount
                : 0;
            dialogData.value.rowData.alertTitle = i18n.global.t('toolbox.clam.alertTitle', [
                dialogData.value.rowData.name,
            ]);
        } else {
            dialogData.value.rowData.alertTitle = '';
            dialogData.value.rowData.alertCount = 0;
            dialogData.value.rowData.hasAlert = false;
        }

        if (dialogData.value.title === 'edit') {
            await updateClam(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    drawerVisible.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });

            return;
        }

        await createClam(dialogData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
