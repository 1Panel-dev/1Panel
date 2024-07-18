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
                :header="title"
                :hideResource="dialogData.title === 'add'"
                :resource="dialogData.rowData?.name"
                :back="handleClose"
            />
        </template>
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
                            <el-radio value="remove">{{ $t('toolbox.clam.remove') }}</el-radio>
                            <el-radio value="move">{{ $t('toolbox.clam.move') }}</el-radio>
                            <el-radio value="copy">{{ $t('toolbox.clam.copy') }}</el-radio>
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
                    <el-form-item v-if="dialogData.rowData!.hasSpec && !isProductPro">
                        <span>{{ $t('toolbox.clam.cronHelper') }}</span>
                        <el-button link type="primary" @click="toUpload">
                            {{ $t('license.levelUpPro') }}
                        </el-button>
                    </el-form-item>
                    <el-form-item prop="spec" v-if="dialogData.rowData!.hasSpec && isProductPro">
                        <el-select
                            class="specTypeClass"
                            v-model="dialogData.rowData!.specObj.specType"
                            @change="changeSpecType()"
                        >
                            <el-option
                                v-for="item in specOptions"
                                :key="item.label"
                                :value="item.value"
                                :label="item.label"
                            />
                        </el-select>
                        <el-select
                            v-if="dialogData.rowData!.specObj.specType === 'perWeek'"
                            class="specClass"
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
                            class="specClass"
                            v-model.number="dialogData.rowData!.specObj.day"
                        >
                            <template #append>
                                <div class="append">{{ $t('cronjob.day') }}</div>
                            </template>
                        </el-input>
                        <el-input
                            v-if="hasHour(dialogData.rowData!.specObj)"
                            class="specClass"
                            v-model.number="dialogData.rowData!.specObj.hour"
                        >
                            <template #append>
                                <div class="append">{{ $t('commons.units.hour') }}</div>
                            </template>
                        </el-input>
                        <el-input
                            v-if="dialogData.rowData!.specObj.specType !== 'perNSecond'"
                            class="specClass"
                            v-model.number="dialogData.rowData!.specObj.minute"
                        >
                            <template #append>
                                <div class="append">{{ $t('commons.units.minute') }}</div>
                            </template>
                        </el-input>
                        <el-input
                            v-if="dialogData.rowData!.specObj.specType === 'perNSecond'"
                            class="specClass"
                            v-model.number="dialogData.rowData!.specObj.second"
                        >
                            <template #append>
                                <div class="append">{{ $t('commons.units.second') }}</div>
                            </template>
                        </el-input>
                    </el-form-item>
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
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import LicenseImport from '@/components/license-import/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Toolbox } from '@/api/interface/toolbox';
import { createClam, updateClam } from '@/api/modules/toolbox';
import { specOptions, transObjToSpec, transSpecToObj, weekOptions } from '../../../cronjob/helper';
import { storeToRefs } from 'pinia';
import { GlobalStore } from '@/store';

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
        callback(new Error(i18n.global.t('toolbox.clam.specErr')));
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
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
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
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
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
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
                return;
            }
            break;
        case 'perDay':
            if (item.hour < 0 || item.hour > 23 || item.minute < 0 || item.minute > 59) {
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
                return;
            }
            break;
        case 'perNHour':
            if (item.hour < 0 || item.hour > 8784 || item.minute < 0 || item.minute > 59) {
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
                return;
            }
            break;
        case 'perHour':
            if (item.minute < 0 || item.minute > 59) {
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
                return;
            }
        case 'perNMinute':
            if (item.minute < 0 || item.minute > 527040) {
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
                return;
            }
            break;
        case 'perNSecond':
            if (item.second < 0 || item.second > 31622400) {
                callback(new Error(i18n.global.t('toolbox.clam.specErr')));
                return;
            }
            break;
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

<style scoped lang="scss">
.specClass {
    width: 20% !important;
    margin-left: 20px;
    .append {
        width: 20px;
    }
}
@media only screen and (max-width: 1000px) {
    .specClass {
        width: 100% !important;
        margin-top: 20px;
        margin-left: 0;
        .append {
            width: 43px;
        }
    }
}
.specTypeClass {
    width: 22% !important;
}
@media only screen and (max-width: 1000px) {
    .specTypeClass {
        width: 100% !important;
    }
}
</style>
